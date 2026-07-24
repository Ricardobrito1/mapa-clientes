package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"mapa-clientes-api/internal/core"
)

func main() {
	// Le variaveis de api/.env se existir (dev local). Em producao (Render),
	// o arquivo nao existe e as variaveis vem do dashboard - por isso o erro
	// e ignorado aqui.
	_ = godotenv.Load()

	db, err := core.OpenDB()
	if err != nil {
		log.Fatalf("falha ao conectar no banco: %v", err)
	}
	defer db.Close()

	if err := core.EnsureSchema(db); err != nil {
		log.Fatalf("falha ao criar schema: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /clientes", handleLista(db))
	mux.HandleFunc("GET /clientes/busca", handleBusca(db))
	mux.HandleFunc("POST /clientes", handleCadastro(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("api ouvindo na porta %s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// withCORS libera o front-end (Cloudflare Pages) a chamar essa API.
func withCORS(next http.Handler) http.Handler {
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "*"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func handleLista(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientes, err := core.ListaClientes(db)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, clientes)
	}
}

func handleBusca(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		termo := r.URL.Query().Get("q")
		if strings.TrimSpace(termo) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "parametro q e obrigatorio"})
			return
		}
		clientes, err := core.BuscaClientes(db, termo)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, clientes)
	}
}

type novoClienteRequest struct {
	Nome     string   `json:"nome"`
	Endereco string   `json:"endereco"`
	Lat      *float64 `json:"lat,omitempty"`
	Lon      *float64 `json:"lon,omitempty"`
}

func handleCadastro(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req novoClienteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "json invalido"})
			return
		}
		if strings.TrimSpace(req.Nome) == "" || strings.TrimSpace(req.Endereco) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "nome e endereco sao obrigatorios"})
			return
		}

		nome := strings.ToUpper(strings.TrimSpace(req.Nome))
		endereco := strings.ToUpper(strings.TrimSpace(req.Endereco))
		cidade := core.DetectarCidade(endereco)
		rua := core.LimparRua(endereco)

		var cep, bairro string
		viaCep, achouViaCep, _ := core.ConsultaViaCep("SP", cidade, rua)
		if achouViaCep {
			cep = viaCep.Cep
			bairro = viaCep.Bairro
			if viaCep.Logradouro != "" {
				rua = viaCep.Logradouro
			}
		}

		var lat, lon float64
		aproximado := false
		if req.Lat != nil && req.Lon != nil {
			lat, lon = *req.Lat, *req.Lon
		} else {
			resultado, achouGeo, err := core.Geocode(rua, cidade)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"erro": "falha ao geocodificar: " + err.Error()})
				return
			}
			if !achouGeo {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
					"erro":   "nao foi possivel geocodificar automaticamente, informe lat/lon manualmente",
					"cidade": cidade,
					"rua":    rua,
					"cep":    cep,
					"bairro": bairro,
				})
				return
			}
			lat, lon = resultado.Lat, resultado.Lon
			// O bbox usado no Nominatim cobre o estado inteiro de SP, entao um
			// resultado que nao cita a cidade esperada pode ser uma rua de
			// nome parecido em outro municipio - marca como aproximado em vez
			// de confiar cegamente.
			aproximado = !resultado.CidadeConfere
		}

		codigo, err := core.ProximoCodigo(db)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
			return
		}

		cliente := core.Cliente{
			Codigo:     codigo,
			Nome:       nome,
			Endereco:   endereco,
			Cep:        cep,
			Bairro:     bairro,
			Lat:        lat,
			Lon:        lon,
			Aproximado: aproximado,
		}

		if err := core.InsereCliente(db, cliente); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, cliente)
	}
}
