package cardinal_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/game"
	"github.com/vidar-team/Cardinal/internal/healthy"
	"github.com/vidar-team/Cardinal/internal/timer"
)

// three team accounts
// vidar (change name to Vidar, login)
// e99 (login)
// John	(delete)

// Team Test
func Test_NewTeams(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name": "vidar",
		"Logo": "",
	})
	req, _ := http.NewRequest("POST", "/api/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error payload
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal([]map[string]interface{}{{
		"Logo": "",
	}})
	req, _ = http.NewRequest("POST", "/api/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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
	req, _ = http.NewRequest("POST", "/api/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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
	req, _ = http.NewRequest("POST", "/api/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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
		Name      string `json:"Name"`
		Password  string `json:"Password"`
		Token     string `json:"token"`
		AccessKey string `json:"access_key"`
	}{Name: password.Data[0].Name, Password: password.Data[0].Password, Token: ""},
		struct {
			Name      string `json:"Name"`
			Password  string `json:"Password"`
			Token     string `json:"token"`
			AccessKey string `json:"access_key"`
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
	req, _ = http.NewRequest("POST", "/api/manager/teams", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func Test_GetAllTeams(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/teams", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_EditTeam(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name": "vidar",
		"Logo": "",
	})
	req, _ := http.NewRequest("PUT", "/api/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// team not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   233,
		"Name": "vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// team name repeat
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   2,
		"Name": "vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":   1,
		"Name": "Vidar",
		"Logo": "",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/team", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_ResetTeamPassword(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"IDd": 3,
	})
	req, _ := http.NewRequest("POST", "/api/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// team not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID": 233,
	})
	req, _ = http.NewRequest("POST", "/api/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID": 3,
	})
	req, _ = http.NewRequest("POST", "/api/manager/team/resetPassword", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_DeleteTeam(t *testing.T) {
	// error id
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/manager/team?id=asdfg", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// id not exist
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/team?id=233", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	// success
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/manager/team?id=3", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_TeamLogin(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name":     123123,
		"Password": "",
	})
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error password
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     team[1].Name,
		"Password": "aaa",
	})
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	// success Vidar
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     team[0].Name,
		"Password": team[0].Password,
	})
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var backJSON = struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  string `json:"data"`
	}{}
	err := json.Unmarshal(w.Body.Bytes(), &backJSON)
	assert.Equal(t, nil, err)
	team[0].Token = backJSON.Data

	// success e99
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"Name":     team[1].Name,
		"Password": team[1].Password,
	})
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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

func Test_GetTeamInfo(t *testing.T) {
	// Team1 Vidar
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/team/info", nil)
	req.Header.Set("Authorization", team[0].Token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var backJSON = struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  struct {
			Name  string
			Logo  string
			Score float64
			Rank  int
			Token string
		} `json:"data"`
	}{}
	err := json.Unmarshal(w.Body.Bytes(), &backJSON)
	assert.Equal(t, nil, err)
	// save access key for test
	team[0].AccessKey = backJSON.Data.Token

	// Team2 e99
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/team/info", nil)
	req.Header.Set("Authorization", team[1].Token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	backJSON = struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  struct {
			Name  string
			Logo  string
			Score float64
			Rank  int
			Token string
		} `json:"data"`
	}{}
	err = json.Unmarshal(w.Body.Bytes(), &backJSON)
	assert.Equal(t, nil, err)
	// save access key for test
	team[1].AccessKey = backJSON.Data.Token
}

func Test_TeamLogout(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/logout", nil)
	req.Header.Set("Authorization", team[0].Token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	//login again
	w = httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"Name":     team[0].Name,
		"Password": team[0].Password,
	})
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var backJSON = struct {
		Error int    `json:"error"`
		Msg   string `json:"msg"`
		Data  string `json:"data"`
	}{}
	err := json.Unmarshal(w.Body.Bytes(), &backJSON)
	assert.Equal(t, nil, err)
	team[0].Token = backJSON.Data
}

