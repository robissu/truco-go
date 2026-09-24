package model

import (
	"errors"
	"fmt"
)

// ErrPartidaNaoEncontrada é devolvido quando não existe partida com o ID pedido.
var ErrPartidaNaoEncontrada = errors.New("partida não encontrada")

// TipoJogada representa uma jogada que pode acontecer numa mão de truco.
// É baseado em string, mas por ser um tipo próprio (não string pura),
// o compilador impede passar qualquer string solta onde se espera um TipoJogada.
type TipoJogada string

const (
	Mao        TipoJogada = "mao"
	Truco      TipoJogada = "truco"
	Retruco    TipoJogada = "retruco"
	ValeQuatro TipoJogada = "vale_quatro"
	Envido     TipoJogada = "envido"
	RealEnvido TipoJogada = "real_envido"
	Flor       TipoJogada = "flor"
	ContraFlor TipoJogada = "contra_flor"
)

// Equipe representa uma das duas equipes da partida.
type Equipe struct {
	Nome      string
	Jogadores []string
	Pontos    int
}

// Partida representa uma partida de truco gaudério em andamento.
type Partida struct {
	ID               int64
	PontosParaVencer int // 12 ou 24
	Equipes          [2]Equipe
	Finalizada       bool
}

// PontosDaJogada devolve quantos pontos uma jogada vale quando é ganha.
// Jogadas cujo valor depende do estado da partida (falta envido,
// contra-flor-e-o-resto) não entram aqui de propósito — ficam pra quando
// a partida tiver acesso ao placar atual.
func PontosDaJogada(jogada TipoJogada) (int, error) {
	switch jogada {
	case Mao:
		return 1, nil
	case Truco:
		return 2, nil
	case Retruco:
		return 3, nil
	case ValeQuatro:
		return 4, nil
	case Envido:
		return 2, nil
	case RealEnvido:
		return 5, nil
	case Flor:
		return 3, nil
	case ContraFlor:
		return 6, nil
	default:
		return 0, fmt.Errorf("jogada desconhecida: %q", jogada)
	}
}

// NovaPartida cria uma partida validando as regras de montagem:
// meta de 12 ou 24 pontos e equipes do mesmo tamanho (1x1 ou 2x2).
func NovaPartida(pontosParaVencer int, a, b Equipe) (*Partida, error) {
	if pontosParaVencer != 12 && pontosParaVencer != 24 {
		return nil, fmt.Errorf("pontos para vencer deve ser 12 ou 24, recebido: %d", pontosParaVencer)
	}
	if len(a.Jogadores) < 1 || len(a.Jogadores) > 2 {
		return nil, fmt.Errorf("equipe %q deve ter 1 ou 2 jogadores, tem %d", a.Nome, len(a.Jogadores))
	}
	if len(a.Jogadores) != len(b.Jogadores) {
		return nil, fmt.Errorf("equipes com tamanhos diferentes: %d x %d", len(a.Jogadores), len(b.Jogadores))
	}

	a.Pontos, b.Pontos = 0, 0
	return &Partida{
		PontosParaVencer: pontosParaVencer,
		Equipes:          [2]Equipe{a, b},
	}, nil
}

// EhDeDuplas informa se a partida é 2x2 (derivado dos jogadores, não armazenado).
func (p *Partida) EhDeDuplas() bool {
	return len(p.Equipes[0].Jogadores) == 2
}

// RegistrarPontos soma pontos à Equipe no índice informado (0 ou 1) e marca
// a partida como finalizada se ela atingir a pontuação necessária.
func (p *Partida) RegistrarPontos(equipeIndex int, pontos int) error {
	if equipeIndex != 0 && equipeIndex != 1 {
		return fmt.Errorf("índice de equipe inválido: %d (esperado 0 ou 1)", equipeIndex)
	}
	if pontos <= 0 {
		return fmt.Errorf("pontos deve ser positivo, recebido: %d", pontos)
	}
	if p.Finalizada {
		return fmt.Errorf("partida %d já está finalizada", p.ID)
	}

	p.Equipes[equipeIndex].Pontos += pontos

	if p.Equipes[equipeIndex].Pontos >= p.PontosParaVencer {
		p.Finalizada = true
	}

	return nil
}
