package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	"github.com/vidar-team/Cardinal/src/utils"
	"regexp"
	"strconv"
	"time"
)

type dockerImage struct {
	Layers []struct {
		Digest      string `json:"digest"`
		Instruction string `json:"instruction"`
		Size        int    `json:"size"`
	} `json:"layers"`
}

func (s *Service) getImageData(c *gin.Context) (int, interface{}) {
	type inputForm struct {
		User  string `binding:"required"`
		Image string `binding:"required"`
		Tag   string `binding:"required"`
	}

	var form inputForm
	err := c.BindJSON(&form)
	if err != nil {
		return utils.MakeErrJSON(400, 40041, "payload error")
	}
	req := gorequest.New().Get(fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/%s/images", form.User, form.Image, form.Tag))
	req.Timeout(5 * time.Second)
	resp, body, _ := req.End()
	if resp == nil || resp.StatusCode != 200 {
		return utils.MakeErrJSON(500, 50028, "request dockerhub failed")
	}

	var imageInfo []dockerImage
	if err := json.Unmarshal([]byte(body), &imageInfo); err != nil {
		return utils.MakeErrJSON(500, 50029, "dockerhub json unmarshal error")
	}
	if len(imageInfo) == 0 {
		return utils.MakeErrJSON(500, 50030, "dockerhub repo is empty")
	}

	var ports []int

	reg := regexp.MustCompile(`EXPOSE\s+(\d+)`)
	// parse dockerfile.
	for _, layer := range imageInfo[0].Layers {
		portStr := reg.FindStringSubmatch(layer.Instruction)
		for _, str := range portStr {
			port, err := strconv.Atoi(str)
			if err == nil {
				ports = append(ports, port)
			}
		}
	}

	return utils.MakeSuccessJSON(gin.H{
		"Ports": ports,
	})
}
