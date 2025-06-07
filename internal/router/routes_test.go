package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 建立路由
	r := BuildRouter()

	// 測試案例表
	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
	}{
		{
			name:       "Attestation Options",
			method:     http.MethodPost,
			target:     "/attestation/options",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Assertion Result",
			method:     http.MethodPost,
			target:     "/assertion/result",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tc.method, tc.target, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("狀態碼錯誤：got=%d want=%d", w.Code, tc.wantStatus)
			}
		})
	}
}