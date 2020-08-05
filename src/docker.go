package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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
		"Image": fmt.Sprintf("%s/%s:%s", form.User, form.Image, form.Tag),
		"Name":  form.Image,
		"Ports": ports,
	})
}

func (s *Service) deployFromDocker(c *gin.Context) (int, interface{}) {
	type port struct {
		In  uint `binding:"required"`
		Out uint `binding:"required"`
	}

	type inputForm struct {
		Image       string `binding:"required"`
		Challenge   uint   `binding:"required"`
		IP          string `binding:"required"`
		ServicePort uint   `binding:"required"`
		SSHPort     uint   `binding:"required"`
		RootSSHName string `binding:"required"`
		UserSSHName string `binding:"required"`
		Description string `binding:"required"`
		Ports       []port `binding:"required"`
	}

	var form inputForm
	err := c.BindJSON(&form)
	if err != nil {
		return utils.MakeErrJSON(400, 40042, "payload error")
	}

	// Pre-check

	// challenge exist
	var challenge Challenge
	s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: form.Challenge}}).Find(&challenge)
	if challenge.ID == 0 {
		return utils.MakeErrJSON(404, 40406, "payload error")
	}
	// port check
	if form.ServicePort == 0 || form.ServicePort > 65536 || form.SSHPort == 0 || form.SSHPort > 65536 {
		return utils.MakeErrJSON(400, 40043, "error port")
	}
	for i1, p1 := range form.Ports {
		if p1.In == 0 || p1.In > 65536 || p1.Out == 0 || p1.Out > 65536 {
			return utils.MakeErrJSON(400, 40043, "error port")
		}
		for i2, p2 := range form.Ports {
			if i1 != i2 && (p1.In == p2.In || p1.Out == p2.Out) {
				return utils.MakeErrJSON(400, 40044, "error port")
			}
		}
	}
	// check name
	if form.RootSSHName == form.UserSSHName {
		return utils.MakeErrJSON(400, 40045, "name repeat")
	}

	// get the docker image
	// TODO

	return utils.MakeSuccessJSON("")
}