// Gamebox Test
func Test_NewGameBoxes(t *testing.T) {
	// error payload
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"ChallengeID": 1,
		"TeamID":      1,
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	})
	req, _ := http.NewRequest("POST", "/api/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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
	req, _ = http.NewRequest("POST", "/api/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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
	req, _ = http.NewRequest("POST", "/api/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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
	req, _ = http.NewRequest("POST", "/api/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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
	req, _ = http.NewRequest("POST", "/api/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
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
	req, _ = http.NewRequest("POST", "/api/manager/gameboxes", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// Set gamebox visible
	// set challenge id 1
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":      1,
		"Visible": true,
	})
	req, _ = http.NewRequest("POST", "/api/manager/challenge/visible", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// set challenge id 3
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":      3,
		"Visible": true,
	})
	req, _ = http.NewRequest("POST", "/api/manager/challenge/visible", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_EditGameBox(t *testing.T) {
	// payload error
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"ID":          "1",
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	})
	req, _ := http.NewRequest("PUT", "/api/manager/gamebox", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// gamebox not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":          233,
		"IP":          "172.0.0.1",
		"Port":        "1234",
		"Description": "web1 for Vidar",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/gamebox", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 500, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"ID":          1,
		"IP":          "172.0.0.1",
		"Port":        "12345",
		"Description": "Web1 for Vidar",
	})
	req, _ = http.NewRequest("PUT", "/api/manager/gamebox", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_GetGameBoxes(t *testing.T) {
	// error query
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/gameboxes", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/gameboxes?page=asda&per=skfdnj", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/gameboxes?page=0&per=1", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/gameboxes?page=1&per=0", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/gameboxes?page=1&per=1", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_GetSelfGameBoxes(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/team/gameboxes", nil)
	req.Header.Set("Authorization", team[0].Token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

// Flag Test
func Test_GenerateFlag(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/manager/flag/generate", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var count int
	db.MySQL.Model(&db.Flag{}).Count(&count)
	assert.NotEqual(t, 0, count)
}

func Test_GetFlags(t *testing.T) {
	// error query
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/flags", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/flags?page=asda&per=skfdnj", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/flags?page=0&per=1", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/flags?page=1&per=0", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/manager/flags?page=1&per=1", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

// Vidar -> e99 web1	flag1
// Vidar -> e99 pwn1	flag2
// e99 -> Vidar pwn1	flag3
func Test_SubmitFlag(t *testing.T) {
	timer.Get().NowRound = 1

	var flag1 db.Flag
	db.MySQL.Model(&db.Flag{}).Where(&db.Flag{
		TeamID:      2,
		ChallengeID: 1,
		Round:       1,
	}).Find(&flag1)
	fmt.Println(flag1)
	assert.NotEqual(t, flag1.Flag, "")

	var flag2 db.Flag
	db.MySQL.Model(&db.Flag{}).Where(&db.Flag{
		TeamID:      2,
		ChallengeID: 3,
		Round:       1,
	}).Find(&flag2)
	fmt.Println(flag2)
	assert.NotEqual(t, flag2.Flag, "")

	var flag3 db.Flag
	db.MySQL.Model(&db.Flag{}).Where(&db.Flag{
		TeamID:      1,
		ChallengeID: 3,
		Round:       1,
	}).Find(&flag3)
	fmt.Println(flag3)
	assert.NotEqual(t, flag3.Flag, "")

	// not begin
	timer.Get().Status = "wait"
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"flag": flag1.Flag,
	})
	req, _ := http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", team[0].AccessKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	timer.Get().Status = "on"

	// empty token
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"flag": flag1.Flag,
	})
	req, _ = http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "")
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	// error token
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"flag": flag1.Flag,
	})
	req, _ = http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "errortoken")
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	// error payload
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"flag": 12312312,
	})
	req, _ = http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", team[0].AccessKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// error flag
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]string{
		"flag": "hctf{here is a error flag}",
	})
	req, _ = http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", team[0].AccessKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	// success flag1
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]string{
		"flag": flag1.Flag,
	})
	req, _ = http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", team[0].AccessKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	fmt.Println("flag1", w.Body.String())

	// success flag2
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]string{
		"flag": flag2.Flag,
	})
	req, _ = http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", team[0].AccessKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	fmt.Println("flag2", w.Body.String())

	// success flag3
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]string{
		"flag": flag3.Flag,
	})
	req, _ = http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", team[1].AccessKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	fmt.Println("flag3", w.Body.String())

	// repeat submit
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]string{
		"flag": flag1.Flag,
	})
	req, _ = http.NewRequest("POST", "/api/flag", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", team[0].AccessKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
}

