package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/hrfee/jfa-go/logger"
)

func newAuthTestApp() *appContext {
	l := logger.NewEmptyLogger()
	return &appContext{LoggerSet: LoggerSet{info: l, debug: l, err: l}}
}

func newAuthTestContext(method, target, authHeader string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(method, target, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	ctx.Request = req
	return ctx, recorder
}

func assertNoPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	f()
}

func TestDecodeValidateLoginHeaderRejectsMalformedAuthorization(t *testing.T) {
	app := newAuthTestApp()
	cases := []string{
		"",
		"Basic",
		"Bearer token",
		"Basic not-base64",
		"Basic " + base64NoColon("username"),
	}

	for _, header := range cases {
		t.Run(header, func(t *testing.T) {
			ctx, recorder := newAuthTestContext(http.MethodGet, "/", header)
			assertNoPanic(t, func() {
				_, _, ok := app.decodeValidateLoginHeader(ctx, false)
				if ok {
					t.Fatal("expected malformed authorization to be rejected")
				}
			})
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", recorder.Code)
			}
		})
	}
}

func TestDecodeValidateAuthHeaderRejectsMalformedBearerWithoutPanic(t *testing.T) {
	app := newAuthTestApp()
	cases := []string{
		"",
		"Bearer",
		"Basic dXNlcjpwYXNz",
		"Bearer not-a-jwt",
	}

	for _, header := range cases {
		t.Run(header, func(t *testing.T) {
			ctx, recorder := newAuthTestContext(http.MethodGet, "/", header)
			assertNoPanic(t, func() {
				_, ok := app.decodeValidateAuthHeader(ctx)
				if ok {
					t.Fatal("expected malformed authorization to be rejected")
				}
			})
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", recorder.Code)
			}
		})
	}
}

func TestDecodeValidateRefreshCookieRejectsMalformedTokenWithoutPanic(t *testing.T) {
	app := newAuthTestApp()
	ctx, recorder := newAuthTestContext(http.MethodGet, "/", "")
	ctx.Request.AddCookie(&http.Cookie{Name: "refresh", Value: "not-a-jwt"})

	assertNoPanic(t, func() {
		_, ok := app.decodeValidateRefreshCookie(ctx, "refresh")
		if ok {
			t.Fatal("expected malformed refresh cookie to be rejected")
		}
	})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestLogoutUserClearsUserRefreshCookie(t *testing.T) {
	app := newAuthTestApp()
	ctx, recorder := newAuthTestContext(http.MethodPost, "https://accounts.example/my/logout", "")
	ctx.Request.AddCookie(&http.Cookie{Name: "user-refresh", Value: "old-refresh"})

	app.LogoutUser(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	setCookie := strings.Join(recorder.Result().Header.Values("Set-Cookie"), "\n")
	if !strings.Contains(setCookie, "user-refresh=") {
		t.Fatalf("expected user-refresh cookie to be cleared, got %q", setCookie)
	}
	if strings.Contains(setCookie, "refresh=") && !strings.Contains(setCookie, "user-refresh=") {
		t.Fatalf("cleared wrong refresh cookie: %q", setCookie)
	}
}

func TestDecodeValidateAuthHeaderRejectsMissingClaimsWithoutPanic(t *testing.T) {
	t.Setenv("JFA_SECRET", "secret")
	app := newAuthTestApp()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp":  time.Now().Add(time.Hour).Unix(),
		"type": "bearer",
	})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, recorder := newAuthTestContext(http.MethodGet, "/", "Bearer "+signed)

	assertNoPanic(t, func() {
		_, ok := app.decodeValidateAuthHeader(ctx)
		if ok {
			t.Fatal("expected token with missing claims to be rejected")
		}
	})
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func base64NoColon(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}
