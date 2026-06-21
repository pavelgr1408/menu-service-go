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
