# Validation report

Дата генерации: 2026-06-21

## Проверка 1 — gofmt

Команда:

```bash
gofmt -l $(find . -name '*.go')
```

Результат: `OK`, неформатированных Go-файлов не найдено.

## Проверка 2 — Go unit tests без внешней загрузки pgx

Команда:

```bash
go test ./internal/domain ./internal/config ./internal/service ./internal/httpapi
```

Результат: `OK`.

Покрытые проверки:

- `/health` публичный и возвращает `restaurant-menu-service`;
- `/api/v1/menu` без `Authorization` возвращает `401 UNAUTHORIZED`;
- списочный ответ `/api/v1/menu` не отдаёт `fullDescription`, `nutrition`, `display_order`, `sortOrder`, `isDefault`, `pagination`;
- детальный ответ `/api/v1/menu/items/big-burger` отдаёт `fullDescription` и `nutrition`;
- неизвестная категория возвращает `404 CATEGORY_NOT_FOUND`;
- неизвестная позиция возвращает `404 MENU_ITEM_NOT_FOUND`.

## Проверка 3 — статическая контрактная проверка проекта

Команда: Python-скрипт проверки структуры, портов, docker-compose, маршрутов, SQL-схемы, seed-данных и отсутствия запрещённых JSON-полей в публичных DTO.

Результат: `OK`.

Проверено:

- все обязательные файлы проекта присутствуют;
- Docker Compose использует уникальные имена `restaurant-menu-service`, `restaurant-menu-postgres`, `restaurant_menu_db`, `restaurant_menu_user`, `restaurant_menu_postgres_data`, `restaurant-menu-network`;
- выбранные порты: сервис `18082:8080`, PostgreSQL `15434:5432`;
- маршруты `GET /health`, `GET /api/v1/menu`, `GET /api/v1/menu/items/{itemId}` присутствуют;
- middleware проверяет `Authorization: Bearer <token>`;
- JWT-валидация в menu-service не реализована намеренно;
- таблицы `menu_categories` и `menu_items` описаны в миграции;
- CHECK-ограничения для `AVAILABLE/UNAVAILABLE` и `price_amount >= 0` присутствуют;
- seed содержит 6 обязательных категорий и 12 обязательных позиций;
- публичные DTO не содержат `display_order`, `sortOrder`, `isDefault`, `isAvailableForCart`, `badges`, `reason`, `portion`, `menuVersion`, `pagination`, `channel`, `fulfillmentType`.

## Ограничение среды проверки

Команды `go test ./...`, `go mod tidy`, `go build ./cmd/menu-service` в текущей sandbox-среде не могут быть завершены полностью, потому что среда не имеет доступа к `proxy.golang.org` для скачивания `github.com/jackc/pgx/v5 v5.7.4`.

Ошибка среды:

```text
lookup proxy.golang.org ... connection refused
```

Это не ошибка исходного кода проекта. На локальной машине с доступом в интернет команда `go mod download` подтянет зависимости по `go.mod` и `go.sum`.
