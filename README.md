# Services Healthcheck and Alert

## Архитектура

```
+-------------------+
|  web/index.html   |  <-- Веб-интерфейс (UI)
+-------------------+
          |
          v (HTTP)
+-------------------+
|   internal/api    |  <-- REST API (CRUD, история)
+-------------------+
          |
          v
+-------------------+
| internal/storage  |  <-- Хранилище (BoltDB/Postgres)
+-------------------+
          ^
          |
+-------------------+
| internal/monitor  |  <-- Мониторинг (Scheduler, Ping, динамика)
+-------------------+
          |
          v
+-------------------+
|  internal/alert   |  <-- Алерты (Email, Telegram)
+-------------------+

Конфигурирование: через internal/config (config.yaml/env)
Запуск и инициализация: internal/app
Reverse proxy: nginx (порт 80)
```

---

## Быстрый старт (Docker Compose)

```sh
git clone ...
cd services-healthchek-and-alert
make docker-run
```

- UI и API: http://localhost/
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- PostgreSQL: порт 5432, пользователь/пароль/db: monitor

---

## Конфигурация (config.yaml)

```yaml
smtp:
  host: smtp.example.com
  port: "587"
  username: user@example.com
  password: secret
  from: user@example.com

telegram:
  bot_token: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
  chat_id: "123456789"

alert_retries: 3                # Кол-во неудачных проверок подряд до алерта
max_alerts_per_incident: 2      # Максимум алертов подряд до первой успешной проверки
```

- Можно также использовать переменные окружения (см. исходный config.go)
- Файл config.yaml и prometheus.yml копируются в Docker-образ автоматически

---

## Web-интерфейс
- Доступен на http://localhost/
- Возможности:
  - Добавление, удаление, редактирование сайтов (интервал, тип, url)
  - Фильтры по имени, типу, статусу
  - Сортировка по столбцам, drag&drop строк
  - Просмотр истории проверок с графиком (Chart.js)
  - Экспорт истории в CSV
  - Все изменения применяются динамически, без перезапуска

---

## API
- `GET    /sites` — список сайтов
- `POST   /sites` — добавить сайт (id генерируется автоматически, если не указан)
- `GET    /sites/{id}` — получить сайт
- `PUT    /sites/{id}` — обновить сайт (интервал, url и др. меняются "на лету")
- `DELETE /sites/{id}` — удалить сайт
- `GET    /sites/{id}/history?limit=N` — история проверок

### Формат сайта (JSON)
```json
{
  "id": "unique-id",
  "name": "My Service",
  "url": "https://example.com",
  "check_type": "http", // или "tcp"
  "interval_seconds": 60
}
```

---

## Логирование
- Все ключевые действия (CRUD, мониторинг, алерты, ошибки, бизнес-логика) логируются в stdout
- Пример:
  - `[INFO] POST /sites` — получен запрос
  - `[INFO] Site added: ...` — сайт добавлен
  - `[PING][HTTP][OK] ...` — успешная проверка
  - `[RETRY] ...` — увеличение счетчика ретраев
  - `[ALERT] ...` — отправка алерта

---

## Настройка алертов
- Email и/или Telegram (можно оба, выбирается первый доступный)
- Порог срабатывания — через `alert_retries` (config.yaml/env)
- Максимум алертов подряд — через `max_alerts_per_incident` (config.yaml/env)
- Формат сообщений: имя сайта, статус, код, ошибка
- После первой успешной проверки счетчики сбрасываются

---

## Бизнес-логика и динамика
- Мониторинг HTTP/TCP с заданным интервалом (можно менять "на лету")
- История проверок хранится в БД
- "Умные" алерты: срабатывают только после N неудач подряд, максимум M алертов подряд
- Все изменения (интервал, url, тип) применяются без перезапуска приложения

---

## Интеграция с Prometheus и Grafana
- Метрики Prometheus доступны на http://localhost/metrics
- В docker-compose уже есть сервисы prometheus (порт 9090) и grafana (порт 3000)
- Пример метрик:
  - `service_check_success_total`, `service_check_fail_total` — успешные/неуспешные проверки
  - `service_check_duration_ms` — гистограмма времени ответа
  - `service_alert_sent_total` — количество отправленных алертов
  - `service_sites_tracked` — количество отслеживаемых сайтов
- В Grafana можно добавить Prometheus как источник данных (`http://prometheus:9090`) и импортировать dashboard из grafana-dashboard.json

---

## Docker
- `Dockerfile` копирует все необходимые файлы: бинарник, web, config.yaml, prometheus.yml
- `docker-compose.yml` включает app, db, nginx (reverse proxy), prometheus, grafana
- Makefile для удобства

---

## TODO/Идеи для развития
- UI: drag&drop сохранение порядка, группировка, расширенные графики
- Поддержка нескольких каналов алертов одновременно
- SLA/аптайм-метрики, алерты по SLA
- Интеграция с внешними CMDB/инвентаризацией