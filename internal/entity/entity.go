package entity

import (
	"encoding/json"
	"time"
)

type User struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type BannerInfo struct {
	BannerID  int             `json:"banner_id"`
	TagIDs    []int           `json:"tag_ids"`
	FeatureID int             `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type BannerCreate struct {
	TagIDs    []int           `json:"tag_ids"`
	FeatureID *int            `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
}

type BannerFilter struct {
	FeatureID *int `json:"feature_id"`
	TagID     *int `json:"tag_id"`
	Limit     *int `json:"limit"`
	Offset    *int `json:"offset"`
}

type BannerUpdate struct {
	TagIDs    []int           `json:"tag_ids"`
	FeatureID *int            `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  *bool           `json:"is_active"`
}
