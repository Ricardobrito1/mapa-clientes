#!/usr/bin/env python3
"""
Script pra adicionar um cliente novo ao mapa.

Uso:
    python3 scripts/adicionar_cliente.py

Ele pergunta o nome e o endereço, geocodifica (acha lat/lon)
e já deixa a linha pronta pra você colar no clientes.js.

Requisitos: pip install requests
"""
import re
import json
import requests

HEADERS = {"User-Agent": "mapa-clientes-script/1.0"}
URL = "https://nominatim.openstreetmap.org/search"
VIEWBOX = "-53.2,-19.7,-44.0,-25.3"  # regiao de Sao Paulo, evita match errado em outro estado

CIDADES_CONHECIDAS = [
    "GUARULHOS", "SANTO ANDRE", "SANTO ANDRÉ", "SAO BERNARDO DO CAMPO",
    "SÃO BERNARDO DO CAMPO", "DIADEMA", "FERRAZ DE VASCONCELOS", "INDAIATUBA",
    "SOROCABA", "ITAQUAQUECETUBA", "ARUJA", "ARUJÁ", "SANTA ISABEL", "IGARATA",
    "IGARATÁ", "MOGI DAS CRUZES", "SUZANO", "POA", "POÁ", "MAUA", "MAUÁ",
    "RIBEIRAO PIRES", "RIBEIRÃO PIRES", "EMBU DAS ARTES", "EMBU", "BARUERI",
    "OSASCO", "COTIA", "CARAPICUIBA", "CARAPICUÍBA", "ITAPEVI",
    "TABOAO DA SERRA", "TABOÃO DA SERRA", "FRANCO DA ROCHA", "CAIEIRAS",
    "FRANCISCO MORATO", "SAO CAETANO DO SUL", "SÃO CAETANO DO SUL", "JANDIRA",
    "ITAPECERICA DA SERRA", "GUARAREMA", "SALESOPOLIS", "SALESÓPOLIS",
    "SAO PAULO", "SÃO PAULO"
]

def detectar_cidade(endereco):
    up = endereco.upper()
    for c in CIDADES_CONHECIDAS:
        if c in up:
            return c.title()
    return "São Paulo"

def limpar_rua(endereco):
    s = re.sub(r'\([^)]*\)', ' ', endereco)
    s = re.sub(r'\b(LOJA|LJ|ANEXO|GALP[ÃA]O|TERREO|TÉRREO|FRENTE|APTO|CASA|SAL[ÃA]O|CONJUNTO|BLOCO)\s*\.?\s*\d*\b', ' ', s, flags=re.IGNORECASE)
    parts = re.split(r'[,\-–]', s)
    parts = [p.strip() for p in parts if p.strip()]
    rua = parts[0] if parts else s
    if len(parts) > 1 and re.match(r'^\d', parts[1]):
        rua = rua + ' ' + parts[1]
    return rua.strip()

def geocode(rua, cidade):
    params = {
        "street": rua, "city": cidade, "state": "São Paulo", "country": "Brasil",
        "format": "json", "limit": 1, "countrycodes": "br",
        "viewbox": VIEWBOX, "bounded": 1
    }
    r = requests.get(URL, params=params, headers=HEADERS, timeout=10)
    data = r.json()
    if data:
        return float(data[0]["lat"]), float(data[0]["lon"])
    return None

def proximo_codigo():
    with open('clientes.js', encoding='utf-8') as f:
        content = f.read()
    json_str = content.split('= ', 1)[1].rsplit(';', 1)[0]
    data = json.loads(json_str)
    return max(d['codigo'] for d in data) + 1

if __name__ == '__main__':
    nome = input("Nome do cliente: ").strip().upper()
    endereco = input("Endereço completo: ").strip().upper()

    cidade = detectar_cidade(endereco)
    rua = limpar_rua(endereco)
    print(f"\nBuscando: '{rua}' em '{cidade}'...")

    res = geocode(rua, cidade)
    if not res:
        print("Não consegui localizar automaticamente. Confirma o endereço ou")
        print("digita a latitude/longitude manualmente (ex: pegando do Google Maps).")
        lat = input("Latitude (ou Enter pra deixar sem): ").strip()
        lon = input("Longitude: ").strip()
        lat = float(lat) if lat else None
        lon = float(lon) if lon else None
    else:
        lat, lon = res
        print(f"Encontrado: {lat}, {lon}")

    codigo = proximo_codigo()
    entry = {
        "codigo": codigo,
        "nome": nome,
        "endereco": endereco,
        "lat": lat,
        "lon": lon,
        "aproximado": False
    }

    print("\n--- Cole isso dentro do array em clientes.js, antes do ']' final ---\n")
    print(json.dumps(entry, ensure_ascii=False) + ",")
