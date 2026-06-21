# Исходный код сервиса `restaurant-menu-service` (menu-service-go)
> Полный текущий снимок доступных исходных файлов сервиса меню в одном файле.
> Назначение файла — передавать его целиком AI-моделям как контекст для анализа и внесения изменений в сервис.
> Репозиторий: `https://github.com/pavelgr1408/menu-service-go` (ветка `master`).

## Содержание файлов

| # | Файл | Назначение |
|---|---|---|
| 1 | `go.mod` | Объявление Go-модуля и зависимостей |
| 2 | `go.sum` | Контрольные суммы зависимостей |
| 3 | `cmd/menu-service/main.go` | Точка входа: конфигурация, PostgreSQL, миграции, сервис, HTTP-сервер, graceful shutdown |
| 4 | `internal/config/config.go` | Загрузка конфигурации из переменных окружения и сборка DATABASE_URL |
| 5 | `internal/domain/domain.go` | Доменная модель: DTO меню, цена, КБЖУ, статусы, коды и ошибки |
| 6 | `internal/store/postgres.go` | PostgreSQL repository на pgx/v5: категории, позиции, проверки существования |
| 7 | `internal/service/service.go` | Бизнес-логика выбора категории и получения карточки позиции меню |
| 8 | `internal/httpapi/api.go` | HTTP-слой: роуты, middleware авторизации, JSON, error mapping, request logging, recover |
| 9 | `internal/httpapi/api_test.go` | Юнит-тесты HTTP API и контрактов ответов |
| 10 | `migrations/embed.go` | Раннер встроенных SQL-миграций под advisory lock |
| 11 | `migrations/001_create_menu_schema.sql` | DDL-схема БД menu_categories/menu_items и индексы |
| 12 | `migrations/002_seed_menu_data.sql` | Seed-данные категорий и позиций меню |
| 13 | `Dockerfile` | Multi-stage сборка Go-бинарника и distroless runtime |
| 14 | `docker-compose.yml` | Локальное окружение PostgreSQL + menu-service |
| 15 | `Makefile` | Команды разработки и локального запуска |
| 16 | `README.md` | Описание сервиса, запуск, API, схема БД и curl-примеры |

> Примечание: README репозитория ссылается на `.env.example`, но файл `.env.example` не найден в ветке `master` при прямой проверке. Поэтому он не включён в снимок.

---

## `go.mod`

```go
module github.com/example/restaurant-menu-service

go 1.23.0

require github.com/jackc/pgx/v5 v5.7.4

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)
```

## `go.sum`

```text
github.com/jackc/pgpassfile v1.0.0 h1:/6Hmqy13Ss2zCq62VdNG8tM1wchn8zjSGOBJ6icpsIM=
github.com/jackc/pgpassfile v1.0.0/go.mod h1:CEx0iS5ambNFdcRtxPj5JhEz+xB6uRky5eyVu/W2HEg=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 h1:iCEnooe7UlwOQYpKFhBabPMi4aNAfoODPEFNiAnClxo=
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761/go.mod h1:5TJZWKEWniPve33vlWYSoGYefn3gLQRzjfDlhSJ9ZKM=
github.com/jackc/pgx/v5 v5.7.4 h1:9wKznZrhWa2QiHL+NjTSPP6yjl3451BX3imWDnokYlg=
github.com/jackc/pgx/v5 v5.7.4/go.mod h1:ncY89UGWxg82EykZUwSpUKEfccBGGYq1xjrOpsbsfGQ=
github.com/jackc/puddle/v2 v2.2.2 h1:PR8nw+E/1w0GLuRFSmiioY6UooMp6KJv0/61nB7icHo=
github.com/jackc/puddle/v2 v2.2.2/go.mod h1:vriiEXHvEE654aYKXXjOvZM39qJ0q+azkZFrfEOc3H4=
golang.org/x/crypto v0.36.0 h1:AnAEvhDddvBdpY+uR+MyHmuZzzNqXSe/GvuDeob5L34=
golang.org/x/crypto v0.36.0/go.mod h1:Y4J0ReaxCR1IMaabaSMugxJES1EpwhBHhv2bDHklZvc=
golang.org/x/sync v0.12.0 h1:MHc5BpPuC30uJk597Ri8TV3CNZcTLu6B6z4lJy+g6Jw=
golang.org/x/sync v0.12.0/go.mod h1:1dzgHSNfp02xaA81J2MS99Qcpr2v7fw1gpm99rleRqA=
golang.org/x/text v0.23.0 h1:D71I7dUrlY+VX0gQShAThNGHFxZ13dGLBHQLVl1mJlY=
golang.org/x/text v0.23.0/go.mod h1:/BLNzu4aZCJ1+kcD0DNRotWKage4q2rGVAg4o22unh4=
```

