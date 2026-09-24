package service

import (
	"context"
	"fmt"

	"github.com/robissu/truco-go/internal/model"
)

// PartidaRepository é tudo que o service precisa de um armazenamento.
// Qualquer tipo com esses quatro métodos serve: memória nos testes,
// Postgres em produção.
type PartidaRepository interface {
	Criar(ctx context.Context, p *model.Partida) error
	Atualizar(ctx context.Context, p *model.Partida) error
	BuscarPorID(ctx context.Context, id int64) (*model.Partida, error)
	Listar(ctx context.Context) ([]model.Partida, error)
}

// PartidaService concentra os casos de uso: criar partida, registrar
// jogada e consultar placar.
type PartidaService struct {
	repo PartidaRepository
}

func NovoPartidaService(repo PartidaRepository) *PartidaService {
	return &PartidaService{repo: repo}
}

func (s *PartidaService) CriarPartida(ctx context.Context, pontosParaVencer int, a, b model.Equipe) (*model.Partida, error) {
	p, err := model.NovaPartida(pontosParaVencer, a, b)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Criar(ctx, p); err != nil {
		return nil, fmt.Errorf("salvar partida: %w", err)
	}
	return p, nil
}

func (s *PartidaService) RegistrarJogada(ctx context.Context, id int64, equipeIndex int, jogada model.TipoJogada) (*model.Partida, error) {
	p, err := s.repo.BuscarPorID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("buscar partida %d: %w", id, err)
	}
	pontos, err := model.PontosDaJogada(jogada)
	if err != nil {
		return nil, err
	}
	if err := p.RegistrarPontos(equipeIndex, pontos); err != nil {
		return nil, err
	}
	if err := s.repo.Atualizar(ctx, p); err != nil {
		return nil, fmt.Errorf("atualizar partida %d: %w", id, err)
	}
	return p, nil
}

func (s *PartidaService) BuscarPartida(ctx context.Context, id int64) (*model.Partida, error) {
	return s.repo.BuscarPorID(ctx, id)
}

func (s *PartidaService) ListarPartidas(ctx context.Context) ([]model.Partida, error) {
	return s.repo.Listar(ctx)
}
