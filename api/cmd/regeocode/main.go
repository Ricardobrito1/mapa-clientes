package main

import (
	"log"
	"math"
	"time"

	"github.com/joho/godotenv"
	"mapa-clientes-api/internal/core"
)

func main() {
	_ = godotenv.Load()
	db, err := core.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	clientes, err := core.ListaClientes(db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("revisando %d clientes...", len(clientes))

	var atualizados, marcadosAproximado, semMudanca, semLocalizacao, erros int

	for i, c := range clientes {
		cidade := core.DetectarCidade(c.Endereco)
		rua := core.LimparRua(c.Endereco)

		resultado, ok, err := core.Geocode(rua, cidade)
		time.Sleep(1100 * time.Millisecond) // respeita 1 req/s do Nominatim

		if err != nil {
			log.Printf("[%d/%d] erro geocodificando %s (cod %d): %v", i+1, len(clientes), c.Nome, c.Codigo, err)
			erros++
			continue
		}
		if !ok {
			semLocalizacao++
			continue
		}

		distancia := math.Hypot(resultado.Lat-c.Lat, resultado.Lon-c.Lon)

		if resultado.CidadeConfere && distancia > 0.05 {
			log.Printf("ATUALIZANDO cod %d %q: (%.4f,%.4f) -> (%.4f,%.4f)", c.Codigo, c.Nome, c.Lat, c.Lon, resultado.Lat, resultado.Lon)
			c.Lat = resultado.Lat
			c.Lon = resultado.Lon
			c.Aproximado = false
			if err := core.UpsertCliente(db, c); err != nil {
				log.Printf("erro atualizando cod %d: %v", c.Codigo, err)
				erros++
				continue
			}
			atualizados++
		} else if !resultado.CidadeConfere && !c.Aproximado {
			log.Printf("MARCANDO APROXIMADO cod %d %q (cidade esperada %q nao confere no resultado)", c.Codigo, c.Nome, cidade)
			c.Aproximado = true
			if err := core.UpsertCliente(db, c); err != nil {
				log.Printf("erro atualizando cod %d: %v", c.Codigo, err)
				erros++
				continue
			}
			marcadosAproximado++
		} else {
			semMudanca++
		}
	}

	log.Printf("concluido: %d atualizados, %d marcados aproximado, %d sem mudanca, %d sem localizacao, %d erros",
		atualizados, marcadosAproximado, semMudanca, semLocalizacao, erros)
}