## `cmd/menu-service/main.go`

```go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/restaurant-menu-service/internal/config"
	"github.com/example/restaurant-menu-service/internal/httpapi"
	"github.com/example/restaurant-menu-service/internal/service"
	"github.com/example/restaurant-menu-service/internal/store"
	"github.com/example/restaurant-menu-service/migrations"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration error", "error", err)
		os.Exit(1)
	}

	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := migrations.Apply(ctx, db.Pool); err != nil {
		log.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	svc := service.New(db)
	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           httpapi.New(svc, log),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("menu service started", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdown); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
}
```

## `internal/config/config.go`

```go
package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	HTTPPort        string
	DatabaseURL     string
	DatabaseHost    string
	DatabasePort    string
	DatabaseName    string
	DatabaseUser    string
	DatabasePass    string
	DatabaseSSLMode string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:        env("HTTP_PORT", "8080"),
		DatabaseHost:    env("DATABASE_HOST", "localhost"),
		DatabasePort:    env("DATABASE_PORT", "15434"),
		DatabaseName:    env("DATABASE_NAME", "restaurant_menu_db"),
		DatabaseUser:    env("DATABASE_USER", "restaurant_menu_user"),
		DatabasePass:    env("DATABASE_PASSWORD", "restaurant_menu_password"),
		DatabaseSSLMode: env("DATABASE_SSLMODE", "disable"),
	}
	if raw := os.Getenv("DATABASE_URL"); raw != "" {
		cfg.DatabaseURL = raw
		return cfg, nil
	}
	cfg.DatabaseURL = buildDatabaseURL(cfg)
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func buildDatabaseURL(cfg Config) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.DatabaseUser, cfg.DatabasePass),
		Host:   fmt.Sprintf("%s:%s", cfg.DatabaseHost, cfg.DatabasePort),
		Path:   cfg.DatabaseName,
	}
	q := u.Query()
	q.Set("sslmode", cfg.DatabaseSSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}
```

## `internal/domain/domain.go`

```go
package domain

import "errors"

type MenuItemStatus string

const (
	StatusAvailable   MenuItemStatus = "AVAILABLE"
	StatusUnavailable MenuItemStatus = "UNAVAILABLE"
)

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Price struct {
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	Formatted string `json:"formatted"`
}

type Nutrition struct {
	CaloriesKcal   int     `json:"caloriesKcal"`
	ProteinsG      float64 `json:"proteinsG"`
	FatsG          float64 `json:"fatsG"`
	CarbohydratesG float64 `json:"carbohydratesG"`
}

type MenuItemShort struct {
	ID               string         `json:"id"`
	CategoryID       string         `json:"categoryId"`
	Name             string         `json:"name"`
	ShortDescription string         `json:"shortDescription"`
	ImageURL         string         `json:"imageUrl"`
	Price            Price          `json:"price"`
	Status           MenuItemStatus `json:"status"`
}

type MenuItemDetails struct {
	ID               string         `json:"id"`
	CategoryID       string         `json:"categoryId"`
	Name             string         `json:"name"`
	ShortDescription string         `json:"shortDescription"`
	FullDescription  string         `json:"fullDescription"`
	ImageURL         string         `json:"imageUrl"`
	Price            Price          `json:"price"`
	Status           MenuItemStatus `json:"status"`
	Nutrition        Nutrition      `json:"nutrition"`
}

type MenuResponse struct {
	SelectedCategoryID string          `json:"selectedCategoryId"`
	Categories         []Category      `json:"categories"`
	Items              []MenuItemShort `json:"items"`
}

type ErrorCode string

const (
	CodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	CodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	CodeCategoryNotFound ErrorCode = "CATEGORY_NOT_FOUND"
	CodeMenuItemNotFound ErrorCode = "MENU_ITEM_NOT_FOUND"
	CodeInternalError    ErrorCode = "INTERNAL_ERROR"
)

var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrCategoryNotFound = errors.New("category not found")
	ErrMenuItemNotFound = errors.New("menu item not found")
)
```

