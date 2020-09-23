package cardinal_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getWebHook(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/webhooks", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_newWebHook(t *testing.T) {
	// empty payload
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/manager/webhook", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error payload
	w = httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"URL":  123123123,
		"Type": 123123123,
	})
	req, _ = http.NewRequest("POST", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// missing param
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"URL":   "https://cardinal.ink",
		"Token": "123123123123123",
	})
	req, _ = http.NewRequest("POST", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// type error
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"URL":  "https://cardinal.ink",
		"Type": "asdadasdasda",
	})
	req, _ = http.NewRequest("POST", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"URL":  "https://cardinal.ink",
		"Type": "any",
	})
	req, _ = http.NewRequest("POST", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"URL":     "https://cardinal.ink",
		"Type":    "any",
		"Retry":   5,
		"Timeout": 10,
	})
	req, _ = http.NewRequest("POST", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_editWebHook(t *testing.T) {
	// empty payload
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/manager/webhook", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error payload
	w = httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"URL":  123123123,
		"Type": 123123123,
	})
	req, _ = http.NewRequest("PUT", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// missing param
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":    1,
		"URL":   "https://cardinal.ink",
		"Token": "123123123123123",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// missing id
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"URL":   "https://cardinal.ink",
		"Type":  "any",
		"Token": "123123123123123",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// type error
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   1,
		"URL":  "https://cardinal.ink",
		"Type": "asdadasdasda",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   233,
		"URL":  "https://cardinal.ink/aaaa",
		"Type": "any",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   1,
		"URL":  "https://cardinal.ink/aaaa",
		"Type": "any",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":      1,
		"URL":     "https://cardinal.ink",
		"Type":    "any",
		"Retry":   5,
		"Timeout": 10,
	})
	req, _ = http.NewRequest("PUT", "/api/manager/webhook", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_deleteWebHook(t *testing.T) {
	// missing param
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/manager/webhook", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// param type error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/webhook?id=aaa", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// not found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/webhook?id=2333", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	// param type error
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/webhook?id=1", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
