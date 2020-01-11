package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

type Utils struct {}

func (s *Service) makeErrJSON(httpStatusCode int, errCode int, msg interface{}) (int, interface{}) {
	return httpStatusCode, gin.H{"error": errCode, "msg": fmt.Sprint(msg)}
}

func (s *Service) makeSuccessJSON(data interface{}) (int, interface{}) {
	return 200, gin.H{"error": 0, "msg": "success", "data": data}
}

func (s *Service) checkPassword(inputPassword string, realPassword string) bool{
	// sha1( sha1(password) + salt )
	return s.sha1Encode(s.sha1Encode(inputPassword) + s.Conf.Base.Salt) == realPassword
}

func (s *Service) addSalt(input string) string{
	return s.sha1Encode(s.sha1Encode(input) + s.Conf.Base.Salt)
}

func (s *Service) generateToken() string{
	return uuid.NewV4().String()
}

func (s *Service) sha1Encode(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}