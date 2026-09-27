# Documentation Index

Архив сессий разработки проекта `big_data_file_archive_processor`.
Каждый файл — одна сессия (промпт или результат).

## Обзор проекта

| Файл | Описание |
|---|---|
| [implementation plan.md](implementation%20plan.md) | Исходный план реализации проекта |
| [prompt_v1.md](prompt_v1.md) | Первый промпт — общее описание проекта (Yandex Disk + YDB) |
| [prompt_v2.md](prompt_v2.md) | Второй промпт — уточнённая архитектура |
| [second_architecture_prompt.md](second_architecture_prompt.md) | Второй архитектурный промпт |
| [prompt_continue_development.md](prompt_continue_development.md) | Промпт для продолжения разработки |
| [prompt_development_eng.md](prompt_development_eng.md) | Промпт на разработку (EN) |
| [prompt_development_rus.md](prompt_development_rus.md) | Промпт на разработку (RU) |
| [prompt_development_v2_eng.md](prompt_development_v2_eng.md) | Промпт на разработку v2 (EN) |

## По фазам

### Phase 3 — Database Layer

| Файл | Тип | Описание |
|---|---|---|
| [prompt_for_phase_3.md](prompt_for_phase_3.md) | Промпт | ТЗ на реализацию слоя БД |
| [result_of_phase_3.md](result_of_phase_3.md) | Результат | Итоги Phase 3 |
| [result_of_phase_3_4.md](result_of_phase_3_4.md) | Результат | Итоги Phase 3.4 (интеграционные тесты БД) |

### Phase 4 — Yandex Disk Client

| Файл | Тип | Описание |
|---|---|---|
| [prompt_for_phase_4.md](prompt_for_phase_4.md) | Промпт | ТЗ на реализацию DiskClient |
| [prompt_for_phase_4_2.md](prompt_for_phase_4_2.md) | Промпт | ТЗ на Phase 4.2 |

### Phase 5 — Service Layer

| Файл | Тип | Описание |
|---|---|---|
| [prompt_for_phase_5.md](prompt_for_phase_5.md) | Промпт | ТЗ на реализацию сервисного слоя (ScanService, CleanupService, TxManager, Retry) |

## Хронология

1. **Обзор и планирование** — `prompt_v1`, `prompt_v2`, `second_architecture_prompt`, `implementation plan`
2. **Phase 3** — слой БД (PostgreSQL, репозитории, миграции, тесты)
3. **Phase 4** — клиент Yandex Disk API
4. **Phase 5** — сервисный слой (сканер, очистка, транзакции, retry)
