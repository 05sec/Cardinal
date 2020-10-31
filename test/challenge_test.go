package cardinal_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewChallenge(t *testing.T) {
	// JSON bind fail
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Title": "Web1",
	})
	req, _ := http.NewRequest("POST", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Title":     "Web1",
		"BaseScore": 800,
	})
	req, _ = http.NewRequest("POST", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Title":     "Pwn2",
		"BaseScore": 1000,
	})
	req, _ = http.NewRequest("POST", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Title":     "Pwn1",
		"BaseScore": 1000,
	})
	req, _ = http.NewRequest("POST", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// repeat
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Title":     "Web1",
		"BaseScore": 800,
	})
	req, _ = http.NewRequest("POST", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
}

func Test_EditChallenge(t *testing.T) {
	// JSON bind fail
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Title":     "Web1",
		"BaseScore": 1000,
	})
	req, _ := http.NewRequest("PUT", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":        233,
		"Title":     "Web1",
		"BaseScore": 1000,
	})
	req, _ = http.NewRequest("PUT", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":        1,
		"Title":     "Web233",
		"BaseScore": 800,
	})
	req, _ = http.NewRequest("PUT", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":        1,
		"Title":     "Web1",
		"BaseScore": 1000,
	})
	req, _ = http.NewRequest("PUT", "/api/manager/challenge", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_GetAllChallenges(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/challenges", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_DeleteChallenge(t *testing.T) {
	// error id
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/manager/challenge?id=asdfg", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// id not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/challenge?id=233", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success delete 2
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/challenge?id=2", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_SetVisible(t *testing.T) {
	// payload error
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"ID":      1,
		"Visible": "true",
	})
	req, _ := http.NewRequest("POST", "/api/manager/challenge/visible", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// challenge not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":      2,
		"Visible": true,
	})
	req, _ = http.NewRequest("POST", "/api/manager/challenge/visible", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}
