package repository

import (
	"database/sql"
	"fmt"

	"my-go-app/internal/model"
)

type FeatureRepository struct {
	db *sql.DB
}

func NewFeatureRepository(db *sql.DB) *FeatureRepository {
	return &FeatureRepository{db: db}
}

// ListAll trả về tất cả features kèm routes
func (r *FeatureRepository) ListAll() ([]model.Feature, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, enabled, created_at, updated_at
		FROM features ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []model.Feature
	for rows.Next() {
		var f model.Feature
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.Enabled, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		routes, err := r.getRoutesByFeatureID(f.ID)
		if err != nil {
			return nil, err
		}
		f.Routes = routes
		features = append(features, f)
	}
	return features, nil
}

// GetByID trả về feature theo ID kèm routes
func (r *FeatureRepository) GetByID(id int) (*model.Feature, error) {
	var f model.Feature
	err := r.db.QueryRow(`
		SELECT id, name, description, enabled, created_at, updated_at
		FROM features WHERE id = ?
	`, id).Scan(&f.ID, &f.Name, &f.Description, &f.Enabled, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	routes, err := r.getRoutesByFeatureID(f.ID)
	if err != nil {
		return nil, err
	}
	f.Routes = routes
	return &f, nil
}

// GetByName trả về feature theo tên
func (r *FeatureRepository) GetByName(name string) (*model.Feature, error) {
	var f model.Feature
	err := r.db.QueryRow(`
		SELECT id, name, description, enabled, created_at, updated_at
		FROM features WHERE name = ?
	`, name).Scan(&f.ID, &f.Name, &f.Description, &f.Enabled, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	routes, err := r.getRoutesByFeatureID(f.ID)
	if err != nil {
		return nil, err
	}
	f.Routes = routes
	return &f, nil
}

// Create tạo feature mới
func (r *FeatureRepository) Create(name, description string, enabled bool) (*model.Feature, error) {
	res, err := r.db.Exec(
		"INSERT INTO features (name, description, enabled) VALUES (?, ?, ?)",
		name, description, enabled,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetByID(int(id))
}

// SetFeatureEnabled bật/tắt toàn bộ feature (và cache)
func (r *FeatureRepository) SetFeatureEnabled(id int, enabled bool) error {
	_, err := r.db.Exec("UPDATE features SET enabled = ? WHERE id = ?", enabled, id)
	return err
}

// SetRouteEnabled bật/tắt một route cụ thể
func (r *FeatureRepository) SetRouteEnabled(routeID int, enabled bool) error {
	_, err := r.db.Exec("UPDATE feature_routes SET enabled = ? WHERE id = ?", enabled, routeID)
	return err
}

// AddRoute thêm route vào feature
func (r *FeatureRepository) AddRoute(featureID int, method, path string, enabled bool) (*model.FeatureRoute, error) {
	res, err := r.db.Exec(
		"INSERT INTO feature_routes (feature_id, method, path, enabled) VALUES (?, ?, ?, ?)",
		featureID, method, path, enabled,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetRouteByID(int(id))
}

// GetRouteByID trả về route theo ID
func (r *FeatureRepository) GetRouteByID(id int) (*model.FeatureRoute, error) {
	var rt model.FeatureRoute
	err := r.db.QueryRow(`
		SELECT id, feature_id, method, path, enabled, created_at, updated_at
		FROM feature_routes WHERE id = ?
	`, id).Scan(&rt.ID, &rt.FeatureID, &rt.Method, &rt.Path, &rt.Enabled, &rt.CreatedAt, &rt.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// IsRouteActive kiểm tra xem route có đang hoạt động không (cả feature lẫn route phải bật)
func (r *FeatureRepository) IsRouteActive(method, path string) (*model.RouteStatus, error) {
	var rs model.RouteStatus
	err := r.db.QueryRow(`
		SELECT fr.method, fr.path, f.name, f.enabled, fr.enabled
		FROM feature_routes fr
		JOIN features f ON f.id = fr.feature_id
		WHERE fr.method = ? AND fr.path = ?
	`, method, path).Scan(&rs.Method, &rs.Path, &rs.FeatureName, &rs.FeatureEnabled, &rs.RouteEnabled)
	if err != nil {
		return nil, fmt.Errorf("route %s %s không tồn tại", method, path)
	}
	rs.Active = rs.FeatureEnabled && rs.RouteEnabled
	return &rs, nil
}

// getRoutesByFeatureID helper nội bộ
func (r *FeatureRepository) getRoutesByFeatureID(featureID int) ([]model.FeatureRoute, error) {
	rows, err := r.db.Query(`
		SELECT id, feature_id, method, path, enabled, created_at, updated_at
		FROM feature_routes WHERE feature_id = ? ORDER BY id
	`, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []model.FeatureRoute
	for rows.Next() {
		var rt model.FeatureRoute
		if err := rows.Scan(&rt.ID, &rt.FeatureID, &rt.Method, &rt.Path, &rt.Enabled, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, rt)
	}
	return routes, nil
}
