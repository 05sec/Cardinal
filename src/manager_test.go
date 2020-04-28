package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestService_ManagerLogout(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/logout", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_ManagerLogin(t *testing.T) {
	w := httptest.NewRecorder()
	// JSON error
	req, _ := http.NewRequest("POST", "/api/manager/login", bytes.NewBuffer([]byte(`Name=e99&Password=qwe1qwe2qwe3`)))
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// Login fail
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name":     "e99",
		"Password": "123456",
	})
	req, _ = http.NewRequest("POST", "/api/manager/login", bytes.NewBuffer(jsonData))
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	// Login success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     "e99",
		"Password": "qwe1qwe2qwe3",
	})
	req, _ = http.NewRequest("POST", "/api/manager/login", bytes.NewBuffer(jsonData))
	service.Router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var backData struct {
		Error string `json:"error"`
		Msg   string `json:"msg"`
		Data  string `json:"data"`
	}

	_ = json.Unmarshal(w.Body.Bytes(), &backData)
	managerToken = backData.Data
}

func TestService_GetAllManager(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/managers", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_NewManager(t *testing.T) {
	w := httptest.NewRecorder()
	// JSON error
	req, _ := http.NewRequest("POST", "/api/manager/manager", bytes.NewBuffer([]byte(`Name=e991111&Password=1122334455`)))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// repeat manager
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name":     "e99",
		"Password": "123456",
	})
	req, _ = http.NewRequest("POST", "/api/manager/manager", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// Wrong check account
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     "e99pppp",
		"Password": "",
		"IsCheck":  false,
	})
	req, _ = http.NewRequest("POST", "/api/manager/manager", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     "admin",
		"Password": "123456",
	})
	req, _ = http.NewRequest("POST", "/api/manager/manager", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     "check",
		"Password": "useless_password",
		"IsCheck":  true,
	})
	req, _ = http.NewRequest("POST", "/api/manager/manager", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_RefreshManagerToken(t *testing.T) {
	// error id
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/manager/token?id=asdfg", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// no id
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/manager/token", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// id not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/manager/token?id=233", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	// success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/manager/token?id=2", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/manager/token?id=3", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	var backJSON = struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  string `json:"data"`
	}{}
	err := json.Unmarshal(w.Body.Bytes(), &backJSON)
	assert.Equal(t, nil, err)
	assert.Equal(t, 200, w.Code)
	checkToken = backJSON.Data
}

func TestService_ChangeManagerPassword(t *testing.T) {
	// error id
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/manager/changePassword?id=asdfg", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// no id
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/manager/changePassword", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// id not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/manager/changePassword?id=233", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	// success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/manager/changePassword?id=2", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_DeleteManager(t *testing.T) {
	// error id
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/manager/manager?id=asdfg", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// no id exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/manager", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// id not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/manager?id=233", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	// success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/manager?id=2", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_ManagerRequired(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/teams", nil)
	req.Header.Set("Authorization", checkToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}
