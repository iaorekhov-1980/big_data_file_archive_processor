# Actual State

Текущее состояние проекта `big_data_file_archive_processor`.
Обновляется после каждой сессии на базе output.

**Последнее обновление:** после Phase 5.1 (recursive scanner)

---

## Обзор

Go-сервис для сканирования, дедупликации и очистки файлов на Yandex Disk с хранением метаданных в PostgreSQL.

- **Язык:** Go 1.25
- **БД:** PostgreSQL (pgx/v5, pgxpool)
- **Внешний API:** Yandex Disk REST API
- **Тесты:** testify
- **Конфиг:** TOML + .env

---

## Реализовано

### Phase 1 — Структура проекта ✅
- Директории `cmd/`, `internal/`, `migrations/`, `docs/`, `testdata/`
- `go.mod` (модуль `github.com/iaorekhov-1980/big_data_file_archive_processor`)
- `.gitignore`, `README.md`, `.env.example`

### Phase 2 — Модели и интерфейсы ✅
- `internal/models/models.go` — `File`, `FilePath`, `SourceProcessing` с lifecycle-методами
- `internal/repository/interface.go` — интерфейс `Repository` + типы ошибок (`NotFoundError`, `DuplicateError`, `RepositoryError`)
- `internal/config/config.go` — загрузка конфигурации из env, валидация

### Phase 3 — Слой БД ✅
- `migrations/000001_init.up.sql` / `.down.sql` — таблицы `files`, `file_paths` (композитный PK `(path, source_id)`), `source_processing`
- `internal/repository/postgres/` — реализация на pgx/v5:
  - `file_repository.go`, `filepath_repository.go`, `sourceprocessing_repository.go`
  - `postgres.go` — композиция репозиториев
  - `utils.go` — хелперы (детект duplicate key)
  - Интерфейс `Querier` (pool + tx), методы `WithTx(tx pgx.Tx)`
- `internal/repository/factory/factory.go` — фабрика репозиториев
- `internal/testutils/` — хелперы тестов (config, db_test_helper, setup_test_config, test_db_connection)
- Интеграционные тесты репозиториев (file, filepath, sourceprocessing, full lifecycle)

### Phase 4 — Клиент Yandex Disk ✅
- `internal/disk/interface.go` — интерфейс `DiskClient` + `Resource`
- `internal/disk/yandexdisk.go` — транспорт, HTTP-клиент, `DiskError`, rate limiting
- `internal/disk/operations.go` — `ListFiles`, `GetFolderContents`, `GetFileInfo`
- `internal/disk/factory.go` — фабрика клиента
- `internal/disk/yandexdisk_test.go`, `operations_test.go`, `connectivity_test.go` — тесты
- Конфиг: `YandexDiskBaseURL`, `YandexDiskTimeout`, `YandexDiskRateLimitDelayMs`

### Phase 5.0 — Сервисный слой (частично) ✅
- `internal/service/tx.go` — `TxManager` + `TxBundle` (транзакции через pgxpool)
- `internal/service/retry.go` — `DoWithRetry`, `RetryConfig`, `RetryableError` (экспоненциальный backoff)
- `internal/service/scanner.go` — `ScanService`:
  - **Рекурсивный** обход дерева папок (не пагинация)
  - Стриппинг префиксов путей (`stripPrefix`: убирает `disk:` и путь источника)
  - Транзакционная обработка файлов
  - Идемпотентность (проверка `ScanStatus == completed`)
  - Retry внешних вызовов
- `internal/service/scan_integration_test.go` — интеграционный тест на реальном Yandex Disk

### Документация ✅
- `docs/input/` — 13 файлов ТЗ/промптов
- `docs/output/` — 2 файла результатов
- `docs/README.md` — индекс
- `.continue/rules/project.md` — правила проекта для AI
- `.continueignore` — исключения индексации

---

## Структура (текущая)

```
cmd/
  scan_source/          # пусто
  cleanup_duplicates/   # пусто
  cleanup_folders/      # пусто
  test/                 # main.go (setup, test-connection)
internal/
  config/               # config.go, config_test.go
  disk/                 # interface, yandexdisk, operations, factory + тесты
  models/               # models.go
  repository/
    interface.go
    factory/            # factory.go
    postgres/           # file, filepath, sourceprocessing, postgres, utils + тесты
  service/              # tx.go, retry.go, scanner.go, scan_integration_test.go
  testutils/            # config, db_test_helper, setup_test_config, test_db_connection
migrations/             # 000001_init.up/down.sql
testdata/               # фикстуры (ods/22_05_2026_test_folder)
docs/                   # input/, output/, overview/, README.md
```

---

## Ключевые технические решения

1. **Рекурсивное сканирование** вместо пагинации — `ScanService` обходит дерево папок через `ListFiles(folderPath, 0, pageSize)`.
2. **Относительные пути** — в БД хранятся без `disk:` и без пути источника (`readme.txt`, `archives/backup.zip`).
3. **Транзакции на уровне сервиса** — `TxManager.WithTransaction` + `TxBundle` с `WithTx`-репозиториями.
4. **Идемпотентность** — проверка статуса в `SourceProcessing` перед работой.
5. **Типизированные ошибки** — `NotFoundError`, `DuplicateError`, `DiskError`, проверка через `errors.As`.

---

## Известные ограничения / TODO

- `CleanupService` не реализован (Phase 5.1 — следующая сессия)
- CLI-команды (`cmd/*`) пусты
- Unit-тесты сервисов с моками отсутствуют
- `SourceProcessing.CurrentOffset` больше не используется как offset пагинации (наследие старого подхода)
