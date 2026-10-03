package model

import "time"

type PartnershipProject struct {
	ID             int64     `json:"id"`
	NameIDN        string    `json:"name_idn"`
	NameEN         string    `json:"name_en"`
	DescriptionIDN string    `json:"description_idn"`
	DescriptionEN  string    `json:"description_en"`
	Image          *string   `json:"image"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PartnershipProjectPartner struct {
	ID                   int64     `json:"id"`
	PartnershipProjectID int64     `json:"partnership_project_id"`
	Logo                 string    `json:"logo"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
