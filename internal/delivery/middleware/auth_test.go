package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/google/uuid"
)

type fakeAuthenticator struct {
	account *domain.Account
}

func (f fakeAuthenticator) Authenticate(context.Context, string) (*domain.Account, error) {
	return f.account, nil
}

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name       string
		isAdmin    bool
		wantStatus int
	}{
		{name: "admin is allowed", isAdmin: true, wantStatus: http.StatusNoContent},
		{name: "regular account is forbidden", isAdmin: false, wantStatus: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			protected := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			handler := Authenticate(
				fakeAuthenticator{account: &domain.Account{ID: uuid.New(), IsAdmin: test.isAdmin}},
				"gooffer_session",
				slog.Default(),
			)(protected)
			request := httptest.NewRequest(http.MethodGet, "/api/admin/card-definitions", nil)
			request.AddCookie(&http.Cookie{Name: "gooffer_session", Value: "token"})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}
