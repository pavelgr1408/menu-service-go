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
