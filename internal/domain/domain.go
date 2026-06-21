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
