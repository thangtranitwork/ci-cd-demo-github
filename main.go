package main

import (
	"log"
	"net/http"

	"my-go-app/internal/cache"
	"my-go-app/internal/db"
	"my-go-app/internal/handler"
	"my-go-app/internal/middleware"
	"my-go-app/internal/repository"
	"my-go-app/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	// ─── Khởi tạo kết nối ────────────────────────────────────────────────────
	db.InitMySQL()
	cache.InitRedis()

	// ─── Dependency Injection ─────────────────────────────────────────────────
	featureRepo := repository.NewFeatureRepository(db.DB)
	featureSvc := service.NewFeatureService(featureRepo)
	featureHandler := handler.NewFeatureHandler(featureSvc)

	// Sync cache từ DB khi khởi động
	if err := featureSvc.SyncCacheFromDB(); err != nil {
		log.Printf("⚠️ Không thể sync cache: %v", err)
	}

	// ─── Router ───────────────────────────────────────────────────────────────
	r := mux.NewRouter()

	// Middleware toàn cục
	r.Use(loggingMiddleware)

	// ── Admin API (quản lý feature flags) ────────────────────────────────────
	admin := r.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/features", featureHandler.ListFeatures).Methods(http.MethodGet)
	admin.HandleFunc("/features", featureHandler.CreateFeature).Methods(http.MethodPost)
	admin.HandleFunc("/features/{id}", featureHandler.GetFeature).Methods(http.MethodGet)
	admin.HandleFunc("/features/{id}/toggle", featureHandler.ToggleFeature).Methods(http.MethodPatch)
	admin.HandleFunc("/features/{id}/routes", featureHandler.AddRoute).Methods(http.MethodPost)
	admin.HandleFunc("/routes/{id}/toggle", featureHandler.ToggleRoute).Methods(http.MethodPatch)
	admin.HandleFunc("/check", featureHandler.CheckRoute).Methods(http.MethodGet)
	admin.HandleFunc("/sync", featureHandler.SyncCache).Methods(http.MethodPost)

	// ── Business Routes (được bảo vệ bởi Feature Flag middleware) ────────────
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.FeatureFlagMiddleware(featureSvc))

	// Feature: user-management
	api.HandleFunc("/users", listUsersHandler).Methods(http.MethodGet)
	api.HandleFunc("/users", createUserHandler).Methods(http.MethodPost)
	api.HandleFunc("/users/{id}", deleteUserHandler).Methods(http.MethodDelete)

	// Feature: product-catalog
	api.HandleFunc("/products", listProductsHandler).Methods(http.MethodGet)
	api.HandleFunc("/products", createProductHandler).Methods(http.MethodPost)
	api.HandleFunc("/products/{id}", getProductHandler).Methods(http.MethodGet)
	api.HandleFunc("/products/{id}", updateProductHandler).Methods(http.MethodPut)

	// Feature: beta-dashboard
	api.HandleFunc("/dashboard", dashboardHandler).Methods(http.MethodGet)
	api.HandleFunc("/dashboard/stats", dashboardStatsHandler).Methods(http.MethodGet)

	// Health check (không qua feature flag)
	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/", rootHandler).Methods(http.MethodGet)

	log.Println("🚀 Server đang chạy tại :8080")
	log.Println("📋 Admin API: http://localhost:8080/admin/features")
	log.Println("🔍 Check route: http://localhost:8080/admin/check?method=GET&path=/api/users")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// ─── Business Handlers (stub) ─────────────────────────────────────────────────

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"data":[{"id":1,"name":"Nguyễn Văn A"},{"id":2,"name":"Trần Thị B"}]}`))
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Tạo người dùng thành công","id":3}`))
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(r)["id"]
	w.Write([]byte(`{"message":"Xóa thành công","id":"` + id + `"}`))
}

func listProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"data":[{"id":1,"name":"Laptop Pro","price":25000000},{"id":2,"name":"Điện thoại X","price":15000000}]}`))
}

func createProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Tạo sản phẩm thành công","id":3}`))
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(r)["id"]
	w.Write([]byte(`{"id":"` + id + `","name":"Laptop Pro","price":25000000}`))
}

func updateProductHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Cập nhật sản phẩm thành công"}`))
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Beta Dashboard","version":"beta-1.0"}`))
}

func dashboardStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"users":1250,"products":340,"orders":89}`))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"feature-flag-demo"}`))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{
		"service": "Feature Flag Demo API",
		"version": "1.0.0",
		"endpoints": {
			"health":         "GET /health",
			"list_features":  "GET /admin/features",
			"create_feature": "POST /admin/features",
			"toggle_feature": "PATCH /admin/features/{id}/toggle",
			"add_route":      "POST /admin/features/{id}/routes",
			"toggle_route":   "PATCH /admin/routes/{id}/toggle",
			"check_route":    "GET /admin/check?method=GET&path=/api/users",
			"sync_cache":     "POST /admin/sync"
		}
	}`))
}

// ─── Logging Middleware ───────────────────────────────────────────────────────

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("→ %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
