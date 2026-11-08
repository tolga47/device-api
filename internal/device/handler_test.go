package device

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store, err := NewStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })

	r := gin.New()
	r.HandleMethodNotAllowed = true
	NewHandler(store).Register(r)
	return r
}

func doRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeDevice(t *testing.T, w *httptest.ResponseRecorder) Device {
	t.Helper()
	var d Device
	if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil {
		t.Fatalf("bad response body %q: %v", w.Body.String(), err)
	}
	return d
}

func TestCreateDevice(t *testing.T) {
	r := setupRouter(t)

	w := doRequest(r, "POST", "/devices", `{"name":"iPhone 15","brand":"Apple"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/devices/1" {
		t.Errorf("expected Location /devices/1, got %q", loc)
	}

	d := decodeDevice(t, w)
	if d.ID != 1 || d.Name != "iPhone 15" || d.Brand != "Apple" {
		t.Errorf("unexpected device: %+v", d)
	}
	if d.CreatedAt.IsZero() {
		t.Error("created_at should be set")
	}
}

func TestCreateDeviceValidation(t *testing.T) {
	r := setupRouter(t)

	bodies := []string{
		`{"name":"iPhone"}`,
		`{"brand":"Apple"}`,
		`{"name":"   ","brand":"Apple"}`,
		`not json`,
	}
	for _, body := range bodies {
		w := doRequest(r, "POST", "/devices", body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("body %s: expected 400, got %d", body, w.Code)
		}
	}
}

func TestCreateDeviceTrimsSpaces(t *testing.T) {
	r := setupRouter(t)

	w := doRequest(r, "POST", "/devices", `{"name":"  iPhone 15 ","brand":" Apple"}`)
	d := decodeDevice(t, w)
	if d.Name != "iPhone 15" || d.Brand != "Apple" {
		t.Errorf("expected trimmed values, got %+v", d)
	}
}

func TestGetDevice(t *testing.T) {
	r := setupRouter(t)
	doRequest(r, "POST", "/devices", `{"name":"iPhone 15","brand":"Apple"}`)

	w := doRequest(r, "GET", "/devices/1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if d := decodeDevice(t, w); d.Name != "iPhone 15" {
		t.Errorf("unexpected device: %+v", d)
	}

	if w := doRequest(r, "GET", "/devices/42", ""); w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	if w := doRequest(r, "GET", "/devices/abc", ""); w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListDevices(t *testing.T) {
	r := setupRouter(t)
	doRequest(r, "POST", "/devices", `{"name":"iPhone 15","brand":"Apple"}`)
	doRequest(r, "POST", "/devices", `{"name":"Galaxy S24","brand":"Samsung"}`)
	doRequest(r, "POST", "/devices", `{"name":"MacBook Air","brand":"Apple"}`)

	tests := []struct {
		path string
		want int
	}{
		{"/devices", 3},
		{"/devices?brand=Apple", 2},
		{"/devices?brand=apple", 2},
		{"/devices?brand=%20Apple%20", 2},
		{"/devices?brand=Nokia", 0},
	}

	for _, tt := range tests {
		w := doRequest(r, "GET", tt.path, "")
		if w.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", tt.path, w.Code)
		}
		var devices []Device
		if err := json.Unmarshal(w.Body.Bytes(), &devices); err != nil {
			t.Fatal(err)
		}
		if len(devices) != tt.want {
			t.Errorf("%s: expected %d devices, got %d", tt.path, tt.want, len(devices))
		}
	}
}

func TestUpdateDevice(t *testing.T) {
	r := setupRouter(t)
	created := decodeDevice(t, doRequest(r, "POST", "/devices", `{"name":"iPhone 15","brand":"Apple"}`))

	w := doRequest(r, "PUT", "/devices/1", `{"name":"Pixel 9","brand":"Google"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	d := decodeDevice(t, w)
	if d.Name != "Pixel 9" || d.Brand != "Google" {
		t.Errorf("unexpected device: %+v", d)
	}
	if !d.CreatedAt.Equal(created.CreatedAt) {
		t.Error("created_at should not change on update")
	}

	// PUT needs both fields
	if w := doRequest(r, "PUT", "/devices/1", `{"name":"Pixel 9"}`); w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	if w := doRequest(r, "PUT", "/devices/42", `{"name":"a","brand":"b"}`); w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestPatchDevice(t *testing.T) {
	r := setupRouter(t)
	doRequest(r, "POST", "/devices", `{"name":"iPhone 15","brand":"Apple"}`)

	w := doRequest(r, "PATCH", "/devices/1", `{"name":"iPhone 16"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	d := decodeDevice(t, w)
	if d.Name != "iPhone 16" || d.Brand != "Apple" {
		t.Errorf("only name should change, got %+v", d)
	}

	if w := doRequest(r, "PATCH", "/devices/1", `{"brand":""}`); w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	if w := doRequest(r, "PATCH", "/devices/1", `{"name":"  "}`); w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for blank name, got %d", w.Code)
	}
	if w := doRequest(r, "PATCH", "/devices/42", `{"name":"x"}`); w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	r := setupRouter(t)

	if w := doRequest(r, "DELETE", "/devices", ""); w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
	if w := doRequest(r, "POST", "/devices/1", `{}`); w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestDeleteDevice(t *testing.T) {
	r := setupRouter(t)
	doRequest(r, "POST", "/devices", `{"name":"iPhone 15","brand":"Apple"}`)

	if w := doRequest(r, "DELETE", "/devices/1", ""); w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if w := doRequest(r, "GET", "/devices/1", ""); w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
	if w := doRequest(r, "DELETE", "/devices/1", ""); w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
