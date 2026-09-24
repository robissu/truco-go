package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/robissu/truco-go/internal/model"
	"github.com/robissu/truco-go/internal/service"
)

// PartidaHandler traduz HTTP <-> service: lê a requisição, chama o caso de
// uso e escreve a resposta. Não tem regra de negócio.
type PartidaHandler struct {
	svc *service.PartidaService
}

func NovoPartidaHandler(svc *service.PartidaService) *PartidaHandler {
	return &PartidaHandler{svc: svc}
}

// Registrar liga as rotas da API no mux.
func (h *PartidaHandler) Registrar(mux *http.ServeMux) {
	mux.HandleFunc("POST /partidas", h.criar)
	mux.HandleFunc("GET /partidas", h.listar)
	mux.HandleFunc("GET /partidas/{id}", h.buscar)
	mux.HandleFunc("POST /partidas/{id}/jogadas", h.registrarJogada)
}

// Formatos JSON da API. Ficam separados das structs do model pra que o
// contrato HTTP e o domínio possam mudar de forma independente.

type novaEquipeJSON struct {
	Nome      string   `json:"nome"`
	Jogadores []string `json:"jogadores"`
}

type criarPartidaRequest struct {
	PontosParaVencer int               `json:"pontos_para_vencer"`
	Equipes          [2]novaEquipeJSON `json:"equipes"`
}

type registrarJogadaRequest struct {
	Equipe int              `json:"equipe"`
	Jogada model.TipoJogada `json:"jogada"`
}

type equipeJSON struct {
	Nome      string   `json:"nome"`
	Jogadores []string `json:"jogadores"`
	Pontos    int      `json:"pontos"`
}

type partidaJSON struct {
	ID               int64         `json:"id"`
	PontosParaVencer int           `json:"pontos_para_vencer"`
	Equipes          [2]equipeJSON `json:"equipes"`
	Finalizada       bool          `json:"finalizada"`
}

func paraJSON(p *model.Partida) partidaJSON {
	resp := partidaJSON{ID: p.ID, PontosParaVencer: p.PontosParaVencer, Finalizada: p.Finalizada}
	for i, e := range p.Equipes {
		resp.Equipes[i] = equipeJSON{Nome: e.Nome, Jogadores: e.Jogadores, Pontos: e.Pontos}
	}
	return resp
}

func (h *PartidaHandler) criar(w http.ResponseWriter, r *http.Request) {
	var req criarPartidaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderErro(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	a := model.Equipe{Nome: req.Equipes[0].Nome, Jogadores: req.Equipes[0].Jogadores}
	b := model.Equipe{Nome: req.Equipes[1].Nome, Jogadores: req.Equipes[1].Jogadores}

	p, err := h.svc.CriarPartida(r.Context(), req.PontosParaVencer, a, b)
	if err != nil {
		responderErroDoService(w, err)
		return
	}
	responderJSON(w, http.StatusCreated, paraJSON(p))
}

func (h *PartidaHandler) listar(w http.ResponseWriter, r *http.Request) {
	partidas, err := h.svc.ListarPartidas(r.Context())
	if err != nil {
		responderErroDoService(w, err)
		return
	}

	resp := make([]partidaJSON, 0, len(partidas))
	for i := range partidas {
		resp = append(resp, paraJSON(&partidas[i]))
	}
	responderJSON(w, http.StatusOK, resp)
}

func (h *PartidaHandler) buscar(w http.ResponseWriter, r *http.Request) {
	id, ok := lerID(w, r)
	if !ok {
		return
	}

	p, err := h.svc.BuscarPartida(r.Context(), id)
	if err != nil {
		responderErroDoService(w, err)
		return
	}
	responderJSON(w, http.StatusOK, paraJSON(p))
}

func (h *PartidaHandler) registrarJogada(w http.ResponseWriter, r *http.Request) {
	id, ok := lerID(w, r)
	if !ok {
		return
	}

	var req registrarJogadaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderErro(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	p, err := h.svc.RegistrarJogada(r.Context(), id, req.Equipe, req.Jogada)
	if err != nil {
		responderErroDoService(w, err)
		return
	}
	responderJSON(w, http.StatusOK, paraJSON(p))
}

// lerID extrai o {id} da URL. Se for inválido, já responde 400 e devolve ok=false.
func lerID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		responderErro(w, http.StatusBadRequest, "id deve ser um número inteiro")
		return 0, false
	}
	return id, true
}

// responderErroDoService escolhe o status HTTP pelo tipo do erro.
func responderErroDoService(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrValidacao):
		responderErro(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, model.ErrPartidaNaoEncontrada):
		responderErro(w, http.StatusNotFound, err.Error())
	case errors.Is(err, model.ErrPartidaFinalizada):
		responderErro(w, http.StatusConflict, err.Error())
	default:
		// Erro nosso (banco fora, bug): detalhe só no log, nunca pro cliente.
		log.Printf("erro interno: %v", err)
		responderErro(w, http.StatusInternalServerError, "erro interno")
	}
}

func responderJSON(w http.ResponseWriter, status int, corpo any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(corpo); err != nil {
		log.Printf("escrever resposta: %v", err)
	}
}

func responderErro(w http.ResponseWriter, status int, msg string) {
	responderJSON(w, status, map[string]string{"erro": msg})
}
