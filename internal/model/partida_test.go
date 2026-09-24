package model

import "testing"

func TestPontosDaJogada(t *testing.T) {
	casos := []struct {
		nome     string
		jogada   TipoJogada
		esperado int
		comErro  bool
	}{
		{nome: "mao vale 1", jogada: Mao, esperado: 1},
		{nome: "truco vale 2", jogada: Truco, esperado: 2},
		{nome: "retruco vale 3", jogada: Retruco, esperado: 3},
		{nome: "vale quatro vale 4", jogada: ValeQuatro, esperado: 4},
		{nome: "envido vale 2", jogada: Envido, esperado: 2},
		{nome: "real envido vale 5", jogada: RealEnvido, esperado: 5},
		{nome: "flor vale 3", jogada: Flor, esperado: 3},
		{nome: "contra flor vale 6", jogada: ContraFlor, esperado: 6},
		{nome: "jogada inválida retorna erro", jogada: TipoJogada("chute"), comErro: true},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			pontos, err := PontosDaJogada(c.jogada)

			if c.comErro {
				if err == nil {
					t.Fatalf("esperava erro para jogada %q, não recebeu nenhum", c.jogada)
				}
				return
			}

			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if pontos != c.esperado {
				t.Errorf("PontosDaJogada(%q) = %d; esperado %d", c.jogada, pontos, c.esperado)
			}
		})
	}
}

func TestRegistrarPontos(t *testing.T) {
	t.Run("soma pontos normalmente", func(t *testing.T) {
		p := &Partida{PontosParaVencer: 12}
		err := p.RegistrarPontos(0, 3)

		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if p.Equipes[0].Pontos != 3 {
			t.Errorf("pontos da equipe 0 = %d; esperado 3", p.Equipes[0].Pontos)
		}
		if p.Finalizada {
			t.Error("partida não deveria estar finalizada ainda")
		}
	})

	t.Run("finaliza partida ao bater a meta", func(t *testing.T) {
		p := &Partida{PontosParaVencer: 12}
		_ = p.RegistrarPontos(1, 10)
		err := p.RegistrarPontos(1, 2)

		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if !p.Finalizada {
			t.Error("partida deveria estar finalizada")
		}
	})

	t.Run("rejeita índice de dupla inválido", func(t *testing.T) {
		p := &Partida{PontosParaVencer: 12}
		err := p.RegistrarPontos(5, 3)

		if err == nil {
			t.Fatal("esperava erro para índice inválido")
		}
	})

	t.Run("rejeita registro em partida já finalizada", func(t *testing.T) {
		p := &Partida{PontosParaVencer: 12, Finalizada: true}
		err := p.RegistrarPontos(0, 1)

		if err == nil {
			t.Fatal("esperava erro ao registrar pontos em partida finalizada")
		}
	})
}

func TestNovaPartida(t *testing.T) {
	casos := []struct {
		nome    string
		meta    int
		a, b    Equipe
		comErro bool
	}{
		{nome: "1x1 até 12", meta: 12,
			a: Equipe{Nome: "A", Jogadores: []string{"Ana"}},
			b: Equipe{Nome: "B", Jogadores: []string{"Bruno"}}},
		{nome: "2x2 até 24", meta: 24,
			a: Equipe{Nome: "A", Jogadores: []string{"Ana", "Caio"}},
			b: Equipe{Nome: "B", Jogadores: []string{"Bruno", "Duda"}}},
		{nome: "meta inválida", meta: 15, comErro: true,
			a: Equipe{Nome: "A", Jogadores: []string{"Ana"}},
			b: Equipe{Nome: "B", Jogadores: []string{"Bruno"}}},
		{nome: "equipe sem jogadores", meta: 12, comErro: true,
			a: Equipe{Nome: "A"},
			b: Equipe{Nome: "B"}},
		{nome: "tamanhos diferentes", meta: 12, comErro: true,
			a: Equipe{Nome: "A", Jogadores: []string{"Ana", "Caio"}},
			b: Equipe{Nome: "B", Jogadores: []string{"Bruno"}}},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			p, err := NovaPartida(c.meta, c.a, c.b)

			if c.comErro {
				if err == nil {
					t.Fatal("esperava erro, não recebeu nenhum")
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if p.EhDeDuplas() != (len(c.a.Jogadores) == 2) {
				t.Errorf("EhDeDuplas() = %v; esperado %v", p.EhDeDuplas(), len(c.a.Jogadores) == 2)
			}
		})
	}
}
