# Target State

Список того, что необходимо реализовать в проекте `big_data_file_archive_processor`.
Обновляется после каждой сессии на базе output.

**Последнее обновление:** после Phase 5.1 (recursive scanner)

---

## Ближайшие задачи

### Phase 5.1 — CleanupService (дедупликация путей) 🔜
**ТЗ:** `docs/input/13_phase5_1_cleanup_service.md`

- `internal/service/cleaner.go` — `CleanupService.CleanupDuplicates(ctx, sourceID, dryRun)`
  - Группировка активных путей по hash
  - Оставить самый старый, деактивировать остальные
  - Dry-run режим
  - Идемпотентность
- `internal/service/cleaner_test.go` — unit-тесты с моками
- `internal/service/clean_integration_test.go` — интеграционный тест
- `cmd/cleanup_duplicates/main.go` — CLI-команда

### Phase 5.2 — Unit-тесты сервисов
- `internal/service/scanner_test.go` — unit-тесты ScanService с моками `DiskClient` и `Repository`
  - Пустая папка, несколько папок, вложенность
  - Пропуск существующих путей
  - Retry при transient-ошибках
  - Отмена контекста
  - Обновление статусов `SourceProcessing`

### Phase 5.3 — CleanupService: очистка пустых папок
- Расширение `CleanupService` методом `CleanupFolders(ctx, sourceID, dryRun)`
- Рекурсивный поиск и удаление пустых папок через `DiskClient`
- `cmd/cleanup_folders/main.go` — CLI-команда

---

## Среднесрочные задачи

### Phase 6 — CLI-команды
- `cmd/scan_source/main.go` — сканирование источника
  - Флаги: `--source-id`, `--source-folder`, `--resume-offset`
- `cmd/cleanup_duplicates/main.go` — дедупликация (см. Phase 5.1)
- `cmd/cleanup_folders/main.go` — очистка папок (см. Phase 5.3)
- Единый стиль: загрузка конфига, подключение к БД, логирование, exit-коды

### Phase 7 — Оркестратор
- `internal/service/orchestrator.go` — координация фаз (scan → cleanup → folders)
- Управление последовательностью операций для источника
- Обработка частичных сбоев

---

## Долгосрочные задачи (DDS / Data Mart)

> Из `docs/input/06_dds_and_datamart_design.md` — пока не детализировано.

### DDS (слой уникальных файлов)
- Копирование уникальных файлов в `/dds/<hash[:2]>/<hash[2:4]>/<hash>`
- Поле `dds_path` в таблице `files`
- Стратегия копирования (сразу / отдельным этапом / через очередь)

### Data Mart (витрина медиафайлов)
- Извлечение даты съёмки (EXIF для фото, метаданные для видео)
- Группировка по годам/месяцам
- Структура `/datamart/YYYY/MM/`
- Выбор подхода (физическое копирование / zero-byte / логическая витрина)

---

## Инфраструктура и качество

### Тестирование
- Unit-тесты сервисов с моками (Phase 5.2)
- E2E-тесты полного цикла
- Покрытие критичных путей

### Документация
- Обновление `README.md` (инструкции по запуску CLI)
- Документация API-клиента и сервисов

### Возможные улучшения
- Метрики/мониторинг
- Circuit breaker для внешних API
- Параллельная обработка файлов
- Батчевые операции БД

---

## Технический долг

- `SourceProcessing.CurrentOffset` — наследие пагинационного подхода, не используется при рекурсивном сканировании. Решить: удалить или переиспользовать.
- `cmd/scan_source/`, `cmd/cleanup_duplicates/`, `cmd/cleanup_folders/` — пустые директории.
