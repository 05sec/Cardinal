package logger

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestService_GetLogs(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/logs", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_Panel(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/panel", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_Panel2(t *testing.T) {
	// Test general router
	//var backJSON = struct {
	//	Error int    `json:"error"`
	//	Msg   string `json:"msg"`
	//	Data  string `json:"data"`
	//}{}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/", nil)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/base", nil)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	//err := json.Unmarshal(w.Body.Bytes(), &backJSON)
	//assert.Equal(t, nil, err)
	//assert.Equal(t, backJSON.Data, "HCTF")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/time", nil)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/404_not_found_router", nil)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// no auth
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/flags", nil)
	req.Header.Set("Authorization", "error_token")
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/flags", nil)
	req.Header.Set("Authorization", "")
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/team/rank", nil)
	req.Header.Set("Authorization", "error_token")
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/team/rank", nil)
	req.Header.Set("Authorization", "")
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
}
