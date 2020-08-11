package utils

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/vidar-team/Cardinal/conf"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"strings"
	"time"
)

// MakeErrJSON makes the error response JSON for gin.
func MakeErrJSON(httpStatusCode int, errCode int, msg interface{}) (int, interface{}) {
	return httpStatusCode, gin.H{"error": errCode, "msg": fmt.Sprint(msg)}
}

// MakeErrJSON makes the successful response JSON for gin.
func MakeSuccessJSON(data interface{}) (int, interface{}) {
	return 200, gin.H{"error": 0, "msg": "success", "data": data}
}

// CheckPassword: Add salt and check the password.
func CheckPassword(inputPassword string, realPassword string) bool {
	// sha1( sha1(password) + salt )
	return HmacSha1Encode(inputPassword, conf.Get().Salt) == realPassword
}

// Sha1Encode: Sha1 encode input string.
func Sha1Encode(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

// AddSalt: Use the config salt as key to HmacSha1Encode.
func AddSalt(input string) string {
	return HmacSha1Encode(input, conf.Get().Salt)
}

// HmacSha1Encode: HMAC SHA1 encode
func HmacSha1Encode(input string, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	_, _ = io.WriteString(h, input)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// GenerateToken: return UUID v4 string.
func GenerateToken() string {
	return uuid.NewV4().String()
}

// FileSize returns the formatter text of the giving size.
func FileSize(size int64) string {
	return humanize.IBytes(uint64(size))
}

// FileIsExist check the file or folder existed.
func FileIsExist(path string) bool {
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
		input = strings.Trim(input, "\r\n")
		if err != nil || input == "" {
			if *str != "" {
				break
			}
		}
		*str = input
	}
}

func SSHExecute(ip string, port string, user string, password string, command string) (string, error) {
	client, err := ssh.Dial("tcp", ip+":"+port, &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var output bytes.Buffer
	session.Stdout = &output
	err = session.Run(command)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
