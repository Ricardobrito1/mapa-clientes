package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type viaCepResultado struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

// ConsultaViaCep busca candidatos de endereco por UF+cidade+rua.
// Retorna o primeiro resultado (ou ok=false se nao achar nada).
// Nao substitui o Nominatim: nao tem lat/lon, so enriquece o endereco
// (rua oficial, bairro, cep) antes da geocodificacao.
func ConsultaViaCep(uf, cidade, rua string) (resultado viaCepResultado, ok bool, err error) {
	path := fmt.Sprintf("https://viacep.com.br/ws/%s/%s/%s/json/",
		url.PathEscape(uf), url.PathEscape(cidade), url.PathEscape(rua))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(path)
	if err != nil {
		return viaCepResultado{}, false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return viaCepResultado{}, false, err
	}

	var resultados []viaCepResultado
	if err := json.Unmarshal(body, &resultados); err != nil {
		// ViaCEP as vezes responde só "[]" ou um objeto de erro isolado
		// quando rua/cidade nao tem match - tratamos como "nao encontrado".
		return viaCepResultado{}, false, nil
	}
	if len(resultados) == 0 {
		return viaCepResultado{}, false, nil
	}
	return resultados[0], true, nil
}
