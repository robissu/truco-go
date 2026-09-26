package aovivo

import (
	"sync"

	"github.com/robissu/truco-go/internal/model"
)

// Transmissor distribui as atualizações de placar para quem está assistindo
// cada partida. Cada espectador recebe as atualizações pelo próprio canal.
type Transmissor struct {
	mu           sync.Mutex
	espectadores map[int64]map[chan model.Partida]struct{}
}

func NovoTransmissor() *Transmissor {
	return &Transmissor{espectadores: make(map[int64]map[chan model.Partida]struct{})}
}

// Inscrever registra um novo espectador da partida e devolve o canal por onde
// ele vai receber as atualizações.
func (t *Transmissor) Inscrever(partidaID int64) chan model.Partida {
	t.mu.Lock()
	defer t.mu.Unlock()

	ch := make(chan model.Partida, 8)
	if t.espectadores[partidaID] == nil {
		t.espectadores[partidaID] = make(map[chan model.Partida]struct{})
	}
	t.espectadores[partidaID][ch] = struct{}{}
	return ch
}

// Cancelar remove o espectador (quando ele desconecta) e fecha o canal dele.
func (t *Transmissor) Cancelar(partidaID int64, ch chan model.Partida) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.espectadores[partidaID], ch)
	if len(t.espectadores[partidaID]) == 0 {
		delete(t.espectadores, partidaID)
	}
	close(ch)
}

// Publicar envia o placar novo para todos os espectadores da partida. Se o
// canal de algum espectador estiver cheio (cliente lento), a atualização é
// descartada só para ele, em vez de travar quem publicou.
func (t *Transmissor) Publicar(p model.Partida) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for ch := range t.espectadores[p.ID] {
		select {
		case ch <- p:
		default:
		}
	}
}
