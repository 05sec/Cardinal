package upload

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/utils"
)

func GetDir(c *gin.Context) (int, interface{}) {
	basePath := c.Query("path")
	folder := c.Query("folder")
	hidden, _ := strconv.ParseBool(c.Query("hidden"))
	folderOnly, _ := strconv.ParseBool(c.Query("folderOnly"))

	if basePath == "" {
		nowPath, err := os.Getwd()
		basePath = nowPath
		if err != nil {
			return utils.MakeErrJSON(500, 50025, "获取当前目录信息失败")
		}
	}
	path := filepath.Join(basePath, folder)

	f, err := os.Stat(path)
	if err != nil {
		return utils.MakeErrJSON(500, 50026, fmt.Sprintf("打开文件 %s 失败", path))
	}
	if !f.IsDir() {
		return utils.MakeErrJSON(500, 50027, fmt.Sprintf("%s 不是目录", path))
	}
	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return utils.MakeErrJSON(500, 50026, fmt.Sprintf("打开文件 %s 失败", path))
	}

	type fileItem struct {
		Name    string
		IsDir   bool
		Size    string
		ModTime int64
	}

	files := make([]fileItem, 0, len(fileInfo))
	for _, f := range fileInfo {
		if f.Name()[0] == '.' && !hidden {
			// skip hidden file.
			continue
		}

		if !f.IsDir() && folderOnly {
			// skip file.
			continue
		}

		files = append(files, fileItem{
			Name:    f.Name(),
			IsDir:   f.IsDir(),
			Size:    utils.FileSize(f.Size()),
			ModTime: f.ModTime().Unix(),
		})
	}

	return utils.MakeSuccessJSON(gin.H{
		"path":  path,
		"files": files,
	})
}

// UploadPicture is upload team logo handler for manager.
func UploadPicture(c *gin.Context) (int, interface{}) {
	file, err := c.FormFile("picture")
	if err != nil {
		return utils.MakeErrJSON(400, 40025,
			locales.I18n.T(c.GetString("lang"), "file.select_picture"),
		)
	}
	fileExt := map[string]string{
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/jpeg": ".jpg",
	}
	ext, ok := fileExt[c.GetHeader("Content-Type")]
	if !ok {
		ext = ".png"
	}
	fileName := randstr.Hex(16) + ext

	err = c.SaveUploadedFile(file, "./uploads/"+fileName)
	if err != nil {
		return utils.MakeErrJSON(500, 50014,
			locales.I18n.T(c.GetString("lang"), "general.server_error"),
		)
	}
	return utils.MakeSuccessJSON(fileName)
}
