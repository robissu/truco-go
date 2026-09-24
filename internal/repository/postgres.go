package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/robissu/truco-go/internal/model"
)

// Postgres guarda as partidas no PostgreSQL.
type Postgres struct {
	pool *pgxpool.Pool
}

func NovoPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

// Criar insere a partida e as duas equipes numa única transação.
func (r *Postgres) Criar(ctx context.Context, p *model.Partida) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO partidas (pontos_para_vencer, finalizada) VALUES ($1, $2) RETURNING id`,
		p.PontosParaVencer, p.Finalizada,
	).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("inserir partida: %w", err)
	}

	for i, e := range p.Equipes {
		_, err = tx.Exec(ctx,
			`INSERT INTO equipes (partida_id, posicao, nome, jogadores, pontos) VALUES ($1, $2, $3, $4, $5)`,
			p.ID, i, e.Nome, e.Jogadores, e.Pontos,
		)
		if err != nil {
			return fmt.Errorf("inserir equipe %d: %w", i, err)
		}
	}

	return tx.Commit(ctx)
}

// Atualizar grava o placar e o estado de finalizada.
func (r *Postgres) Atualizar(ctx context.Context, p *model.Partida) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx,
		`UPDATE partidas SET finalizada = $1 WHERE id = $2`,
		p.Finalizada, p.ID,
	)
	if err != nil {
		return fmt.Errorf("atualizar partida: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrPartidaNaoEncontrada
	}

	for i, e := range p.Equipes {
		_, err = tx.Exec(ctx,
			`UPDATE equipes SET pontos = $1 WHERE partida_id = $2 AND posicao = $3`,
			e.Pontos, p.ID, i,
		)
		if err != nil {
			return fmt.Errorf("atualizar equipe %d: %w", i, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *Postgres) BuscarPorID(ctx context.Context, id int64) (*model.Partida, error) {
	partidas, err := r.consultar(ctx, "WHERE p.id = $1", id)
	if err != nil {
		return nil, err
	}
	if len(partidas) == 0 {
		return nil, model.ErrPartidaNaoEncontrada
	}
	return &partidas[0], nil
}

func (r *Postgres) Listar(ctx context.Context) ([]model.Partida, error) {
	return r.consultar(ctx, "")
}

const selectPartidas = `
	SELECT p.id, p.pontos_para_vencer, p.finalizada, e.posicao, e.nome, e.jogadores, e.pontos
	FROM partidas p
	JOIN equipes e ON e.partida_id = p.id`

// consultar roda o SELECT com um filtro opcional. O filtro vem sempre do
// próprio código (nunca do usuário); os valores vão como parâmetros $1, $2...
func (r *Postgres) consultar(ctx context.Context, filtro string, args ...any) ([]model.Partida, error) {
	rows, err := r.pool.Query(ctx, selectPartidas+" "+filtro+" ORDER BY p.id, e.posicao", args...)
	if err != nil {
		return nil, fmt.Errorf("consultar partidas: %w", err)
	}
	defer rows.Close()

	partidas := []model.Partida{}
	for rows.Next() {
		var (
			p       model.Partida
			e       model.Equipe
			posicao int
		)
		if err := rows.Scan(&p.ID, &p.PontosParaVencer, &p.Finalizada, &posicao, &e.Nome, &e.Jogadores, &e.Pontos); err != nil {
			return nil, fmt.Errorf("ler linha: %w", err)
		}
		// O JOIN devolve duas linhas por partida (uma por equipe), já ordenadas
		// por id: só abre uma partida nova quando o id muda.
		if len(partidas) == 0 || partidas[len(partidas)-1].ID != p.ID {
			partidas = append(partidas, p)
		}
		partidas[len(partidas)-1].Equipes[posicao] = e
	}
	return partidas, rows.Err()
}
