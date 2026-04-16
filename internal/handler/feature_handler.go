package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"my-go-app/internal/service"

	"github.com/gorilla/mux"
)

type FeatureHandler struct {
	svc *service.FeatureService
}

func NewFeatureHandler(svc *service.FeatureService) *FeatureHandler {
	return &FeatureHandler{svc: svc}
}

// respond helper
func respond(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// respondError helper
func respondError(w http.ResponseWriter, code int, msg string) {
	respond(w, code, map[string]string{"error": msg})
}

// GET /admin/features — Danh sách tất cả features
func (h *FeatureHandler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	features, err := h.svc.ListFeatures()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"data":  features,
		"total": len(features),
	})
}

// GET /admin/features/{id} — Chi tiết một feature
func (h *FeatureHandler) GetFeature(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}
	f, err := h.svc.GetFeature(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Feature không tồn tại")
		return
	}
	respond(w, http.StatusOK, f)
}

// POST /admin/features — Tạo feature mới
func (h *FeatureHandler) CreateFeature(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		respondError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ, cần có 'name'")
		return
	}
	f, err := h.svc.CreateFeature(body.Name, body.Description, body.Enabled)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusCreated, f)
}

// PATCH /admin/features/{id}/toggle — Bật/tắt feature
func (h *FeatureHandler) ToggleFeature(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	if err := h.svc.ToggleFeature(id, body.Enabled); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"message": "Cập nhật thành công",
		"enabled": body.Enabled,
	})
}

// POST /admin/features/{id}/routes — Thêm route vào feature
func (h *FeatureHandler) AddRoute(w http.ResponseWriter, r *http.Request) {
	featureID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Feature ID không hợp lệ")
		return
	}
	var body struct {
		Method  string `json:"method"`
		Path    string `json:"path"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Method == "" || body.Path == "" {
		respondError(w, http.StatusBadRequest, "Cần có 'method' và 'path'")
		return
	}
	rt, err := h.svc.AddRoute(featureID, body.Method, body.Path, body.Enabled)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respond(w, http.StatusCreated, rt)
}

// PATCH /admin/routes/{id}/toggle — Bật/tắt một route
func (h *FeatureHandler) ToggleRoute(w http.ResponseWriter, r *http.Request) {
	routeID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Route ID không hợp lệ")
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	if err := h.svc.ToggleRoute(routeID, body.Enabled); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respond(w, http.StatusOK, map[string]any{
		"message": "Cập nhật route thành công",
		"enabled": body.Enabled,
	})
}

// GET /admin/check?method=GET&path=/api/users — Kiểm tra trạng thái route
func (h *FeatureHandler) CheckRoute(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Query().Get("method")
	path := r.URL.Query().Get("path")
	if method == "" || path == "" {
		respondError(w, http.StatusBadRequest, "Cần có query 'method' và 'path'")
		return
	}
	rs, err := h.svc.CheckRoute(method, path)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respond(w, http.StatusOK, rs)
}

// POST /admin/sync — Đồng bộ cache từ DB
func (h *FeatureHandler) SyncCache(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.SyncCacheFromDB(); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, map[string]string{"message": "Đồng bộ cache thành công"})
}
