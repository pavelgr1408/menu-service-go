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
