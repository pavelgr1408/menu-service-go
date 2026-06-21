# restaurant-menu-service

MVP микросервис меню для учебной микросервисной платформы онлайн-ресторана.

Сервис хранит категории и позиции меню в PostgreSQL, отдаёт REST API в JSON и запускается локально через Docker Compose. Регистрацию, login, refresh-token, выпуск JWT и управление пользователями сервис не реализует: эти задачи остаются в `restaurant-auth-service` или API Gateway.

## Технологии

- Go 1.24 в Dockerfile;
- стандартный `net/http` роутинг;
- PostgreSQL 18;
- `pgx/v5 v5.7.4`;
- встроенные SQL-миграции через `embed.FS`;
- структурированное JSON-логирование через `log/slog`;
- Docker Compose;
- distroless runtime image `gcr.io/distroless/static-debian12:nonroot`.

## Порты

Так как порт `18081` уже используется auth-service в локальном окружении, для menu-service выбрана следующая свободная пара из правила совместимости.

| Компонент | Внутренний порт | Порт на host |
|---|---:|---:|
| `restaurant-menu-service` | `8080` | `18082` |
| `restaurant-menu-postgres` | `5432` | `15434` |

## Быстрый запуск

```bash
docker compose up --build
```

После запуска сервис доступен по адресу:

```text
http://localhost:18082
```

Остановка:

```bash
docker compose down
```

Остановка с удалением volume PostgreSQL:

```bash
docker compose down -v
```

## Makefile

```bash
make test          # запустить Go-тесты
make build         # собрать бинарник в bin/menu-service
make run           # запустить сервис локально без Docker
make compose-up    # docker compose up --build
make compose-down  # docker compose down
make compose-logs  # смотреть логи контейнеров
make clean         # удалить bin/
```

Для локального запуска без Docker Compose нужна доступная PostgreSQL БД и переменные окружения из `.env.example`.

## Переменные окружения

| Переменная | Значение по умолчанию | Назначение |
|---|---|---|
| `HTTP_PORT` | `8080` | Порт HTTP-сервера внутри контейнера |
| `DATABASE_HOST` | `localhost` | Хост PostgreSQL |
| `DATABASE_PORT` | `15434` | Порт PostgreSQL при локальном запуске без Docker Compose |
| `DATABASE_NAME` | `restaurant_menu_db` | Имя БД |
| `DATABASE_USER` | `restaurant_menu_user` | Пользователь БД |
| `DATABASE_PASSWORD` | `restaurant_menu_password` | Пароль БД |
| `DATABASE_SSLMODE` | `disable` | SSL mode для PostgreSQL |
| `DATABASE_URL` | пусто | Если задан, имеет приоритет над раздельными DATABASE-переменными |
| `MENU_SERVICE_PORT` | `18082` | Внешний порт сервиса в Docker Compose |
| `MENU_POSTGRES_PORT` | `15434` | Внешний порт PostgreSQL в Docker Compose |

## API

| Метод | Путь | Авторизация | Назначение |
|---|---|---|---|
| `GET` | `/health` | не нужна | Проверка, что сервис запущен |
| `GET` | `/api/v1/menu` | `Authorization: Bearer <token>` | Категории и позиции выбранной категории |
| `GET` | `/api/v1/menu?categoryId=burgers` | `Authorization: Bearer <token>` | Категории и позиции конкретной категории |
| `GET` | `/api/v1/menu/items/{itemId}` | `Authorization: Bearer <token>` | Детальная карточка позиции меню |

Для MVP сервис проверяет только наличие непустого заголовка `Authorization` формата `Bearer <token>`. Подпись JWT в этом сервисе не валидируется.

## Примеры curl

### Health

```bash
curl -X GET "http://localhost:18082/health"
```

Ответ:

```json
{
  "service": "restaurant-menu-service",
  "status": "OK"
}
```

### Меню первой категории

```bash
curl -X GET "http://localhost:18082/api/v1/menu" \
  -H "Authorization: Bearer test-token" \
  -H "Accept: application/json"
```

