package repository

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"github.com/robissu/truco-go/internal/model"
)

// Memoria guarda as partidas num map. Serve pros testes e pra rodar a API
// sem banco; os dados somem quando o programa encerra.
type Memoria struct {
	mu        sync.Mutex
	partidas  map[int64]model.Partida
	proximoID int64
}

func NovaMemoria() *Memoria {
	return &Memoria{partidas: make(map[int64]model.Partida), proximoID: 1}
}

func (m *Memoria) Criar(ctx context.Context, p *model.Partida) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p.ID = m.proximoID
	m.proximoID++
	m.partidas[p.ID] = *p
	return nil
}

func (m *Memoria) Atualizar(ctx context.Context, p *model.Partida) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.partidas[p.ID]; !ok {
		return model.ErrPartidaNaoEncontrada
	}
	m.partidas[p.ID] = *p
	return nil
}

func (m *Memoria) BuscarPorID(ctx context.Context, id int64) (*model.Partida, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.partidas[id]
	if !ok {
		return nil, model.ErrPartidaNaoEncontrada
	}
	return &p, nil
}

func (m *Memoria) Listar(ctx context.Context) ([]model.Partida, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lista := make([]model.Partida, 0, len(m.partidas))
	for _, p := range m.partidas {
		lista = append(lista, p)
	}
	slices.SortFunc(lista, func(a, b model.Partida) int { return cmp.Compare(a.ID, b.ID) })
	return lista, nil
}
