package main

import (
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
)

// UploadPicture is upload team logo handler for manager.
func (s *Service) UploadPicture(c *gin.Context) (int, interface{}) {
	file, err := c.FormFile("picture")
	if err != nil {
		return s.makeErrJSON(400, 40000,
			s.I18n.T(c.GetString("lang"), "file.select_picture"),
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
		return s.makeErrJSON(500, 50000,
			s.I18n.T(c.GetString("lang"), "general.server_error"),
		)
	}
	return s.makeSuccessJSON(fileName)
}
