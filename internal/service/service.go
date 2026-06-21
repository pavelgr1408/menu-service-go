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
