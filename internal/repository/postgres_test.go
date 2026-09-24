package repository

import (
	"context"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/robissu/truco-go/internal/model"
)

// novoPostgresDeTeste conecta no banco apontado por DATABASE_URL. Sem a
// variável, o teste é pulado: assim `go test ./...` roda em qualquer máquina.
func novoPostgresDeTeste(t *testing.T) *Postgres {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL não definida; pulando teste de integração")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("conectar no banco: %v", err)
	}
	t.Cleanup(pool.Close)

	return NovoPostgres(pool)
}

func TestPostgresCicloCompleto(t *testing.T) {
	repo := novoPostgresDeTeste(t)
	ctx := context.Background()

	p, err := model.NovaPartida(12,
		model.Equipe{Nome: "Nós", Jogadores: []string{"Ana", "Caio"}},
		model.Equipe{Nome: "Eles", Jogadores: []string{"Bruno", "Duda"}},
	)
	if err != nil {
		t.Fatalf("nova partida: %v", err)
	}

	if err := repo.Criar(ctx, p); err != nil {
		t.Fatalf("criar: %v", err)
	}
	if p.ID == 0 {
		t.Fatal("Criar deveria preencher o ID gerado pelo banco")
	}

	if err := p.RegistrarPontos(1, 12); err != nil {
		t.Fatalf("registrar pontos: %v", err)
	}
	if err := repo.Atualizar(ctx, p); err != nil {
		t.Fatalf("atualizar: %v", err)
	}

	salva, err := repo.BuscarPorID(ctx, p.ID)
	if err != nil {
		t.Fatalf("buscar: %v", err)
	}
	if salva.Equipes[1].Pontos != 12 || !salva.Finalizada {
		t.Errorf("placar salvo errado: %+v", salva)
	}
	if !slices.Equal(salva.Equipes[0].Jogadores, []string{"Ana", "Caio"}) {
		t.Errorf("jogadores salvos errado: %v", salva.Equipes[0].Jogadores)
	}
}

func TestPostgresPartidaInexistente(t *testing.T) {
	repo := novoPostgresDeTeste(t)
	ctx := context.Background()

	if _, err := repo.BuscarPorID(ctx, -1); !errors.Is(err, model.ErrPartidaNaoEncontrada) {
		t.Errorf("BuscarPorID: esperava ErrPartidaNaoEncontrada, recebeu: %v", err)
	}
	if err := repo.Atualizar(ctx, &model.Partida{ID: -1}); !errors.Is(err, model.ErrPartidaNaoEncontrada) {
		t.Errorf("Atualizar: esperava ErrPartidaNaoEncontrada, recebeu: %v", err)
	}
}
