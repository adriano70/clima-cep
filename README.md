# Clima por CEP

API HTTP em Go que recebe um CEP brasileiro, localiza a cidade pelo [ViaCEP](https://viacep.com.br/) e consulta a temperatura atual na [WeatherAPI](https://www.weatherapi.com/). A resposta contém Celsius, Fahrenheit e Kelvin.

- Repositório: <https://github.com/adriano70/clima-cep>
- URL no Cloud Run: <https://clima-cep-747741575647.southamerica-east1.run.app>
- Projeto GCP: `gen-lang-client-0004403156`
- Região: `southamerica-east1`

## Contrato HTTP

### Consultar clima

```http
GET /weather/{cep}
```

Exemplo:

```bash
curl http://localhost:8080/weather/01001000
```

Resposta `200 OK`:

```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.65
}
```

O serviço aceita somente oito dígitos ASCII, sem hífen ou espaços.

| Cenário | Status | Corpo |
| --- | ---: | --- |
| CEP em formato inválido | `422` | `invalid zipcode` |
| CEP válido, mas inexistente | `404` | `can not find zipcode` |
| ViaCEP ou WeatherAPI indisponível | `502` | `não foi possível consultar o clima` |

O ponto de acesso `GET /health` retorna `200 ok` e pode ser usado como verificação de saúde. A rota `/healthz` também está disponível localmente por compatibilidade, mas `/health` deve ser usada na URL pública do Cloud Run.

As conversões usadas são `F = C × 1,8 + 32` e `K = C + 273,15`. O deslocamento de `273,15` é a conversão física e corresponde ao exemplo do desafio (`28,5 °C = 301,65 K`).

## Configuração

| Variável | Obrigatória | Padrão | Descrição |
| --- | --- | --- | --- |
| `WEATHER_API_KEY` | Sim | — | Chave da WeatherAPI |
| `PORT` | Não | `8080` | Porta HTTP, definida automaticamente pelo Cloud Run |
| `HTTP_TIMEOUT` | Não | `5s` | Timeout das chamadas externas no formato de duração do Go |
| `VIACEP_BASE_URL` | Não | `https://viacep.com.br` | Endpoint do ViaCEP |
| `WEATHER_API_BASE_URL` | Não | `https://api.weatherapi.com/v1` | Endpoint da WeatherAPI |

Não grave a chave da WeatherAPI no código, na imagem Docker ou no repositório.

## Testes

Os testes usam servidores HTTP locais e não fazem chamadas reais às APIs:

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Executar localmente com Go

Requer Go 1.21 ou superior:

```bash
export WEATHER_API_KEY="sua-chave"
go run ./cmd/server
```

## Executar localmente com Docker

```bash
docker build -t clima-cep .
docker run --rm \
  -p 8080:8080 \
  -e WEATHER_API_KEY="sua-chave" \
  clima-cep
```

Em outro terminal:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/weather/01001000
```

## Implantação no Google Cloud Run

Os comandos abaixo pressupõem o [Google Cloud CLI](https://cloud.google.com/sdk/docs/install), um projeto com faturamento configurado e as APIs Cloud Run e Cloud Build habilitadas.

```bash
gcloud auth login
gcloud config set project SEU_PROJECT_ID

gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com

gcloud run deploy clima-cep \
  --source . \
  --region southamerica-east1 \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY="sua-chave"
```

Para obter a URL implantada:

```bash
gcloud run services describe clima-cep \
  --region southamerica-east1 \
  --format='value(status.url)'
```

Para produção, prefira armazenar `WEATHER_API_KEY` no Gerenciador de Segredos e vinculá-la ao serviço em vez de passá-la diretamente na linha de comando.

## Estrutura

```text
cmd/server/             composição e ciclo de vida do servidor
internal/config/        configuração por ambiente
internal/httpapi/       rotas e contrato HTTP
internal/viacep/        cliente de localização
internal/weather/       regras de domínio e conversões
internal/weatherapi/    cliente de temperatura atual
```

## Exemplos de requisições à URL pública

A aplicação está disponível em:

```text
https://clima-cep-747741575647.southamerica-east1.run.app
```

Para facilitar os testes, defina a URL em uma variável:

```bash
SERVICE_URL="https://clima-cep-747741575647.southamerica-east1.run.app"
```

### Verificar a saúde

```bash
curl -i "$SERVICE_URL/health"
```

Resposta esperada:

```http
HTTP/2 200

ok
```

### Consultar um CEP válido

```bash
curl -i "$SERVICE_URL/weather/01001000"
```

Resposta esperada:

```http
HTTP/2 200
Content-Type: application/json; charset=utf-8

{"temp_C":19.3,"temp_F":66.74,"temp_K":292.45}
```

As temperaturas variam conforme as condições meteorológicas no momento da consulta.

### Consultar um CEP inexistente

```bash
curl -i "$SERVICE_URL/weather/99999999"
```

Resposta esperada:

```http
HTTP/2 404

can not find zipcode
```

### Consultar um CEP em formato inválido

```bash
curl -i "$SERVICE_URL/weather/01001-000"
```

Resposta esperada:

```http
HTTP/2 422

invalid zipcode
```

### Usar um método não permitido

```bash
curl -i -X POST "$SERVICE_URL/weather/01001000"
```

Resposta esperada:

```http
HTTP/2 405
Allow: GET

método não permitido
```
