package model

import "time"

type ProductCategory struct {
	ID        int64     `json:"id"`
	Slug      string    `json:"slug"`
	NameIDN   string    `json:"name_idn"`
	NameEN    string    `json:"name_en"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID             int64     `json:"id"`
	CategoryID     int64     `json:"category_id"`
	Slug           string    `json:"slug"`
	NameIDN        string    `json:"name_idn"`
	NameEN         string    `json:"name_en"`
	DescriptionIDN string    `json:"description_idn"`
	DescriptionEN  string    `json:"description_en"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ProductImage struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	FileURL   string    `json:"file_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductExcellence struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	TitleIDN  string    `json:"title_idn"`
	TitleEN   string    `json:"title_en"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductSpecification struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	NameIDN   string    `json:"name_idn"`
	NameEN    string    `json:"name_en"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
