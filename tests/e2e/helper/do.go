package helper

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Do sends one request through the router. body may be nil, a string or []byte (sent as is),
// or any other value (sent as JSON). A non-empty token is sent as "Authorization: Bearer <token>".
func Do(t *testing.T, router http.Handler, method string, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := NewRequest(t, method, path, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return Serve(router, req)
}

// DoWithCookies is Do without a token, sending cookies instead (e.g. the refresh cookie).
func DoWithCookies(t *testing.T, router http.Handler, method string, path string, body any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	req := NewRequest(t, method, path, body)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return Serve(router, req)
}

// NewRequest builds a request with body encoded as Do describes.
func NewRequest(t *testing.T, method string, path string, body any) *http.Request {
	t.Helper()

	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		reader = strings.NewReader(b)
	case []byte:
		reader = bytes.NewReader(b)
	default:
		reader = StructToJSON(t, b)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

// Serve runs req through the router and records the response.
func Serve(router http.Handler, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// Cookie returns the cookie named name that the response sets, or nil.
func Cookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}

	return nil
}
