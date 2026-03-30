# DSP Demo

English: This repository contains a minimal DSP prototype that receives a bid request, filters eligible campaigns, ranks them, and returns the best bid.

Русский: Этот репозиторий содержит минимальный прототип DSP, который принимает bid request, фильтрует подходящие кампании, ранжирует их и возвращает лучшую ставку.

## Project Structure

English:
- `cmd/app/main.go` starts the HTTP server, creates demo campaigns in memory, and wires dependencies.
- `internal/transport/http/handler.go` decodes `/bid` requests and maps engine results to HTTP responses.
- `internal/engine/filter.go` applies cheap eligibility checks.
- `internal/engine/scorer.go` converts eligible campaigns into comparable scores.
- `internal/engine/engine.go` orchestrates filtering, scoring, and deterministic winner selection.
- `internal/model/*.go` contains request, campaign, and response DTOs.

Русский:
- `cmd/app/main.go` запускает HTTP-сервер, создаёт демонстрационные кампании в памяти и связывает зависимости.
- `internal/transport/http/handler.go` декодирует запросы `/bid` и преобразует результат движка в HTTP-ответ.
- `internal/engine/filter.go` выполняет дешёвые eligibility-проверки.
- `internal/engine/scorer.go` переводит допустимые кампании в сравнимые score.
- `internal/engine/engine.go` координирует фильтрацию, скоринг и детерминированный выбор победителя.
- `internal/model/*.go` содержит DTO для запроса, кампании и ответа.

## Request Flow

English:
1. Client sends `POST /bid` with a JSON body matching `model.BidRequest`.
2. HTTP handler decodes the payload. Invalid JSON returns `400 Bad Request`.
3. Handler passes the request and the current campaign list to `engine.Engine`.
4. `TargetingFilter` removes campaigns that do not match:
   - `SiteID`
   - `DeviceType`
   - `FloorPrice` (`campaign.Price` must be greater than or equal to request floor)
5. `Scorer` assigns a numeric score to each remaining campaign.
6. Engine picks the best campaign:
   - higher score wins
   - if scores are equal, higher price wins
   - if price is also equal, lexicographically smaller `ID` wins
7. If no campaign survives, handler returns `204 No Content`.
8. If a winner exists, handler returns `200 OK` with `model.BidResponse`.

Русский:
1. Клиент отправляет `POST /bid` с JSON-телом в формате `model.BidRequest`.
2. HTTP handler декодирует payload. Некорректный JSON приводит к `400 Bad Request`.
3. Handler передаёт запрос и текущий список кампаний в `engine.Engine`.
4. `TargetingFilter` отбрасывает кампании, которые не совпадают по:
   - `SiteID`
   - `DeviceType`
   - `FloorPrice` (`campaign.Price` должен быть больше либо равен floor из запроса)
5. `Scorer` назначает каждой оставшейся кампании числовой score.
6. Движок выбирает лучшую кампанию:
   - выигрывает больший score
   - при равных score выигрывает большая цена
   - если цена тоже равна, выигрывает лексикографически меньший `ID`
7. Если ни одна кампания не прошла фильтр, handler возвращает `204 No Content`.
8. Если победитель найден, handler возвращает `200 OK` с `model.BidResponse`.

## Current Design Choices

English:
- Campaigns are stored in memory inside `main.go`; there is no database or repository layer in use yet.
- The default scoring strategy is `PriceScorer`, so the highest price wins among eligible campaigns.
- Engine dependencies are injectable, which makes it possible to replace filtering or scoring logic in tests and future integrations.

Русский:
- Кампании хранятся в памяти внутри `main.go`; база данных или полноценный repository layer пока не используются.
- Стратегия скоринга по умолчанию — `PriceScorer`, поэтому среди допустимых кампаний побеждает максимальная цена.
- Зависимости движка инъектируются, поэтому фильтр и скоринг можно подменять в тестах и будущих интеграциях.

## Run And Test

English:
```bash
go run ./cmd/app
go test ./...
```

Example request:

```bash
curl -i -X POST http://localhost:8080/bid \
  -H 'Content-Type: application/json' \
  -d '{
    "request_id":"r1",
    "imp_id":"imp1",
    "site_id":"1",
    "placement_id":"p1",
    "floor_price":1.0,
    "user_id":"u1",
    "device_type":"mobile",
    "ts":1710000000
  }'
```

## Observability

English:
- The service now exposes Prometheus metrics on `GET /metrics` on the same HTTP port as `/bid`.
- Start the service with `go run ./cmd/app`, then open `http://localhost:8080/metrics`.
- Successful bid flow is exported to metrics instead of per-request happy-path logs.
- There are no downstream dependency calls in the current DSP implementation, so downstream metrics are not emitted yet.

Русский:
- Сервис теперь отдаёт Prometheus-метрики на `GET /metrics` на том же порту, что и `/bid`.
- Запустите сервис командой `go run ./cmd/app`, затем откройте `http://localhost:8080/metrics`.
- Успешный поток обработки ставок теперь уходит в метрики, а не в per-request happy-path логи.
- Во внешние зависимости текущая реализация DSP не ходит, поэтому downstream-метрики пока не эмитятся.

### Prometheus Config

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: dsp
    static_configs:
      - targets:
          - localhost:8080
```

### Metrics Overview

English:
- `http_requests_total{route,method,status}`: total HTTP requests by normalized route template, method, and response status.
- `http_request_duration_seconds{route,method,status}`: HTTP latency histogram for normalized routes.
- `http_requests_in_flight`: current number of in-flight HTTP requests.
- `dsp_bid_requests_total`: total number of incoming bid requests before validation.
- `dsp_bid_responses_total{result}`: total DSP outcomes with `result` values such as `bid`, `no_bid`, `invalid_request`, `error`.
- `dsp_bid_processing_duration_seconds{result}`: DSP processing latency histogram by final result.

Русский:
- `http_requests_total{route,method,status}`: общее число HTTP-запросов по нормализованному шаблону route, методу и статусу ответа.
- `http_request_duration_seconds{route,method,status}`: гистограмма задержки HTTP-запросов по нормализованным route.
- `http_requests_in_flight`: текущее количество HTTP-запросов в обработке.
- `dsp_bid_requests_total`: общее количество входящих bid requests до валидации.
- `dsp_bid_responses_total{result}`: итоговые DSP-результаты с `result` вроде `bid`, `no_bid`, `invalid_request`, `error`.
- `dsp_bid_processing_duration_seconds{result}`: гистограмма времени обработки DSP-запросов по итоговому результату.

Русский:
```bash
go run ./cmd/app
go test ./...
```

Пример запроса:

```bash
curl -i -X POST http://localhost:8080/bid \
  -H 'Content-Type: application/json' \
  -d '{
    "request_id":"r1",
    "imp_id":"imp1",
    "site_id":"1",
    "placement_id":"p1",
    "floor_price":1.0,
    "user_id":"u1",
    "device_type":"mobile",
    "ts":1710000000
  }'
```
