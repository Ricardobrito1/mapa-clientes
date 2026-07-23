# Mapa de Clientes

Mapa interativo dos clientes, com busca e botão pra abrir a rota no Waze. 100% estático (sem backend, sem banco de dados, sem custo).

**Site:** https://SEU-USUARIO.github.io/mapa-clientes/  (substitua depois de configurar o GitHub Pages)

## Como funciona

- `index.html` — a página em si (mapa Leaflet + lista lateral + busca). Não precisa mexer aqui na maioria das vezes.
- `clientes.js` — os dados dos clientes (nome, endereço, latitude/longitude). **É esse arquivo que você edita pra adicionar/remover cliente.**
- `scripts/adicionar_cliente.py` — script auxiliar que geocodifica um endereço novo (acha a lat/lon automaticamente) e já gera a linha pronta pra colar.

## Adicionar um cliente novo

### Opção A — usando o script (recomendado)

```bash
cd mapa-clientes
pip install requests
python3 scripts/adicionar_cliente.py
```

Ele pergunta nome e endereço, geocodifica, e imprime um bloco JSON pronto. Cola esse bloco dentro do array `CLIENTES` em `clientes.js` (antes do `]` final), salva, e sobe pro GitHub:

```bash
git add clientes.js
git commit -m "adiciona cliente X"
git push
```

O GitHub Pages atualiza sozinho em ~1 minuto.

### Opção B — na mão

Abre `clientes.js`, copia um objeto existente como modelo e ajusta os campos:

```json
{"codigo": 9999, "nome": "NOME DO CLIENTE", "endereco": "ENDEREÇO COMPLETO", "lat": -23.55, "lon": -46.63, "aproximado": false}
```

Pra achar lat/lon manualmente: abre o Google Maps, clica com botão direito no local exato, copia as coordenadas que aparecem.

## Cuidado com geocodificação errada

O serviço de busca de endereço (Nominatim/OpenStreetMap) às vezes confunde nome de bairro com nome de cidade de outro estado (ex: bairro "São Mateus" em SP vs. cidade "São Mateus" no ES). O script já restringe a busca pra dentro do estado de São Paulo, mas **sempre vale conferir o pino no mapa depois de adicionar** — se ele aparecer em lugar estranho, corrige a mão com o Google Maps.

## Rodando localmente antes de publicar

```bash
python3 -m http.server 8000
```

Abre `http://localhost:8000` no navegador.

## Publicar / atualizar no GitHub Pages

1. `git add . && git commit -m "atualiza mapa" && git push`
2. Settings → Pages → confirma que branch `main` / `root` tá selecionado
3. Espera ~1 min e o link atualiza sozinho
