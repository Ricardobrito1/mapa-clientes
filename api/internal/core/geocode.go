package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const nominatimURL = "https://nominatim.openstreetmap.org/search"
const viewbox = "-53.2,-19.7,-44.0,-25.3" // bbox regiao de Sao Paulo

var cidadesConhecidas = []string{
	"GUARULHOS", "SANTO ANDRE", "SANTO ANDRÉ", "SAO BERNARDO DO CAMPO",
	"SÃO BERNARDO DO CAMPO", "DIADEMA", "FERRAZ DE VASCONCELOS", "INDAIATUBA",
	"SOROCABA", "ITAQUAQUECETUBA", "ARUJA", "ARUJÁ", "SANTA ISABEL", "IGARATA",
	"IGARATÁ", "MOGI DAS CRUZES", "SUZANO", "POA", "POÁ", "MAUA", "MAUÁ",
	"RIBEIRAO PIRES", "RIBEIRÃO PIRES", "EMBU DAS ARTES", "EMBU", "BARUERI",
	"OSASCO", "COTIA", "CARAPICUIBA", "CARAPICUÍBA", "ITAPEVI",
	"TABOAO DA SERRA", "TABOÃO DA SERRA", "FRANCO DA ROCHA", "CAIEIRAS",
	"FRANCISCO MORATO", "SAO CAETANO DO SUL", "SÃO CAETANO DO SUL", "JANDIRA",
	"ITAPECERICA DA SERRA", "GUARAREMA", "SALESOPOLIS", "SALESÓPOLIS",
	"SAO PAULO", "SÃO PAULO",
}

var reParenteses = regexp.MustCompile(`\([^)]*\)`)
var reComplemento = regexp.MustCompile(`(?i)\b(LOJA|LJ|ANEXO|GALP[ÃA]O|TERREO|TÉRREO|FRENTE|APTO|CASA|SAL[ÃA]O|CONJUNTO|BLOCO)\s*\.?\s*\d*\b`)
var reSeparadores = regexp.MustCompile(`[,\-–]`)
var reComecaComDigito = regexp.MustCompile(`^\d`)

func DetectarCidade(endereco string) string {
	up := strings.ToUpper(endereco)
	for _, c := range cidadesConhecidas {
		if strings.Contains(up, c) {
			return strings.Title(strings.ToLower(c))
		}
	}
	return "São Paulo"
}

func LimparRua(endereco string) string {
	s := reParenteses.ReplaceAllString(endereco, " ")
	s = reComplemento.ReplaceAllString(s, " ")

	partsRaw := reSeparadores.Split(s, -1)
	var parts []string
	for _, p := range partsRaw {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}

	if len(parts) == 0 {
		return strings.TrimSpace(s)
	}

	rua := parts[0]
	if len(parts) > 1 && reComecaComDigito.MatchString(parts[1]) {
		rua = rua + " " + parts[1]
	}
	return strings.TrimSpace(rua)
}

type nominatimResultado struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

// Geocode busca lat/lon de uma rua+cidade via Nominatim, restrito a bbox de SP.
// Retorna ok=false se nao encontrar nada.
func Geocode(rua, cidade string) (lat, lon float64, ok bool, err error) {
	params := url.Values{}
	params.Set("street", rua)
	params.Set("city", cidade)
	params.Set("state", "São Paulo")
	params.Set("country", "Brasil")
	params.Set("format", "json")
	params.Set("limit", "1")
	params.Set("countrycodes", "br")
	params.Set("viewbox", viewbox)
	params.Set("bounded", "1")

	// url.Values.Encode() usa "+" pra espaco (application/x-www-form-urlencoded),
	// mas o Nominatim so interpreta espaco corretamente como "%20" nos campos
	// estruturados (city/street) - com "+" ele ignora a cidade e retorna
	// qualquer rua de nome parecido em outro municipio. Troca e seguro: um "+"
	// literal nos dados originais ja teria sido escapado como "%2B" pelo Encode.
	query := strings.ReplaceAll(params.Encode(), "+", "%20")
	fmt.Println("DEBUG geocode url:", nominatimURL+"?"+query)

	req, err := http.NewRequest(http.MethodGet, nominatimURL+"?"+query, nil)
	if err != nil {
		return 0, 0, false, err
	}
	req.Header.Set("User-Agent", "mapa-clientes-api/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, false, err
	}

	var resultados []nominatimResultado
	if err := json.Unmarshal(body, &resultados); err != nil {
		return 0, 0, false, fmt.Errorf("resposta invalida do nominatim: %w", err)
	}
	if len(resultados) == 0 {
		return 0, 0, false, nil
	}

	var latF, lonF float64
	fmt.Sscanf(resultados[0].Lat, "%f", &latF)
	fmt.Sscanf(resultados[0].Lon, "%f", &lonF)
	return latF, lonF, true, nil
}