## `internal/store/postgres.go`

```go
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/example/restaurant-menu-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct{ Pool *pgxpool.Pool }

func Open(ctx context.Context, databaseURL string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	config.MaxConns, config.MinConns, config.MaxConnLifetime = 20, 2, time.Hour
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Postgres{Pool: pool}, nil
}

func (s *Postgres) Close()                         { s.Pool.Close() }
func (s *Postgres) Ping(ctx context.Context) error { return s.Pool.Ping(ctx) }

func (s *Postgres) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, name
		FROM menu_categories
		ORDER BY display_order, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}

func (s *Postgres) CategoryExists(ctx context.Context, categoryID string) (bool, error) {
	var exists bool
	err := s.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM menu_categories WHERE id=$1)
	`, categoryID).Scan(&exists)
	return exists, err
}

func (s *Postgres) ListMenuItemsByCategory(ctx context.Context, categoryID string) ([]domain.MenuItemShort, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id,
		       category_id,
		       name,
		       short_description,
		       image_url,
		       price_amount,
		       price_currency,
		       price_formatted,
		       status
		FROM menu_items
		WHERE category_id=$1
		ORDER BY display_order, id
	`, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.MenuItemShort, 0)
	for rows.Next() {
		var item domain.MenuItemShort
		if err := rows.Scan(
			&item.ID,
			&item.CategoryID,
			&item.Name,
			&item.ShortDescription,
			&item.ImageURL,
			&item.Price.Amount,
			&item.Price.Currency,
			&item.Price.Formatted,
			&item.Status,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Postgres) MenuItemByID(ctx context.Context, itemID string) (domain.MenuItemDetails, error) {
	var item domain.MenuItemDetails
	err := s.Pool.QueryRow(ctx, `
		SELECT id,
		       category_id,
		       name,
		       short_description,
		       full_description,
		       image_url,
		       price_amount,
		       price_currency,
		       price_formatted,
		       status,
		       calories_kcal,
		       proteins_g::double precision,
		       fats_g::double precision,
		       carbohydrates_g::double precision
		FROM menu_items
		WHERE id=$1
	`, itemID).Scan(
		&item.ID,
		&item.CategoryID,
		&item.Name,
		&item.ShortDescription,
		&item.FullDescription,
		&item.ImageURL,
		&item.Price.Amount,
		&item.Price.Currency,
		&item.Price.Formatted,
		&item.Status,
		&item.Nutrition.CaloriesKcal,
		&item.Nutrition.ProteinsG,
		&item.Nutrition.FatsG,
		&item.Nutrition.CarbohydratesG,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.MenuItemDetails{}, domain.ErrMenuItemNotFound
	}
	if err != nil {
		return domain.MenuItemDetails{}, err
	}
	return item, nil
}
```

## `internal/service/service.go`

```go
package service

import (
	"context"
	"strings"

	"github.com/example/restaurant-menu-service/internal/domain"
)

type Repository interface {
	ListCategories(ctx context.Context) ([]domain.Category, error)
	CategoryExists(ctx context.Context, categoryID string) (bool, error)
	ListMenuItemsByCategory(ctx context.Context, categoryID string) ([]domain.MenuItemShort, error)
	MenuItemByID(ctx context.Context, itemID string) (domain.MenuItemDetails, error)
}

type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) GetMenu(ctx context.Context, categoryID string) (domain.MenuResponse, error) {
	categories, err := s.repo.ListCategories(ctx)
	if err != nil {
		return domain.MenuResponse{}, err
	}

	selectedCategoryID := strings.TrimSpace(categoryID)
	if selectedCategoryID == "" {
		if len(categories) > 0 {
			selectedCategoryID = categories[0].ID
		}
	} else {
		exists, err := s.repo.CategoryExists(ctx, selectedCategoryID)
		if err != nil {
			return domain.MenuResponse{}, err
		}
		if !exists {
			return domain.MenuResponse{}, domain.ErrCategoryNotFound
		}
	}

	items := make([]domain.MenuItemShort, 0)
	if selectedCategoryID != "" {
		items, err = s.repo.ListMenuItemsByCategory(ctx, selectedCategoryID)
		if err != nil {
			return domain.MenuResponse{}, err
		}
	}

	return domain.MenuResponse{
		SelectedCategoryID: selectedCategoryID,
		Categories:         categories,
		Items:              items,
	}, nil
}

func (s *Service) GetMenuItem(ctx context.Context, itemID string) (domain.MenuItemDetails, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return domain.MenuItemDetails{}, domain.ErrInvalidRequest
	}
	return s.repo.MenuItemByID(ctx, itemID)
}
```

## `internal/httpapi/api.go`

```go
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/example/restaurant-menu-service/internal/domain"
)

type MenuService interface {
	GetMenu(ctx context.Context, categoryID string) (domain.MenuResponse, error)
	GetMenuItem(ctx context.Context, itemID string) (domain.MenuItemDetails, error)
}

type API struct {
	service MenuService
	log     *slog.Logger
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    domain.ErrorCode `json:"code"`
	Message string           `json:"message"`
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func New(svc MenuService, log *slog.Logger) http.Handler {
	a := &API{service: svc, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.health)
	mux.Handle("GET /api/v1/menu", a.authenticate(http.HandlerFunc(a.menu)))
	mux.Handle("GET /api/v1/menu/items/{itemId}", a.authenticate(http.HandlerFunc(a.menuItem)))
	return a.recover(a.requestLogger(a.jsonContent(mux)))
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	a.write(w, http.StatusOK, map[string]string{"status": "OK", "service": "restaurant-menu-service"})
}

func (a *API) menu(w http.ResponseWriter, r *http.Request) {
	response, err := a.service.GetMenu(r.Context(), r.URL.Query().Get("categoryId"))
	if err != nil {
		a.handle(w, r, err)
		return
	}
	a.write(w, http.StatusOK, response)
}

func (a *API) menuItem(w http.ResponseWriter, r *http.Request) {
	itemID := strings.TrimSpace(r.PathValue("itemId"))
	item, err := a.service.GetMenuItem(r.Context(), itemID)
	if err != nil {
		a.handle(w, r, err)
		return
	}
	a.write(w, http.StatusOK, item)
}

func (a *API) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			a.problem(w, http.StatusUnauthorized, domain.CodeUnauthorized, "Требуется заголовок Authorization: Bearer <token>")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			a.problem(w, http.StatusUnauthorized, domain.CodeUnauthorized, "Требуется непустой Bearer-токен")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (a *API) problem(w http.ResponseWriter, status int, code domain.ErrorCode, message string) {
	a.write(w, status, errorEnvelope{Error: errorBody{Code: code, Message: message}})
}

func (a *API) handle(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRequest):
		a.problem(w, http.StatusBadRequest, domain.CodeInvalidRequest, "Некорректный запрос")
	case errors.Is(err, domain.ErrUnauthorized):
		a.problem(w, http.StatusUnauthorized, domain.CodeUnauthorized, "Необходима авторизация")
	case errors.Is(err, domain.ErrCategoryNotFound):
		a.problem(w, http.StatusNotFound, domain.CodeCategoryNotFound, "Категория меню не найдена")
	case errors.Is(err, domain.ErrMenuItemNotFound):
		a.problem(w, http.StatusNotFound, domain.CodeMenuItemNotFound, "Позиция меню не найдена")
	default:
		a.log.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		a.problem(w, http.StatusInternalServerError, domain.CodeInternalError, "Внутренняя ошибка сервиса")
	}
}

func (a *API) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				a.log.Error("panic", "value", value)
				a.problem(w, http.StatusInternalServerError, domain.CodeInternalError, "Внутренняя ошибка сервиса")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *API) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		a.log.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

func (a *API) jsonContent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		next.ServeHTTP(w, r)
	})
}
```

## `internal/httpapi/api_test.go`

```go
package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/restaurant-menu-service/internal/domain"
)

type fakeMenuService struct{}

func (fakeMenuService) GetMenu(ctx context.Context, categoryID string) (domain.MenuResponse, error) {
	if categoryID == "unknown" {
		return domain.MenuResponse{}, domain.ErrCategoryNotFound
	}
	return domain.MenuResponse{
		SelectedCategoryID: "burgers",
		Categories: []domain.Category{
			{ID: "burgers", Name: "Бургеры"},
			{ID: "sauces", Name: "Соусы"},
		},
		Items: []domain.MenuItemShort{{
			ID:               "big-burger",
			CategoryID:       "burgers",
			Name:             "Большой бургер",
			ShortDescription: "Бургер с говяжьей котлетой, сыром, салатом и фирменным соусом",
			ImageURL:         "https://cdn.example.com/menu/items/big-burger.png",
			Price:            domain.Price{Amount: 29900, Currency: "RUB", Formatted: "299 ₽"},
			Status:           domain.StatusAvailable,
		}},
	}, nil
}

func (fakeMenuService) GetMenuItem(ctx context.Context, itemID string) (domain.MenuItemDetails, error) {
	if itemID == "unknown" {
		return domain.MenuItemDetails{}, domain.ErrMenuItemNotFound
	}
	return domain.MenuItemDetails{
		ID:               "big-burger",
		CategoryID:       "burgers",
		Name:             "Большой бургер",
		ShortDescription: "Бургер с говяжьей котлетой, сыром, салатом и фирменным соусом",
		FullDescription:  "Большой бургер с говяжьей котлетой, сыром, свежим салатом, маринованными огурцами и фирменным соусом.",
		ImageURL:         "https://cdn.example.com/menu/items/big-burger.png",
		Price:            domain.Price{Amount: 29900, Currency: "RUB", Formatted: "299 ₽"},
		Status:           domain.StatusAvailable,
		Nutrition:        domain.Nutrition{CaloriesKcal: 540, ProteinsG: 27.5, FatsG: 29.1, CarbohydratesG: 42.4},
	}, nil
}

func testHandler() http.Handler {
	return New(fakeMenuService{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestHealthIsPublic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	testHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), `"service":"restaurant-menu-service"`) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestMenuRequiresAuthorization(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/menu", nil)
	rr := httptest.NewRecorder()
	testHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rr.Body.Bytes(), string(domain.CodeUnauthorized))
}

func TestMenuListDoesNotExposeDetailsFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/menu", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()
	testHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d, body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	for _, forbidden := range []string{"fullDescription", "nutrition", "display_order", "sortOrder", "isDefault", "pagination"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("list response exposes forbidden field %q: %s", forbidden, body)
		}
	}
}

func TestMenuItemDetailsExposeNutritionAndFullDescription(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/menu/items/big-burger", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()
	testHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d, body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	for _, required := range []string{"fullDescription", "nutrition", "caloriesKcal"} {
		if !strings.Contains(body, required) {
			t.Fatalf("details response does not contain %q: %s", required, body)
		}
	}
}

func TestCategoryNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/menu?categoryId=unknown", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()
	testHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusNotFound)
	}
	assertErrorCode(t, rr.Body.Bytes(), string(domain.CodeCategoryNotFound))
}

func TestMenuItemNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/menu/items/unknown", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rr := httptest.NewRecorder()
	testHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusNotFound)
	}
	assertErrorCode(t, rr.Body.Bytes(), string(domain.CodeMenuItemNotFound))
}

func assertErrorCode(t *testing.T, body []byte, want string) {
	t.Helper()
	var got struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode error body: %v; body=%s", err, string(body))
	}
	if got.Error.Code != want {
		t.Fatalf("error.code=%q, want %q; body=%s", got.Error.Code, want, string(body))
	}
}
```

## `migrations/embed.go`

```go
package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var files embed.FS

func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock(8142027)`); err != nil {
		return err
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock(8142027)`)

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(version BIGINT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return err
	}

	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, err := strconv.ParseInt(strings.SplitN(entry.Name(), "_", 2)[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid migration %s", entry.Name())
		}

		var applied bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}

		sql, err := files.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, version); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}
```

## `migrations/001_create_menu_schema.sql`

```sql
CREATE TABLE menu_categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    display_order INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_menu_categories_id_not_blank CHECK (length(trim(id)) > 0),
    CONSTRAINT chk_menu_categories_name_not_blank CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_menu_categories_display_order_non_negative CHECK (display_order >= 0)
);

CREATE UNIQUE INDEX ux_menu_categories_display_order ON menu_categories(display_order);

CREATE TABLE menu_items (
    id TEXT PRIMARY KEY,
    category_id TEXT NOT NULL REFERENCES menu_categories(id),
    name TEXT NOT NULL,
    short_description TEXT NOT NULL,
    full_description TEXT NOT NULL,
    image_url TEXT NOT NULL,
    price_amount INTEGER NOT NULL,
    price_currency TEXT NOT NULL,
    price_formatted TEXT NOT NULL,
    status TEXT NOT NULL,
    calories_kcal INTEGER NOT NULL,
    proteins_g NUMERIC(6,2) NOT NULL,
    fats_g NUMERIC(6,2) NOT NULL,
    carbohydrates_g NUMERIC(6,2) NOT NULL,
    display_order INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_menu_items_id_not_blank CHECK (length(trim(id)) > 0),
    CONSTRAINT chk_menu_items_name_not_blank CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_menu_items_status CHECK (status IN ('AVAILABLE','UNAVAILABLE')),
    CONSTRAINT chk_menu_items_price_amount CHECK (price_amount >= 0),
    CONSTRAINT chk_menu_items_price_currency_not_blank CHECK (length(trim(price_currency)) > 0),
    CONSTRAINT chk_menu_items_calories_non_negative CHECK (calories_kcal >= 0),
    CONSTRAINT chk_menu_items_proteins_non_negative CHECK (proteins_g >= 0),
    CONSTRAINT chk_menu_items_fats_non_negative CHECK (fats_g >= 0),
    CONSTRAINT chk_menu_items_carbohydrates_non_negative CHECK (carbohydrates_g >= 0),
    CONSTRAINT chk_menu_items_display_order_non_negative CHECK (display_order >= 0)
);

CREATE INDEX ix_menu_items_category_id_display_order ON menu_items(category_id, display_order);
CREATE UNIQUE INDEX ux_menu_items_category_display_order ON menu_items(category_id, display_order);
```

## `migrations/002_seed_menu_data.sql`

```sql
INSERT INTO menu_categories(id, name, display_order) VALUES
('burgers', 'Бургеры', 10),
('chicken', 'Курица', 20),
('potatoes-and-snacks', 'Картофель и закуски', 30),
('drinks', 'Напитки', 40),
('desserts', 'Десерты', 50),
('sauces', 'Соусы', 60);

INSERT INTO menu_items(
    id,
    category_id,
    name,
    short_description,
    full_description,
    image_url,
    price_amount,
    price_currency,
    price_formatted,
    status,
    calories_kcal,
    proteins_g,
    fats_g,
    carbohydrates_g,
    display_order
) VALUES
('big-burger', 'burgers', 'Большой бургер', 'Бургер с говяжьей котлетой, сыром, салатом и фирменным соусом', 'Большой бургер с говяжьей котлетой, сыром, свежим салатом, маринованными огурцами и фирменным соусом в булочке с кунжутом.', 'https://cdn.example.com/menu/items/big-burger.png', 29900, 'RUB', '299 ₽', 'AVAILABLE', 540, 27.50, 29.10, 42.40, 10),
('cheese-burger', 'burgers', 'Чизбургер', 'Бургер с сыром, говяжьей котлетой, луком и маринованными огурцами', 'Классический чизбургер с говяжьей котлетой, сыром, луком, маринованными огурцами, кетчупом и горчицей.', 'https://cdn.example.com/menu/items/cheese-burger.png', 15900, 'RUB', '159 ₽', 'AVAILABLE', 320, 16.30, 14.80, 31.20, 20),
('double-burger', 'burgers', 'Двойной бургер', 'Бургер с двумя говяжьими котлетами и сыром', 'Сытный бургер с двумя говяжьими котлетами, двумя ломтиками сыра, соусом и свежими овощами.', 'https://cdn.example.com/menu/items/double-burger.png', 34900, 'RUB', '349 ₽', 'UNAVAILABLE', 710, 39.70, 42.30, 44.80, 30),
('chicken-burger', 'chicken', 'Чикенбургер', 'Бургер с куриной котлетой, салатом и сливочным соусом', 'Чикенбургер с хрустящей куриной котлетой, листовым салатом, свежими овощами и сливочным соусом.', 'https://cdn.example.com/menu/items/chicken-burger.png', 21900, 'RUB', '219 ₽', 'AVAILABLE', 430, 22.20, 18.60, 39.50, 10),
('nuggets', 'chicken', 'Наггетсы', 'Куриные наггетсы в хрустящей панировке', 'Сочные куриные наггетсы в хрустящей панировке. Хорошо сочетаются с сырным или барбекю-соусом.', 'https://cdn.example.com/menu/items/nuggets.png', 18900, 'RUB', '189 ₽', 'AVAILABLE', 360, 21.00, 20.40, 23.80, 20),
('french-fries', 'potatoes-and-snacks', 'Картофель фри', 'Золотистый картофель фри с солью', 'Классический золотистый картофель фри с хрустящей корочкой и мягкой серединкой.', 'https://cdn.example.com/menu/items/french-fries.png', 11900, 'RUB', '119 ₽', 'AVAILABLE', 310, 4.20, 15.60, 39.90, 10),
('cola', 'drinks', 'Кола', 'Газированный напиток со вкусом колы', 'Охлаждённый газированный напиток со вкусом колы.', 'https://cdn.example.com/menu/items/cola.png', 9900, 'RUB', '99 ₽', 'AVAILABLE', 140, 0.00, 0.00, 35.00, 10),
('orange-juice', 'drinks', 'Апельсиновый сок', 'Апельсиновый сок без газа', 'Освежающий апельсиновый сок без газа.', 'https://cdn.example.com/menu/items/orange-juice.png', 12900, 'RUB', '129 ₽', 'AVAILABLE', 120, 1.20, 0.20, 28.60, 20),
('ice-cream', 'desserts', 'Мороженое', 'Ванильное мороженое в стаканчике', 'Нежное ванильное мороженое в удобном стаканчике.', 'https://cdn.example.com/menu/items/ice-cream.png', 10900, 'RUB', '109 ₽', 'AVAILABLE', 210, 4.10, 9.80, 27.20, 10),
('apple-pie', 'desserts', 'Яблочный пирожок', 'Горячий пирожок с яблочной начинкой', 'Хрустящий горячий пирожок с ароматной яблочной начинкой.', 'https://cdn.example.com/menu/items/apple-pie.png', 9900, 'RUB', '99 ₽', 'AVAILABLE', 260, 3.20, 12.40, 34.50, 20),
('cheese-sauce', 'sauces', 'Сырный соус', 'Порционный соус с сырным вкусом', 'Нежный порционный соус с выраженным сырным вкусом.', 'https://cdn.example.com/menu/items/cheese-sauce.png', 4900, 'RUB', '49 ₽', 'AVAILABLE', 80, 1.10, 7.20, 2.70, 10),
('barbecue-sauce', 'sauces', 'Соус барбекю', 'Порционный соус барбекю', 'Пикантный порционный соус барбекю с копчёными нотами.', 'https://cdn.example.com/menu/items/barbecue-sauce.png', 4900, 'RUB', '49 ₽', 'AVAILABLE', 60, 0.50, 0.20, 13.40, 20);
```

## `Dockerfile`

```dockerfile
FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/menu-service ./cmd/menu-service

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/menu-service /app/menu-service

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/menu-service"]
```

## `docker-compose.yml`

```yaml
name: restaurant-menu-local

services:
  restaurant-menu-postgres:
    image: postgres:18
    container_name: restaurant-menu-postgres
    environment:
      POSTGRES_DB: restaurant_menu_db
      POSTGRES_USER: restaurant_menu_user
      POSTGRES_PASSWORD: restaurant_menu_password
      PGDATA: /var/lib/postgresql/data/pgdata
    ports:
      - "${MENU_POSTGRES_PORT:-15434}:5432"
    volumes:
      - restaurant_menu_postgres_data:/var/lib/postgresql/data
    networks:
      - restaurant-menu-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U restaurant_menu_user -d restaurant_menu_db"]
      interval: 5s
      timeout: 5s
      retries: 10

  restaurant-menu-service:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: restaurant-menu-service
    environment:
      HTTP_PORT: 8080
      DATABASE_HOST: restaurant-menu-postgres
      DATABASE_PORT: 5432
      DATABASE_NAME: restaurant_menu_db
      DATABASE_USER: restaurant_menu_user
      DATABASE_PASSWORD: restaurant_menu_password
      DATABASE_SSLMODE: disable
    ports:
      - "${MENU_SERVICE_PORT:-18082}:8080"
    depends_on:
      restaurant-menu-postgres:
        condition: service_healthy
    networks:
      - restaurant-menu-network

volumes:
  restaurant_menu_postgres_data:
    name: restaurant_menu_postgres_data

networks:
  restaurant-menu-network:
    name: restaurant-menu-network
```

## `Makefile`

```makefile
.PHONY: test build run compose-up compose-down compose-logs clean

test:
	go test ./...

build:
	go build -trimpath -o bin/menu-service ./cmd/menu-service

run:
	go run ./cmd/menu-service

compose-up:
	docker compose up --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f restaurant-menu-service restaurant-menu-postgres

clean:
	rm -rf bin
```

## `README.md`

```markdown
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
```

