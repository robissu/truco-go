package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/robissu/truco-go/internal/repository"
	"github.com/robissu/truco-go/internal/service"
)

// novoServidor monta a API completa (handler + service + memória), sem banco.
func novoServidor() *http.ServeMux {
	mux := http.NewServeMux()
	svc := service.NovoPartidaService(repository.NovaMemoria())
	NovoPartidaHandler(svc).Registrar(mux)
	return mux
}

// fazer entrega a requisição direto ao mux, sem rede nem porta, e devolve a
// resposta gravada.
func fazer(mux *http.ServeMux, metodo, url, corpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, url, strings.NewReader(corpo))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

const corpoPartida = `{"pontos_para_vencer": 12, "equipes": [{"nome": "Nós", "jogadores": ["Ana"]}, {"nome": "Eles", "jogadores": ["Bruno"]}]}`

func TestCriarPartida(t *testing.T) {
	casos := []struct {
		nome   string
		corpo  string
		status int
	}{
		{nome: "partida válida", corpo: corpoPartida, status: http.StatusCreated},
		{nome: "JSON quebrado", corpo: `{"pontos_para_vencer": 12,`, status: http.StatusBadRequest},
		{nome: "meta inválida", status: http.StatusBadRequest,
			corpo: `{"pontos_para_vencer": 15, "equipes": [{"nome": "A", "jogadores": ["Ana"]}, {"nome": "B", "jogadores": ["Bruno"]}]}`},
		{nome: "equipes de tamanhos diferentes", status: http.StatusBadRequest,
			corpo: `{"pontos_para_vencer": 12, "equipes": [{"nome": "A", "jogadores": ["Ana", "Caio"]}, {"nome": "B", "jogadores": ["Bruno"]}]}`},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec := fazer(novoServidor(), http.MethodPost, "/partidas", c.corpo)
			if rec.Code != c.status {
				t.Errorf("status = %d; esperado %d (corpo: %s)", rec.Code, c.status, rec.Body)
			}
		})
	}
}

func TestFluxoDaPartida(t *testing.T) {
	mux := novoServidor()

	// Memória nova: a primeira partida sempre recebe o ID 1.
	rec := fazer(mux, http.MethodPost, "/partidas", corpoPartida)
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar: status %d, corpo %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q; esperado charset utf-8 explícito", ct)
	}

	// contra_flor (6) + contra_flor (6) = 12: fecha a partida.
	for i := 1; i <= 2; i++ {
		rec = fazer(mux, http.MethodPost, "/partidas/1/jogadas", `{"equipe": 0, "jogada": "contra_flor"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("jogada %d: status %d, corpo %s", i, rec.Code, rec.Body)
		}
	}

	var partida partidaJSON
	if err := json.NewDecoder(rec.Body).Decode(&partida); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}
	if partida.Equipes[0].Pontos != 12 || !partida.Finalizada {
		t.Errorf("partida = %+v; esperado 12 pontos e finalizada", partida)
	}

	casos := []struct {
		nome   string
		metodo string
		url    string
		corpo  string
		status int
	}{
		{nome: "jogada em partida finalizada", metodo: http.MethodPost, url: "/partidas/1/jogadas",
			corpo: `{"equipe": 1, "jogada": "truco"}`, status: http.StatusConflict},
		{nome: "partida inexistente", metodo: http.MethodGet, url: "/partidas/999", status: http.StatusNotFound},
		{nome: "id não numérico", metodo: http.MethodGet, url: "/partidas/abc", status: http.StatusBadRequest},
		{nome: "listar", metodo: http.MethodGet, url: "/partidas", status: http.StatusOK},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			rec := fazer(mux, c.metodo, c.url, c.corpo)
			if rec.Code != c.status {
				t.Errorf("status = %d; esperado %d (corpo: %s)", rec.Code, c.status, rec.Body)
			}
		})
	}
}
