package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"mapa-clientes-api/internal/core"
)

func main() {
	_ = godotenv.Load()

	caminho := flag.String("file", "../clientes.js", "caminho pro clientes.js a importar")
	flag.Parse()

	conteudo, err := os.ReadFile(*caminho)
	if err != nil {
		log.Fatalf("erro lendo %s: %v", *caminho, err)
	}

	jsonStr, err := extraiArrayJSON(string(conteudo))
	if err != nil {
		log.Fatalf("erro extraindo array do clientes.js: %v", err)
	}

	var clientes []core.Cliente
	if err := json.Unmarshal([]byte(jsonStr), &clientes); err != nil {
		log.Fatalf("erro parseando json: %v", err)
	}
	log.Printf("%d clientes encontrados em %s", len(clientes), *caminho)

	db, err := core.OpenDB()
	if err != nil {
		log.Fatalf("erro conectando no turso: %v", err)
	}
	defer db.Close()

	if err := core.EnsureSchema(db); err != nil {
		log.Fatalf("erro criando schema: %v", err)
	}

	importados := 0
	for _, c := range clientes {
		if err := core.UpsertCliente(db, c); err != nil {
			log.Printf("erro importando cliente %d (%s): %v", c.Codigo, c.Nome, err)
			continue
		}
		importados++
	}
	log.Printf("importacao concluida: %d/%d clientes gravados", importados, len(clientes))
}

// extraiArrayJSON pega o texto entre "= " e ";" de "const CLIENTES = [...];"
func extraiArrayJSON(conteudo string) (string, error) {
	_, depois, encontrou := strings.Cut(conteudo, "= ")
	if !encontrou {
		return "", os.ErrInvalid
	}
	antes, _, encontrou := strings.Cut(strings.TrimRight(depois, "\n\r\t "), ";")
	if !encontrou {
		return "", os.ErrInvalid
	}
	return antes, nil
}