// e99 pwn1 ID:4
func Test_CheckDown(t *testing.T) {
	// not begin
	timer.Get().Status = "wait"
	w := httptest.NewRecorder()
	jsonData, _ := json.Marshal(map[string]interface{}{
		"GameBoxID": 4,
	})
	req, _ := http.NewRequest("POST", "/api/manager/checkDown", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	timer.Get().Status = "on"
	// payload error
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"GameBoxID": "4",
	})
	req, _ = http.NewRequest("POST", "/api/manager/checkDown", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// gamebox not found
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"GameBoxID": 233,
	})
	req, _ = http.NewRequest("POST", "/api/manager/checkDown", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	// success
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"GameBoxID": 4,
	})
	req, _ = http.NewRequest("POST", "/api/manager/checkDown", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	fmt.Println("checkdown", w.Body.String())

	// repeat
	w = httptest.NewRecorder()
	jsonData, _ = json.Marshal(map[string]interface{}{
		"GameBoxID": 4,
	})
	req, _ = http.NewRequest("POST", "/api/manager/checkDown", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
}

func Test_CalculateRoundScore(t *testing.T) {

	game.CalculateRoundScore(1)
	// Check team score
	var vidar db.Team
	db.MySQL.Model(&db.Team{}).Where(&db.Team{Model: gorm.Model{ID: 1}}).Find(&vidar)
	var e99 db.Team
	db.MySQL.Model(&db.Team{}).Where(&db.Team{Model: gorm.Model{ID: 2}}).Find(&e99)
	assert.Equal(t, 2020.0, vidar.Score)
	assert.Equal(t, 1980.0, e99.Score)

	// Check gamebox score
	var gameboxes []db.GameBox
	db.MySQL.Model(&db.GameBox{}).Order("`id` ASC").Find(&gameboxes)
	assert.Equal(t, 1010.0, gameboxes[0].Score)
	assert.Equal(t, 990.0, gameboxes[1].Score)
	assert.Equal(t, 1010.0, gameboxes[2].Score)
	assert.Equal(t, 990.0, gameboxes[3].Score)
}

// Healthy check
func Test_PreviousRoundScore(t *testing.T) {
	assert.Equal(t, healthy.PreviousRoundScore(), float64(0))
}

func Test_TotalScore(t *testing.T) {
	assert.Equal(t, healthy.TotalScore(), float64(0))
}

func Test_GetAllBulletins2(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/team/bulletins", nil)
	req.Header.Set("Authorization", team[0].Token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

// Rank
func Test_GetRankList(t *testing.T) {
	assert.Equal(t, len(game.GetRankList()), 2)
	assert.Equal(t, game.GetRankList()[0].TeamName, "Vidar")
	assert.Equal(t, game.GetRankList()[1].TeamName, "E99")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/manager/rank", nil)
	req.Header.Set("Authorization", managerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/team/rank", nil)
	req.Header.Set("Authorization", team[0].Token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_GetManagerRankList(t *testing.T) {
	assert.Equal(t, len(game.GetRankList()), 2)
	assert.Equal(t, game.GetManagerRankList()[0].TeamName, "Vidar")
	assert.Equal(t, game.GetManagerRankList()[1].TeamName, "E99")
}

func Test_GetRankListTitle(t *testing.T) {
	assert.Equal(t, len(game.GetRankList()), 2)
}
