package service

import (
	"context"
	"errors"
	"testing"

	"github.com/robissu/truco-go/internal/model"
	"github.com/robissu/truco-go/internal/repository"
)

func TestRegistrarJogada(t *testing.T) {
	ctx := context.Background()
	a := model.Equipe{Nome: "Nós", Jogadores: []string{"Ana"}}
	b := model.Equipe{Nome: "Eles", Jogadores: []string{"Bruno"}}

	casos := []struct {
		nome            string
		jogadas         []model.TipoJogada // todas ganhas pela equipe 0
		pontosEsperados int
		finalizada      bool
	}{
		{nome: "um truco", jogadas: []model.TipoJogada{model.Truco}, pontosEsperados: 2},
		{nome: "envido e retruco", jogadas: []model.TipoJogada{model.Envido, model.Retruco}, pontosEsperados: 5},
		{nome: "chega a 12 e finaliza", jogadas: []model.TipoJogada{model.ContraFlor, model.RealEnvido, model.Mao}, pontosEsperados: 12, finalizada: true},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			svc := NovoPartidaService(repository.NovaMemoria())
			p, err := svc.CriarPartida(ctx, 12, a, b)
			if err != nil {
				t.Fatalf("criar partida: %v", err)
			}

			for _, j := range c.jogadas {
				if _, err := svc.RegistrarJogada(ctx, p.ID, 0, j); err != nil {
					t.Fatalf("registrar %q: %v", j, err)
				}
			}

			salva, err := svc.BuscarPartida(ctx, p.ID)
			if err != nil {
				t.Fatalf("buscar partida: %v", err)
			}
			if salva.Equipes[0].Pontos != c.pontosEsperados {
				t.Errorf("pontos = %d; esperado %d", salva.Equipes[0].Pontos, c.pontosEsperados)
			}
			if salva.Finalizada != c.finalizada {
				t.Errorf("finalizada = %v; esperado %v", salva.Finalizada, c.finalizada)
			}
		})
	}
}

func TestRegistrarJogadaPartidaInexistente(t *testing.T) {
	svc := NovoPartidaService(repository.NovaMemoria())

	_, err := svc.RegistrarJogada(context.Background(), 99, 0, model.Truco)

	if !errors.Is(err, model.ErrPartidaNaoEncontrada) {
		t.Fatalf("esperava ErrPartidaNaoEncontrada, recebeu: %v", err)
	}
}
