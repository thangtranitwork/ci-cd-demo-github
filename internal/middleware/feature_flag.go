package middleware

import (
	"fmt"
	"net/http"

	"my-go-app/internal/service"
)

// FeatureFlagMiddleware chặn request nếu feature hoặc route bị tắt
func FeatureFlagMiddleware(svc *service.FeatureService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := r.Method
			path := r.URL.Path

			rs, err := svc.CheckRoute(method, path)
			if err != nil {
				// Route không được quản lý bởi feature flag → cho qua
				next.ServeHTTP(w, r)
				return
			}

			if !rs.FeatureEnabled {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, `{"error":"Feature '%s' hiện đang bị tắt","active":false}`, rs.FeatureName)
				return
			}

			if !rs.RouteEnabled {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, `{"error":"Route %s %s hiện đang bị tắt","active":false}`, method, path)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
