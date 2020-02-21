package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"io"
	"os"
	"strings"
)

func (s *Service) makeErrJSON(httpStatusCode int, errCode int, msg interface{}) (int, interface{}) {
	return httpStatusCode, gin.H{"error": errCode, "msg": fmt.Sprint(msg)}
}

func (s *Service) makeSuccessJSON(data interface{}) (int, interface{}) {
	return 200, gin.H{"error": 0, "msg": "success", "data": data}
}

func (s *Service) checkPassword(inputPassword string, realPassword string) bool {
	// sha1( sha1(password) + salt )
	return s.addSalt(inputPassword) == realPassword
}

func (s *Service) addSalt(input string) string {
	return s.sha1Encode(s.sha1Encode(input) + s.Conf.Base.Salt)
}

func (s *Service) generateToken() string {
	return uuid.NewV4().String()
}

func (s *Service) sha1Encode(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func (s *Service) hmacSha1Encode(input string, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	_, _ = io.WriteString(h, input)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// FileSize returns the formatter text of the giving size.
func (s *Service) FileSize(size int64) string {
	return humanize.IBytes(uint64(size))
}

// IsExist check the file or folder existed.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// InputString used in the install.go for the config file guide.
func InputString(str *string, hint string) {
	var err error
	var input string
	for input == "" {
		fmt.Println(">", hint)

		stdin := bufio.NewReader(os.Stdin)
		input, err = stdin.ReadString('\n')
		input = strings.Trim(input, "\n")
		if err != nil || input == ""{
			if *str != "" {
				break
			}
		}
		*str = input
	}
}
