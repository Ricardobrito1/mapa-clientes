package core

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// OpenDB conecta no Turso usando as variaveis de ambiente
// TURSO_DATABASE_URL (ex: libsql://mapa-clientes-seuuser.turso.io)
// e TURSO_AUTH_TOKEN (token gerado no dashboard/CLI do Turso).
func OpenDB() (*sql.DB, error) {
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")
	if dbURL == "" || token == "" {
		return nil, fmt.Errorf("defina TURSO_DATABASE_URL e TURSO_AUTH_TOKEN")
	}

	connStr := fmt.Sprintf("%s?authToken=%s", dbURL, token)
	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("erro abrindo conexao com turso: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro conectando no turso: %w", err)
	}
	return db, nil
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS clientes (
	codigo     INTEGER PRIMARY KEY,
	nome       TEXT NOT NULL,
	endereco   TEXT NOT NULL,
	cep        TEXT,
	bairro     TEXT,
	lat        REAL,
	lon        REAL,
	aproximado INTEGER NOT NULL DEFAULT 0
);
`

func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}