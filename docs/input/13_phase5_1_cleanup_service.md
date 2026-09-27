# Phase 5.1: CleanupService — дедупликация путей

**Цель сессии:** реализовать `CleanupService`, который находит дубликаты путей файлов в источнике и деактивирует их (оставляя один активный путь на каждый хеш).

---

## Контекст

Проект `big_data_file_archive_processor` — Go-сервис для сканирования, дедупликации и очистки файлов на Yandex Disk с метаданными в PostgreSQL.

### Что уже сделано (Phase 5.0)

- ✅ `internal/service/tx.go` — `TxManager` + `TxBundle` (транзакции через `pgxpool.Pool`)
- ✅ `internal/service/retry.go` — `DoWithRetry` с экспоненциальным backoff, `RetryConfig`, `RetryableError`
- ✅ `internal/service/scanner.go` — `ScanService` с **рекурсивным** обходом папок, стриппингом префиксов путей, транзакционной обработкой
- ✅ `internal/service/scan_integration_test.go` — интеграционный тест на реальном Yandex Disk
- ✅ `internal/repository/postgres/*` — репозитории с `WithTx(tx pgx.Tx)` и интерфейсом `Querier`

### Что НЕ сделано (эта сессия)

- ❌ `internal/service/cleaner.go` — `CleanupService`
- ❌ `internal/service/cleaner_test.go` — unit-тесты с моками
- ❌ `internal/service/clean_integration_test.go` — интеграционный тест
- ❌ `cmd/cleanup_duplicates/main.go` — CLI-команда

---

## Важные расхождения с исходным ТЗ (файл 12)

1. **Сканирование рекурсивное, не пагинированное.** `ScanService` обходит дерево папок рекурсивно через `ListFiles(folderPath, 0, pageSize)`, а не листает один список с offset. `SourceProcessing.CurrentOffset` больше не используется как offset пагинации.
2. **Пути хранятся относительными.** `stripPrefix` убирает `disk:` и путь источника. В БД: `readme.txt`, `archives/backup.zip` и т.д.
3. **`Repository` — интерфейс**, сервисы зависят от него, не от `*postgres.PostgresRepository`.

---

## Что реализовать

### 1. `internal/service/cleaner.go` — CleanupService

```go
package service

// CleanupService handles duplicate file path cleanup.
type CleanupService struct {
    repo      repository.Repository
    txManager *TxManager
    logger    *slog.Logger
}

func NewCleanupService(repo repository.Repository, txManager *TxManager, logger *slog.Logger) *CleanupService

// CleanupResult summarizes the cleanup operation.
type CleanupResult struct {
    SourceID          string
    DuplicatesFound   int64
    DuplicatesRemoved int64
    Errors            []string
}

// CleanupDuplicates finds and deactivates duplicate file paths for a source.
// A duplicate is two or more active FilePath records with the same hash but different paths.
// The record with the earliest CreatedAt is kept active; all others are marked inactive
// (is_active=false, deleted_at=NOW()).
//
// Parameters:
//   - ctx: context for cancellation
//   - sourceID: the source to clean up
//   - dryRun: if true, only report what would be done, make no DB changes
func (s *CleanupService) CleanupDuplicates(ctx context.Context, sourceID string, dryRun bool) (*CleanupResult, error)
```

**Алгоритм:**

1. Загрузить `SourceProcessing` для `sourceID`. Если нет — вернуть `NotFoundError`.
2. Проверить, что `ScanStatus == "completed"` (предусловие). Иначе — ошибка.
3. Если `CleanupStatus == "completed"` — вернуть результат немедленно (идемпотентность).
4. Пометить `StartCleanup()` и сохранить.
5. Получить все **активные** пути: `repo.GetFilePathsBySourceAndActive(ctx, sourceID, true)`.
6. Сгруппировать по `hash`.
7. Для групп с `len > 1`:
   - Отсортировать по `CreatedAt` (возрастание).
   - Первый (самый старый) — оставить активным.
   - Остальные — деактивировать через `UpdateFilePathDeletedAtByPathAndSource(ctx, path, sourceID, &now)`.
   - В `dryRun` — только логировать, не менять БД.
8. Обработать каждую группу в отдельной транзакции через `txManager.WithTransaction`.
9. По завершении: `CompleteCleanup(foldersDeleted)` — для дедупликации используем `FilesDeleted`/`FilesProcessed` или `CompleteProcessing`. **Уточнить:** какое поле использовать для счётчика деактивированных путей. Рекомендация — `CompleteProcessing(filesProcessed, filesDeleted)`, где `filesProcessed = DuplicatesFound`, `filesDeleted = DuplicatesRemoved`.
10. При ошибке — `FailCleanup(errMsg)`.

**Ключевые поведения:**
- `dryRun` — никаких изменений в БД, только подсчёт и логирование.
- Идемпотентность — повторный запуск после `completed` возвращает результат без работы.
- Транзакция на группу дубликатов (атомарность деактивации).
- Логирование через `s.logger` с полем `source_id`.

### 2. `internal/service/cleaner_test.go` — unit-тесты

Моки: `mockRepository` (реализует `repository.Repository`), `mockTxManager` или реальный `TxManager` с мок-пулом.

Сценарии:
- Нет дубликатов — ничего не делаем, `DuplicatesFound=0`.
- Одна группа дубликатов — оставить самый старый, деактивировать остальные.
- Несколько групп дубликатов.
- `dryRun` — БД не меняется, счётчики корректны.
- Источник без записей — no-op.
- Источник не отсканирован (`ScanStatus != completed`) — ошибка.
- Повторный запуск после `completed` — идемпотентность.
- Проверка обновления `SourceProcessing` (cleanup-статус).

### 3. `internal/service/clean_integration_test.go` — интеграционный тест

По образцу `scan_integration_test.go`:
- Пропуск в `-short`.
- `testutils.SetupTestDatabase` + `testutils.CleanTestData`.
- Сканировать реальный источник (или подготовить данные напрямую через репозиторий).
- Запустить `CleanupDuplicates`.
- Проверить: активных путей стало меньше, деактивированные имеют `deleted_at != nil`, `is_active=false`.
- Проверить идемпотентность повторного запуска.

### 4. `cmd/cleanup_duplicates/main.go` — CLI-команда

- Загрузить конфиг (`internal/config`).
- Подключиться к БД (`pgxpool`).
- Создать `TxManager`, `CleanupService`.
- Флаги: `--source-id` (обязательный), `--dry-run` (опционально).
- Логирование через `log/slog`.
- Корректные exit-коды.

---

## Соглашения (из `.continue/rules/project.md`)

- Ошибки типизированные: `repository.NotFoundError`, `repository.DuplicateError`. Проверка через `errors.As`.
- Логирование: `log/slog`, передаётся через конструктор.
- `ctx context.Context` — первый аргумент всех I/O-методов.
- Экспортируемые типы/функции — doc-комментарий на английском.
- Одна задача = один файл, СТОП после каждого файла.

---

## Порядок реализации

1. `internal/service/cleaner.go`
2. `internal/service/cleaner_test.go`
3. `internal/service/clean_integration_test.go`
4. `cmd/cleanup_duplicates/main.go`

## Проверка

- `go build ./...`
- `go test ./internal/service/ -short -count=1` — unit-тесты
- `go test ./internal/service/ -run TestCleanupService_Integration -v -count=1` — интеграционный

## Git

- Коммиты: `Phase 5.1: <описание>`
- Автор: `iaorekhov-1980 <i.a.orekhov@gmail.com>`
