# Documentation Index

Архив сессий разработки проекта `big_data_file_archive_processor`.

## Структура

- **`input/`** — входные материалы сессий (ТЗ, промпты, планы). С них начинается сессия.
- **`output/`** — результаты сессий (итоги, отчёты). Ими сессия завершается.

Имя файла отражает основную цель сессии.

---

## Input (ТЗ и промпты)

### Обзор и планирование

| Файл | Цель сессии |
|---|---|
| [01_project_overview_yandex_disk_ydb.md](input/01_project_overview_yandex_disk_ydb.md) | Общее описание проекта: Yandex Disk + YDB, дедупликация, витрина |
| [02_ods_module_postgres_abstraction.md](input/02_ods_module_postgres_abstraction.md) | ODS-модуль с абстракцией БД (PostgreSQL) |
| [03_ods_module_ydb_eng.md](input/03_ods_module_ydb_eng.md) | ODS-модуль на YDB (EN) |
| [04_ods_module_ydb_rus.md](input/04_ods_module_ydb_rus.md) | ODS-модуль на YDB (RU) |
| [05_ods_module_postgres_eng_v2.md](input/05_ods_module_postgres_eng_v2.md) | ODS-модуль с PostgreSQL (EN, v2) |
| [06_dds_and_datamart_design.md](input/06_dds_and_datamart_design.md) | Проектирование слоёв DDS и Data Mart |
| [07_architecture_review_post_phase4.md](input/07_architecture_review_post_phase4.md) | Архитектурное ревью после Phase 4 |
| [08_implementation_plan.md](input/08_implementation_plan.md) | План реализации проекта по фазам |

### По фазам

| Файл | Цель сессии |
|---|---|
| [09_phase3_database_layer.md](input/09_phase3_database_layer.md) | Phase 3 — слой БД (PostgreSQL, репозитории, миграции) |
| [10_phase4_disk_client.md](input/10_phase4_disk_client.md) | Phase 4 — клиент Yandex Disk API |
| [11_phase4_2_disk_client_continue.md](input/11_phase4_2_disk_client_continue.md) | Phase 4.2 — продолжение клиента Disk API |
| [12_phase5_service_layer.md](input/12_phase5_service_layer.md) | Phase 5 — сервисный слой (ScanService, CleanupService, TxManager, Retry) |
| [13_phase5_1_cleanup_service.md](input/13_phase5_1_cleanup_service.md) | Phase 5.1 — CleanupService (дедупликация путей) + тесты + CLI |

---

## Output (результаты)

| Файл | Результат сессии |
|---|---|
| [phase3_database_layer_result.md](output/phase3_database_layer_result.md) | Итоги Phase 3 — слой БД реализован и протестирован |
| [phase3_4_database_tests_result.md](output/phase3_4_database_tests_result.md) | Итоги Phase 3.4 — интеграционные тесты БД и инфраструктура |

---

## Хронология

1. **Обзор и планирование** — `01`–`08`
2. **Phase 3** — слой БД (PostgreSQL, репозитории, миграции, тесты)
3. **Phase 4** — клиент Yandex Disk API
4. **Phase 5** — сервисный слой (сканер, очистка, транзакции, retry)
