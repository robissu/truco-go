# truco-go

API REST de placar para partidas de **Truco Gaudério**, escrita em Go.
Registra partidas (1x1 ou 2x2, até 12 ou 24 pontos), as jogadas de cada mão
(truco, retruco, envido, flor...) e mantém o placar persistido em PostgreSQL.

É um port da lógica de domínio do meu projeto Truco Gaudério em Java
para uma API em Go, feito para aprender a linguagem na prática.

## Stack

- Go 1.27, apenas `net/http` da biblioteca padrão (sem framework web)
- PostgreSQL 17 com o driver [pgx](https://github.com/jackc/pgx)
- Docker e Docker Compose

## Como rodar

Com Docker (API + banco):

```bash
docker compose up -d --build
```

A API fica em `http://localhost:8080`. As tabelas são criadas automaticamente
pelo `migrations/001_init.sql` na primeira subida do banco.

Sem Docker, com o repositório em memória (os dados somem ao encerrar):

```bash
go run ./cmd/api
```

Para usar um Postgres existente, defina `DATABASE_URL`
(ex.: `postgres://truco:truco@localhost:5432/truco`).

## Endpoints

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/health` | verificação de saúde |
| `POST` | `/partidas` | cria uma partida |
| `GET` | `/partidas` | lista as partidas |
| `GET` | `/partidas/{id}` | consulta o placar de uma partida |
| `POST` | `/partidas/{id}/jogadas` | registra uma jogada ganha por uma equipe |

Criar uma partida:

```json
POST /partidas
{
  "pontos_para_vencer": 12,
  "equipes": [
    { "nome": "Nós",  "jogadores": ["Ana"] },
    { "nome": "Eles", "jogadores": ["Bruno"] }
  ]
}
```

Registrar uma jogada (`equipe` é 0 ou 1):

```json
POST /partidas/1/jogadas
{ "equipe": 0, "jogada": "truco" }
```

Jogadas aceitas e seus valores:

| Jogada | Pontos |
|---|---|
| `mao` | 1 |
| `truco` / `retruco` / `vale_quatro` | 2 / 3 / 4 |
| `envido` / `real_envido` | 2 / 5 |
| `flor` / `contra_flor` | 3 / 6 |

Erros voltam como `{"erro": "mensagem"}`, com o status HTTP correspondente:

| Status | Quando |
|---|---|
| 400 | JSON inválido, meta diferente de 12/24, jogada desconhecida, equipes de tamanhos diferentes |
| 404 | partida inexistente |
| 409 | jogada em partida já finalizada |
| 500 | falha interna (o detalhe vai só para o log) |

## Estrutura

```
cmd/api/              ponto de entrada: escolhe o repositório e liga as camadas
internal/
  handler/            HTTP: lê a requisição, chama o service, escreve a resposta
  service/            casos de uso e a interface PartidaRepository
  model/              structs e regras do truco, sem dependências externas
  repository/         implementações da interface: memória e PostgreSQL
migrations/           SQL de criação das tabelas
```

## Decisões de arquitetura

### Camadas

- **handler**: traduz o protocolo HTTP para os tipos Go e vice-versa.
- **service**: organiza as funções na ordem certa de cada requisição.
- **model**: contém as regras do jogo.
- **repository**: armazena os dados de cada partida, incluindo jogadores, e implementa a interface.

Caso queira mudar o valor de alguma jogada do jogo, é necessário apenas alterar o valor no model.

### O service depende de uma interface

O service usa uma interface com 4 funções que exigem "criar", "atualizar", "buscar" e "listar". A ideia de usar uma interface foi considerada para permitir a migração do repositório apenas trocando a `DATABASE_URL`, além de poder testar sem banco usando memória, sem mudar o service. O Postgres e a memória implementam essa interface. Diferente do Java, quem implementa não utiliza o termo `implements`.

### Erros

Existem 4 tipos de erros:

- erro de validação: 400
- partida não encontrada: 404
- partida encerrada: 409
- outros: 500

### Transações

No caso de os dados estarem certos e acontecer algo errado no meio dos `INSERT`s, existe um rollback que retira todos os dados que já foram inseridos no banco e volta ao estado anterior antes da transação. Sem a transação, a partida ficaria salva sem equipes, por exemplo.

### Sem framework web

Na versão 1.22 do Go, o mux passou a entender o método (`GET`, `POST`) e as partes variáveis da URL (como o `{id}`), sendo que antes era necessário inserir manualmente a URL. Essas seriam as principais razões para se usar o `chi` ou o `gorilla/mux`, que seriam opções de frameworks.

## Testes

```bash
go test ./...
```

| Pacote | Tipo | O que cobre |
|---|---|---|
| `model` | unidade | regras de pontuação e validações |
| `service` | unidade (repositório em memória) | fluxo dos casos de uso |
| `handler` | `httptest` (repositório em memória) | rotas, JSON e códigos de status |
| `repository` | integração (PostgreSQL real) | SQL, transações, leitura com JOIN |

Os testes de integração são pulados se `DATABASE_URL` não estiver definida.
Para rodá-los contra o banco do Compose:

```bash
docker run --rm --network truco-go_default \
  -e DATABASE_URL=postgres://truco:truco@db:5432/truco \
  -v "$PWD:/app" -w /app golang:1.27 go test ./...
```

## Limitações e próximos passos

- **Falta envido** e **contra-flor e o resto** ainda não são suportados, porque o valor deles depende do placar atual.
- Não há histórico das jogadas, apenas o placar acumulado.
- Sem *graceful shutdown*: ao encerrar, as requisições em andamento não são aguardadas.
- As migrations rodam só na primeira subida do banco. Uma ferramenta como `golang-migrate` resolveria a evolução do schema.