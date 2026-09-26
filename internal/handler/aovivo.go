package handler

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// aoVivo abre uma conexão WebSocket e envia o placar da partida a cada jogada
// registrada, até o cliente desconectar.
func (h *PartidaHandler) aoVivo(w http.ResponseWriter, r *http.Request) {
	id, ok := lerID(w, r)
	if !ok {
		return
	}

	// Inscreve antes de ler o placar: assim nenhuma jogada registrada entre a
	// leitura e a inscrição se perde.
	atualizacoes := h.transmissor.Inscrever(id)
	defer h.transmissor.Cancelar(id, atualizacoes)

	// Partida inexistente recebe um 404 HTTP comum, antes de abrir a conexão.
	p, err := h.svc.BuscarPartida(r.Context(), id)
	if err != nil {
		responderErroDoService(w, err)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return // o Accept já respondeu o erro ao cliente
	}
	defer conn.CloseNow()

	// O cliente só assiste, não manda nada. CloseRead descarta o que chegar e
	// devolve um contexto que é cancelado quando o cliente desconecta.
	ctx := conn.CloseRead(r.Context())

	if err := wsjson.Write(ctx, conn, paraJSON(p)); err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case partida := <-atualizacoes:
			if err := wsjson.Write(ctx, conn, paraJSON(&partida)); err != nil {
				return
			}
		}
	}
}