Пример ответа:

```json
{
  "selectedCategoryId": "burgers",
  "categories": [
    {
      "id": "burgers",
      "name": "Бургеры"
    },
    {
      "id": "chicken",
      "name": "Курица"
    },
    {
      "id": "potatoes-and-snacks",
      "name": "Картофель и закуски"
    },
    {
      "id": "drinks",
      "name": "Напитки"
    },
    {
      "id": "desserts",
      "name": "Десерты"
    },
    {
      "id": "sauces",
      "name": "Соусы"
    }
  ],
  "items": [
    {
      "id": "big-burger",
      "categoryId": "burgers",
      "name": "Большой бургер",
      "shortDescription": "Бургер с говяжьей котлетой, сыром, салатом и фирменным соусом",
      "imageUrl": "https://cdn.example.com/menu/items/big-burger.png",
      "price": {
        "amount": 29900,
        "currency": "RUB",
        "formatted": "299 ₽"
      },
      "status": "AVAILABLE"
    }
  ]
}
```

В списочном методе нет `fullDescription`, `nutrition`, `display_order`, `sortOrder`, `pagination` и других запрещённых клиентских полей.

### Меню конкретной категории

```bash
curl -X GET "http://localhost:18082/api/v1/menu?categoryId=burgers" \
  -H "Authorization: Bearer test-token" \
  -H "Accept: application/json"
```

### Детальная карточка позиции меню

```bash
curl -X GET "http://localhost:18082/api/v1/menu/items/big-burger" \
  -H "Authorization: Bearer test-token" \
  -H "Accept: application/json"
```

Пример ответа:

```json
{
  "id": "big-burger",
  "categoryId": "burgers",
  "name": "Большой бургер",
  "shortDescription": "Бургер с говяжьей котлетой, сыром, салатом и фирменным соусом",
  "fullDescription": "Большой бургер с говяжьей котлетой, сыром, свежим салатом, маринованными огурцами и фирменным соусом в булочке с кунжутом.",
  "imageUrl": "https://cdn.example.com/menu/items/big-burger.png",
  "price": {
    "amount": 29900,
    "currency": "RUB",
    "formatted": "299 ₽"
  },
  "status": "AVAILABLE",
  "nutrition": {
    "caloriesKcal": 540,
    "proteinsG": 27.5,
    "fatsG": 29.1,
    "carbohydratesG": 42.4
  }
}
```

### Ошибка авторизации

```bash
curl -X GET "http://localhost:18082/api/v1/menu" \
  -H "Accept: application/json"
```

Ответ `401`:

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Требуется заголовок Authorization: Bearer <token>"
  }
}
```

### Неизвестная категория

```bash
curl -X GET "http://localhost:18082/api/v1/menu?categoryId=unknown" \
  -H "Authorization: Bearer test-token" \
  -H "Accept: application/json"
```

Ответ `404`:

```json
{
  "error": {
    "code": "CATEGORY_NOT_FOUND",
    "message": "Категория меню не найдена"
  }
}
```

### Неизвестная позиция меню

```bash
curl -X GET "http://localhost:18082/api/v1/menu/items/unknown" \
  -H "Authorization: Bearer test-token" \
  -H "Accept: application/json"
```

Ответ `404`:

```json
{
  "error": {
    "code": "MENU_ITEM_NOT_FOUND",
    "message": "Позиция меню не найдена"
  }
}
```

## Формат ошибок

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Описание ошибки"
  }
}
```

Минимальные коды ошибок:

| HTTP | code | Когда возникает |
|---:|---|---|
| `400` | `INVALID_REQUEST` | Некорректный запрос |
| `401` | `UNAUTHORIZED` | Нет `Authorization: Bearer <token>` или токен пустой |
| `404` | `CATEGORY_NOT_FOUND` | Передан неизвестный `categoryId` |
| `404` | `MENU_ITEM_NOT_FOUND` | Позиция меню не найдена |
| `500` | `INTERNAL_ERROR` | Неожиданная внутренняя ошибка |

