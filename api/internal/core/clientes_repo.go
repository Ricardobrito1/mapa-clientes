package core

import (
	"database/sql"
)

func ListaClientes(db *sql.DB) ([]Cliente, error) {
	rows, err := db.Query(`SELECT codigo, nome, endereco, cep, bairro, lat, lon, aproximado FROM clientes ORDER BY nome`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClientes(rows)
}

func BuscaClientes(db *sql.DB, termo string) ([]Cliente, error) {
	like := "%" + termo + "%"
	rows, err := db.Query(
		`SELECT codigo, nome, endereco, cep, bairro, lat, lon, aproximado
		 FROM clientes
		 WHERE nome LIKE ? OR endereco LIKE ? OR bairro LIKE ? OR cep LIKE ?
		 ORDER BY nome`,
		like, like, like, like,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClientes(rows)
}

func scanClientes(rows *sql.Rows) ([]Cliente, error) {
	clientes := []Cliente{}
	for rows.Next() {
		var c Cliente
		var cep, bairro sql.NullString
		if err := rows.Scan(&c.Codigo, &c.Nome, &c.Endereco, &cep, &bairro, &c.Lat, &c.Lon, &c.Aproximado); err != nil {
			return nil, err
		}
		c.Cep = cep.String
		c.Bairro = bairro.String
		clientes = append(clientes, c)
	}
	return clientes, rows.Err()
}

func ProximoCodigo(db *sql.DB) (int, error) {
	var max sql.NullInt64
	err := db.QueryRow(`SELECT MAX(codigo) FROM clientes`).Scan(&max)
	if err != nil {
		return 0, err
	}
	if !max.Valid {
		return 1, nil
	}
	return int(max.Int64) + 1, nil
}

func InsereCliente(db *sql.DB, c Cliente) error {
	_, err := db.Exec(
		`INSERT INTO clientes (codigo, nome, endereco, cep, bairro, lat, lon, aproximado)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Codigo, c.Nome, c.Endereco, c.Cep, c.Bairro, c.Lat, c.Lon, c.Aproximado,
	)
	return err
}

// UpsertCliente insere ou substitui um cliente pelo codigo (PK).
// Usado pela importacao em lote, que pode ser rodada mais de uma vez sem erro.
func UpsertCliente(db *sql.DB, c Cliente) error {
	_, err := db.Exec(
		`INSERT OR REPLACE INTO clientes (codigo, nome, endereco, cep, bairro, lat, lon, aproximado)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Codigo, c.Nome, c.Endereco, c.Cep, c.Bairro, c.Lat, c.Lon, c.Aproximado,
	)
	return err
}
