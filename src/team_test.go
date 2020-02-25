package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// three team accounts
// vidar (change name to Vidar, login)
// e99 (login)
// John	(delete)

// Team Test
func TestService_NewTeams(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name": "vidar",
		"Logo": "",
	})
	req, _ := http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error payload
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Logo": "",
	}})
	req, _ = http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// repeat in form
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Name": "vidar",
		"Logo": "",
	}, {
		"Name": "vidar",
		"Logo": "test",
	}})
	req, _ = http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Name": "vidar",
		"Logo": "",
	}, {
		"Name": "E99",
		"Logo": "test_image.png",
	}, {
		"Name": "John",
		"Logo": "test_image123.png",
	},
	})
	req, _ = http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// save the team password
	var password struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  []struct {
			Name     string `json:"Name"`
			Password string `json:"Password"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &password)
	assert.Equal(t, nil, err)
	// save two teams' password
	team = append(team, struct {
		Name     string `json:"Name"`
		Password string `json:"Password"`
		Token    string `json:"token"`
	}{Name: password.Data[0].Name, Password: password.Data[0].Password, Token: ""},
		struct {
			Name     string `json:"Name"`
			Password string `json:"Password"`
			Token    string `json:"token"`
		}{Name: password.Data[1].Name, Password: password.Data[1].Password, Token: ""},
	)

	// repeat in database
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Name": "vidar",
		"Logo": "",
	}, {
		"Name": "E99",
		"Logo": "test_image.png",
	}})
	req, _ = http.NewRequest("POST", "/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestService_GetAllTeams(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/manager/teams", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_GetTeamInfo(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/manager/teams", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_EditTeam(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name": "vidar",
		"Logo": "",
	})
	req, _ := http.NewRequest("PUT", "/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// team not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   233,
		"Name": "vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// team name repeat
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   2,
		"Name": "vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   1,
		"Name": "Vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_ResetTeamPassword(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"IDd": 3,
	})
	req, _ := http.NewRequest("POST", "/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// team not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID": 233,
	})
	req, _ = http.NewRequest("POST", "/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID": 3,
	})
	req, _ = http.NewRequest("POST", "/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_DeleteTeam(t *testing.T) {
	// error id
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/manager/team?id=asdfg", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// id not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/manager/team?id=233", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/manager/team?id=3", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_TeamLogin(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name":     123123,
		"Password": "",
	})
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error password
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     team[1].Name,
		"Password": "aaa",
	})
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     team[0].Name,
		"Password": team[0].Password,
	})
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var backJSON = struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  string `json:"data"`
	}{}
	err := json.Unmarshal(w.Body.Bytes(), &backJSON)
	assert.Equal(t, nil, err)
	team[0].Token = backJSON.Data

	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     team[1].Name,
		"Password": team[1].Password,
	})
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	backJSON = struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  string `json:"data"`
	}{}
	err = json.Unmarshal(w.Body.Bytes(), &backJSON)
	assert.Equal(t, nil, err)
	team[1].Token = backJSON.Data
}

func TestService_TeamLogout(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logout", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

// Gamebox Test
func TestService_NewGameBoxes(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"ChallengeID": 1,
		"TeamID":      1,
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	})
	req, _ := http.NewRequest("POST", "/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error payload
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"ChallengeID": 1,
		"TeamID":      "1",
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	}})
	req, _ = http.NewRequest("POST", "/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// challenge not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"ChallengeID": 233,
		"TeamID":      1,
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	}})
	req, _ = http.NewRequest("POST", "/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// team not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"ChallengeID": 1,
		"TeamID":      3,
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	}})
	req, _ = http.NewRequest("POST", "/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"ChallengeID": 1,
		"TeamID":      1,
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	}, {
		"ChallengeID": 1,
		"TeamID":      2,
		"IP":          "172.0.0.2",
		"Port":        "1234",
		"Description": "web1 for E99",
	}, {
		"ChallengeID": 3,
		"TeamID":      1,
		"IP":          "192.168.0.1",
		"Port":        "2345",
		"Description": "pwn1 for Vidar",
	}, {
		"ChallengeID": 3,
		"TeamID":      2,
		"IP":          "192.168.0.2",
		"Port":        "2345",
		"Description": "pwn1 for E99",
	}})
	req, _ = http.NewRequest("POST", "/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// repeat
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"ChallengeID": 1,
		"TeamID":      1,
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	}})
	req, _ = http.NewRequest("POST", "/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestService_EditGameBox(t *testing.T) {
	// payload error
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"ID":          "1",
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	})
	req, _ := http.NewRequest("PUT", "/manager/gamebox", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// gamebox not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":          233,
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	})
	req, _ = http.NewRequest("PUT", "/manager/gamebox", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":          1,
		"IP":          "172.0.0.1",
		"Port":        "12345",
		"Description": "Web1 for Vidar",
	})
	req, _ = http.NewRequest("PUT", "/manager/gamebox", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_GetGameBoxes(t *testing.T) {
	// error query
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/manager/gameboxes", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/manager/gameboxes?page=asda&per=skfdnj", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/manager/gameboxes?page=0&per=1", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/manager/gameboxes?page=1&per=0", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/manager/gameboxes?page=1&per=1", nil)
	req.Header.Set("Authorization", managerToken)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestService_GetSelfGameBoxes(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/team/gameboxes", nil)
	fmt.Println(team[0].Token)
	req.Header.Set("Authorization", team[0].Token)
	service.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