## Схема БД

### `menu_categories`

| Поле | Тип | Описание |
|---|---|---|
| `id` | `TEXT PRIMARY KEY` | Публичный идентификатор категории |
| `name` | `TEXT NOT NULL` | Название категории |
| `display_order` | `INTEGER NOT NULL` | Внутренний порядок сортировки |
| `created_at` | `TIMESTAMPTZ NOT NULL DEFAULT NOW()` | Дата создания |
| `updated_at` | `TIMESTAMPTZ NOT NULL DEFAULT NOW()` | Дата обновления |

### `menu_items`

| Поле | Тип | Описание |
|---|---|---|
| `id` | `TEXT PRIMARY KEY` | Публичный идентификатор позиции |
| `category_id` | `TEXT NOT NULL REFERENCES menu_categories(id)` | Категория |
| `name` | `TEXT NOT NULL` | Название |
| `short_description` | `TEXT NOT NULL` | Краткое описание для списка |
| `full_description` | `TEXT NOT NULL` | Полное описание для детальной карточки |
| `image_url` | `TEXT NOT NULL` | URL изображения |
| `price_amount` | `INTEGER NOT NULL` | Цена в минимальных денежных единицах, для RUB — копейки |
| `price_currency` | `TEXT NOT NULL` | Валюта, для seed-данных `RUB` |
| `price_formatted` | `TEXT NOT NULL` | Готовая строка цены для клиента |
| `status` | `TEXT NOT NULL` | `AVAILABLE` или `UNAVAILABLE` |
| `calories_kcal` | `INTEGER NOT NULL` | Калории |
| `proteins_g` | `NUMERIC(6,2) NOT NULL` | Белки, г |
| `fats_g` | `NUMERIC(6,2) NOT NULL` | Жиры, г |
| `carbohydrates_g` | `NUMERIC(6,2) NOT NULL` | Углеводы, г |
| `display_order` | `INTEGER NOT NULL` | Внутренний порядок сортировки |
| `created_at` | `TIMESTAMPTZ NOT NULL DEFAULT NOW()` | Дата создания |
| `updated_at` | `TIMESTAMPTZ NOT NULL DEFAULT NOW()` | Дата обновления |

`display_order` используется только в SQL для стабильной сортировки и не отдаётся в JSON API.

## Seed-данные

Категории:

1. `burgers` — Бургеры
2. `chicken` — Курица
3. `potatoes-and-snacks` — Картофель и закуски
4. `drinks` — Напитки
5. `desserts` — Десерты
6. `sauces` — Соусы

Позиции:

- `big-burger`
- `cheese-burger`
- `double-burger`
- `chicken-burger`
- `nuggets`
- `french-fries`
- `cola`
- `orange-juice`
- `ice-cream`
- `apple-pie`
- `cheese-sauce`
- `barbecue-sauce`

## Архитектура проекта

```text
cmd/menu-service              # entrypoint: config, db, migrations, service, HTTP server, graceful shutdown
internal/config               # ENV-конфигурация
internal/domain               # DTO, доменные ошибки, коды ошибок
internal/httpapi              # маршруты, handlers, auth middleware, error mapping, request logging
internal/service              # бизнес-сценарии меню
internal/store                # PostgreSQL repository на pgx/v5
migrations                    # embed-миграции и seed-данные
```

## Проверка после запуска

```bash
curl -s "http://localhost:18082/health"
curl -s "http://localhost:18082/api/v1/menu" -H "Authorization: Bearer test-token"
curl -s "http://localhost:18082/api/v1/menu?categoryId=burgers" -H "Authorization: Bearer test-token"
curl -s "http://localhost:18082/api/v1/menu/items/big-burger" -H "Authorization: Bearer test-token"
curl -i "http://localhost:18082/api/v1/menu"
curl -i "http://localhost:18082/api/v1/menu?categoryId=unknown" -H "Authorization: Bearer test-token"
curl -i "http://localhost:18082/api/v1/menu/items/unknown" -H "Authorization: Bearer test-token"
```
