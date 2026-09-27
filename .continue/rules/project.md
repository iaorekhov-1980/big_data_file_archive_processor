---
name: Project Rules
description: Core conventions and architecture for big_data_file_archive_processor
alwaysApply: true
---

# Project: big_data_file_archive_processor

Go-сервис для сканирования, дедупликации и очистки файлов на Yandex Disk с хранением метаданных в PostgreSQL.

## Стек

- **Язык:** Go 1.25
- **БД:** PostgreSQL (драйвер `github.com/jackc/pgx/v5`, пул `pgxpool`)
- **Внешний API:** Yandex Disk REST API (`https://cloud-api.yandex.net/v1/disk`)
- **Тесты:** `github.com/stretchr/testify` (assert, require)
- **Конфиг:** TOML (`github.com/pelletier/go-toml/v2`), `.env` (`github.com/joho/godotenv`)

## Структура проекта

```
cmd/                          # CLI-команды (по одной на операцию)
  scan_source/                # сканирование источника
  cleanup_duplicates/         # удаление дубликатов путей
  cleanup_folders/            # очистка пустых папок
internal/
  config/                     # загрузка конфигурации
  disk/                       # Yandex Disk API клиент
  models/                     # доменные модели (File, FilePath, SourceProcessing)
  repository/                 # интерфейс Repository + типы ошибок
    postgres/                 # реализация на PostgreSQL
  service/                    # бизнес-логика (scanner, cleaner, tx, retry)
  testutils/                  # хелперы для тестов
migrations/                   # SQL-миграции
testdata/                     # тестовые данные (фикстуры)
docs/                         # документация
```

## Архитектурные принципы

1. **Слоистость:** `cmd` → `service` → `repository` → `postgres`. Сервисы не знают про SQL, репозитории не знают про бизнес-логику.
2. **Интерфейсы:** `repository.Repository` и `disk.DiskClient` — интерфейсы. Сервисы зависят от интерфейсов, не от реализаций.
3. **Транзакции:** через `service.TxManager` + `TxBundle`. Репозитории имеют методы `WithTx(querier)` для работы внутри транзакции.
4. **Retry:** внешние API-вызовы оборачиваются в `service.DoWithRetry` с экспоненциальным backoff.
5. **Идемпотентность:** повторный запуск операции не должен создавать дубликаты. Проверяй статус в `SourceProcessing` перед работой.

## Доменные модели

- **File** — уникальный файл по `hash` (MD5). Поля: hash, name, size, mime_type.
- **FilePath** — путь к файлу в источнике. Уникален по (`path`, `source_id`). Поля: path, source_id, hash, is_active, deleted_at.
- **SourceProcessing** — статус обработки источника. Три фазы: scan, processing, cleanup. Каждая фаза имеет статус (`scanning`/`completed`/`failed` и т.д.) и таймстемпы.

## Соглашения по коду

- **Ошибки:** типизированные — `repository.NotFoundError`, `repository.DuplicateError`, `disk.DiskError`. Проверяй через `errors.As`.
- **Логирование:** `log/slog`, передаётся в сервис через конструктор. Не использовать глобальный логгер.
- **Контекст:** все методы, работающие с I/O, принимают `ctx context.Context` первым аргументом.
- **Пути:** хранить относительные пути (без префикса `disk:` и без пути источника). Использовать `stripPrefix`.
- **Комментарии:** экспортируемые функции/типы — с doc-комментарием на английском.

## Тестирование

- **Unit-тесты:** моки интерфейсов (`mockDiskClient`, `mockRepository`), без БД и сети.
- **Интеграционные тесты:** реальная БД + реальный Yandex Disk. Пропускаются в `-short` режиме.
- **Перед тестом:** всегда `testutils.CleanTestData(ctx, pool)` для очистки БД.
- **Конфиг тестов:** `test_config.toml` (не коммитится), шаблон — `test_config.example.toml`.
- **Запуск:**
  - Все тесты: `go test ./... -count=1`
  - Только unit: `go test ./... -short -count=1`
  - Конкретный тест: `go test ./internal/service/ -run TestName -v -count=1`
  - Через Makefile: `make test-all`, `make test-short`

## Git-соглашения

- **Формат коммитов:** `Phase X.Y: краткое описание`
- **Автор:** `iaorekhov-1980 <i.a.orekhov@gmail.com>`
- **Коммит:** `git commit --author="iaorekhov-1980 <i.a.orekhov@gmail.com>" -m "Phase X.Y: ..."`
- **Пуш:** `git push` (или `--force-with-lease` после amend)

## Правила работы агента

- Одна задача = один файл. После создания/правки файла — СТОП, ждать подтверждения.
- Не продолжать автоматически. Ждать «continue», «next», «proceed».
- Не выполнять `git commit`/`git push` без явной команды.
- Перед правкой существующего файла — прочитать его актуальное содержимое.
- Не добавлять зависимости без явного запроса.
- Не менять `test_config.toml` (локальный, с секретами).

## Рабочий паттерн (цикл разработки)

Проект ведётся итерациями. Каждая итерация — это цикл из 4 шагов между двумя ролями:
**Reasoner** (планирование) и **Agent** (реализация).

### Роли

- **Reasoner** (`deepseek-reasoner`) — режим Chat. Планирует, проектирует, обновляет состояние.
  Не пишет код. Сильные стороны: декомпозиция, trade-off'ы, анализ.
- **Agent** (`deepseek-chat`) — режим Agent. Реализует код по ТЗ. Пошагово, один файл за раз.

### Цикл

1. **Reasoner готовит `input`** на базе `docs/overview/actual_state.md` и `docs/overview/target_state.md`.
   Результат — файл `docs/input/NN_phaseX_Y_<цель>.md` (ТЗ на одну под-фазу).

2. **Agent реализует `input`** и готовит `output`.
   Результат — файл `docs/output/phaseX_Y_<цель>_result.md` (что сделано, итоги).

3. **Reasoner обновляет состояние** на базе `output`:
   - `docs/overview/actual_state.md` — что теперь реализовано
   - `docs/overview/target_state.md` — что осталось

4. **Возврат к шагу 1** для следующей под-фазы.

### Правила для Reasoner

- Всегда читай `actual_state.md` и `target_state.md` перед планированием.
- Разбивай работу на **под-фазы** (X.1, X.2, ...) — одна под-фаза = одна сессия агента.
- ТЗ (`input`) должно быть самодостаточным: контекст, что сделать, файлы, алгоритм, тесты, проверка.
- Учитывай расхождения между старым ТЗ и текущим кодом (см. `actual_state.md`).
- Не пиши код — только ТЗ и обновление состояния.

### Правила для Agent

- Работай строго по `input`-файлу текущей под-фазы.
- Одна задача = один файл. После каждого файла — СТОП, ждать подтверждения.
- В конце сессии подготовь `output`-файл с итогами (что сделано, какие файлы, результаты тестов).
- Не выходи за рамки ТЗ без явного запроса.

### Структура документации

```
docs/
  overview/
    actual_state.md     # текущее состояние (обновляет Reasoner)
    target_state.md     # что осталось (обновляет Reasoner)
  input/                # ТЗ на под-фазы (создаёт Reasoner)
  output/               # результаты сессий (создаёт Agent)
  README.md             # индекс
```
