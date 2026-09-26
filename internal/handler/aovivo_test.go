package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// urlWS troca o "http://" do servidor de teste por "ws://".
func urlWS(srv *httptest.Server, caminho string) string {
	return "ws" + strings.TrimPrefix(srv.URL, "http") + caminho
}

func TestAoVivo(t *testing.T) {
	mux := novoServidor()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	if rec := fazer(mux, http.MethodPost, "/partidas", corpoPartida); rec.Code != http.StatusCreated {
		t.Fatalf("criar: status %d, corpo %s", rec.Code, rec.Body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, urlWS(srv, "/partidas/1/ao-vivo"), nil)
	if err != nil {
		t.Fatalf("conectar: %v", err)
	}
	defer conn.CloseNow()

	var inicial partidaJSON
	if err := wsjson.Read(ctx, conn, &inicial); err != nil {
		t.Fatalf("ler placar inicial: %v", err)
	}
	if inicial.Equipes[0].Pontos != 0 {
		t.Errorf("placar inicial = %d; esperado 0", inicial.Equipes[0].Pontos)
	}

	if rec := fazer(mux, http.MethodPost, "/partidas/1/jogadas", `{"equipe": 0, "jogada": "truco"}`); rec.Code != http.StatusOK {
		t.Fatalf("jogada: status %d, corpo %s", rec.Code, rec.Body)
	}

	var atualizada partidaJSON
	if err := wsjson.Read(ctx, conn, &atualizada); err != nil {
		t.Fatalf("ler atualização: %v", err)
	}
	if atualizada.Equipes[0].Pontos != 2 {
		t.Errorf("placar atualizado = %d; esperado 2", atualizada.Equipes[0].Pontos)
	}
}

func TestAoVivoPartidaInexistente(t *testing.T) {
	srv := httptest.NewServer(novoServidor())
	defer srv.Close()

	_, resp, err := websocket.Dial(context.Background(), urlWS(srv, "/partidas/999/ao-vivo"), nil)
	if err == nil {
		t.Fatal("esperava falha ao conectar numa partida inexistente")
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Errorf("esperava resposta 404, recebeu: %v", resp)
	}
}
