package model

import "time"

// Feature đại diện cho một tính năng có thể bật/tắt
type Feature struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Enabled     bool           `json:"enabled"`
	Routes      []FeatureRoute `json:"routes,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// FeatureRoute đại diện cho một route thuộc một feature
type FeatureRoute struct {
	ID        int       `json:"id"`
	FeatureID int       `json:"feature_id"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RouteStatus trạng thái kết hợp của route (feature phải bật VÀ route phải bật)
type RouteStatus struct {
	Method         string `json:"method"`
	Path           string `json:"path"`
	FeatureName    string `json:"feature_name"`
	FeatureEnabled bool   `json:"feature_enabled"`
	RouteEnabled   bool   `json:"route_enabled"`
	Active         bool   `json:"active"` // true khi cả feature lẫn route đều bật
}
