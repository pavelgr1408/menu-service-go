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
