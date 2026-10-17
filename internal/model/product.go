package model

import "time"

type Product struct {
	ID             int64                  `json:"id"`
	Category       string                 `json:"category"`
	Slug           string                 `json:"slug"`
	NameIDN        string                 `json:"name_idn"`
	NameEN         string                 `json:"name_en"`
	DescriptionIDN string                 `json:"description_idn"`
	DescriptionEN  string                 `json:"description_en"`
	Images         []ProductImage         `json:"images"`
	Excellences    []ProductExcellence    `json:"excellences"`
	Specifications []ProductSpecification `json:"specifications"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
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
	ValueIDN  string    `json:"value_idn"`
	ValueEN   string    `json:"value_en"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
