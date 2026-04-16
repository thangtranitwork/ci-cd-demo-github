package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"my-go-app/internal/cache"
	"my-go-app/internal/model"
	"my-go-app/internal/repository"

	"github.com/redis/go-redis/v9"
)

type FeatureService struct {
	repo *repository.FeatureRepository
}

func NewFeatureService(repo *repository.FeatureRepository) *FeatureService {
	return &FeatureService{repo: repo}
}

// ListFeatures lấy danh sách tất cả features (ưu tiên Redis cache)
func (s *FeatureService) ListFeatures() ([]model.Feature, error) {
	var features []model.Feature
	if err := cache.GetJSON(cache.AllFeatures, &features); err == nil {
		return features, nil
	}

	features, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}

	// Lưu vào cache
	if err := cache.SetJSON(cache.AllFeatures, features, 5*time.Minute); err != nil {
		log.Printf("⚠️ Không thể lưu cache AllFeatures: %v", err)
	}
	return features, nil
}

// GetFeature lấy feature theo ID
func (s *FeatureService) GetFeature(id int) (*model.Feature, error) {
	key := fmt.Sprintf("feature:id:%d", id)
	var feature model.Feature
	if err := cache.GetJSON(key, &feature); err == nil {
		return &feature, nil
	}

	f, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	cache.SetJSON(key, f, 5*time.Minute)
	return f, nil
}

// CreateFeature tạo feature mới và xóa cache
func (s *FeatureService) CreateFeature(name, description string, enabled bool) (*model.Feature, error) {
	f, err := s.repo.Create(name, description, enabled)
	if err != nil {
		return nil, err
	}
	s.invalidateListCache()
	return f, nil
}

// ToggleFeature bật/tắt toàn bộ feature
func (s *FeatureService) ToggleFeature(id int, enabled bool) error {
	f, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("feature không tồn tại")
	}

	if err := s.repo.SetFeatureEnabled(id, enabled); err != nil {
		return err
	}

	// Xóa cache của feature và tất cả routes liên quan
	s.invalidateFeatureCache(id)
	for _, rt := range f.Routes {
		cache.Delete(cache.RouteKey(rt.Method, rt.Path))
	}
	return nil
}

// ToggleRoute bật/tắt một route cụ thể
func (s *FeatureService) ToggleRoute(routeID int, enabled bool) error {
	rt, err := s.repo.GetRouteByID(routeID)
	if err != nil {
		return errors.New("route không tồn tại")
	}

	if err := s.repo.SetRouteEnabled(routeID, enabled); err != nil {
		return err
	}

	// Xóa cache của route và feature cha
	cache.Delete(cache.RouteKey(rt.Method, rt.Path))
	s.invalidateFeatureCache(rt.FeatureID)
	return nil
}

// AddRoute thêm route vào feature
func (s *FeatureService) AddRoute(featureID int, method, path string, enabled bool) (*model.FeatureRoute, error) {
	if _, err := s.repo.GetByID(featureID); err != nil {
		return nil, errors.New("feature không tồn tại")
	}

	rt, err := s.repo.AddRoute(featureID, method, path, enabled)
	if err != nil {
		return nil, err
	}

	s.invalidateFeatureCache(featureID)
	return rt, nil
}

// CheckRoute kiểm tra route có active không (ưu tiên Redis cache)
func (s *FeatureService) CheckRoute(method, path string) (*model.RouteStatus, error) {
	key := cache.RouteKey(method, path)
	var rs model.RouteStatus
	if err := cache.GetJSON(key, &rs); err == nil {
		return &rs, nil
	}
	if err := cache.GetJSON(key, &rs); err != nil && err != redis.Nil {
		log.Printf("⚠️ Redis error: %v", err)
	}

	rs2, err := s.repo.IsRouteActive(method, path)
	if err != nil {
		return nil, err
	}

	cache.SetJSON(key, rs2, 5*time.Minute)
	return rs2, nil
}

// SyncCacheFromDB đồng bộ toàn bộ cache từ MySQL
func (s *FeatureService) SyncCacheFromDB() error {
	features, err := s.repo.ListAll()
	if err != nil {
		return err
	}

	// Lưu danh sách tổng
	if err := cache.SetJSON(cache.AllFeatures, features, 5*time.Minute); err != nil {
		return err
	}

	// Lưu từng feature
	for _, f := range features {
		key := fmt.Sprintf("feature:id:%d", f.ID)
		cache.SetJSON(key, f, 5*time.Minute)
		cache.SetJSON(cache.FeatureKey(f.Name), f, 5*time.Minute)

		// Lưu từng route
		for _, rt := range f.Routes {
			rs := model.RouteStatus{
				Method:         rt.Method,
				Path:           rt.Path,
				FeatureName:    f.Name,
				FeatureEnabled: f.Enabled,
				RouteEnabled:   rt.Enabled,
				Active:         f.Enabled && rt.Enabled,
			}
			cache.SetJSON(cache.RouteKey(rt.Method, rt.Path), rs, 5*time.Minute)
		}
	}

	log.Printf("✅ Đã sync %d features vào Redis", len(features))
	return nil
}

func (s *FeatureService) invalidateListCache() {
	cache.Delete(cache.AllFeatures)
}

func (s *FeatureService) invalidateFeatureCache(id int) {
	cache.Delete(fmt.Sprintf("feature:id:%d", id))
	cache.Delete(cache.AllFeatures)
}
