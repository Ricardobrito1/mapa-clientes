package core

type Cliente struct {
	Codigo     int     `json:"codigo"`
	Nome       string  `json:"nome"`
	Endereco   string  `json:"endereco"`
	Cep        string  `json:"cep"`
	Bairro     string  `json:"bairro"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Aproximado bool    `json:"aproximado"`
}
