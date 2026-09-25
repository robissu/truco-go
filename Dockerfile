# ---- Estágio 1: compila ----
FROM golang:1.27 AS build
WORKDIR /src

# Dependências primeiro: essa camada só é refeita quando go.mod/go.sum mudam.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

# ---- Estágio 2: roda ----
FROM gcr.io/distroless/static-debian12
COPY --from=build /api /api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]