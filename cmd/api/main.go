package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/robissu/truco-go/internal/handler"
	"github.com/robissu/truco-go/internal/repository"
	"github.com/robissu/truco-go/internal/service"
)

func main() {
	repo := escolherRepositorio()
	svc := service.NovoPartidaService(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	handler.NovoPartidaHandler(svc).Registrar(mux)

	addr := ":8080"
	log.Printf("servidor ouvindo em %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// escolherRepositorio usa o Postgres se DATABASE_URL estiver definida e a
// memória caso contrário. O resto do programa não sabe qual dos dois recebeu.
func escolherRepositorio() service.PartidaRepository {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Println("DATABASE_URL não definida: usando repositório em memória")
		return repository.NovaMemoria()
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatalf("configurar conexão com o banco: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("conectar no banco: %v", err)
	}
	log.Println("usando repositório Postgres")
	return repository.NovoPostgres(pool)
}
