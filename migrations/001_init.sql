CREATE TABLE partidas (
    id                 BIGSERIAL PRIMARY KEY,
    pontos_para_vencer INT         NOT NULL CHECK (pontos_para_vencer IN (12, 24)),
    finalizada         BOOLEAN     NOT NULL DEFAULT FALSE,
    criada_em          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE equipes (
    partida_id BIGINT   NOT NULL REFERENCES partidas (id) ON DELETE CASCADE,
    posicao    SMALLINT NOT NULL CHECK (posicao IN (0, 1)),
    nome       TEXT     NOT NULL,
    jogadores  TEXT[]   NOT NULL,
    pontos     INT      NOT NULL DEFAULT 0 CHECK (pontos >= 0),
    PRIMARY KEY (partida_id, posicao)
);