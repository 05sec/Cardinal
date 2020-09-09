package container

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	uuid "github.com/satori/go.uuid"
	"github.com/vidar-team/Cardinal/internal/livelog"
	"github.com/vidar-team/Cardinal/internal/utils"
	"io"
	"regexp"
	"strconv"
	"time"
)

const DockerHubRegistry = "docker.io"

type dockerImage struct {
	Architecture string `json:"architecture"`
	Digest       string `json:"digest"`
	Layers       []struct {
		Digest      string `json:"digest"`
		Instruction string `json:"instruction"`
		Size        int    `json:"size"`
	} `json:"layers"`
}

// GetImageList returns the docker instance images list.
func GetImageList(c *gin.Context) (int, interface{}) {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}
	containers, err := cli.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}
	return utils.MakeSuccessJSON(containers)
}

// GetImageData returns the image data from dockerhub.
func GetImageData(c *gin.Context) (int, interface{}) {
	type inputForm struct {
		User      string `binding:"required"`
		ImageName string `binding:"required"`
		Tag       string `binding:"required"`
	}

	var form inputForm
	err := c.BindJSON(&form)
	if err != nil {
		return utils.MakeErrJSON(400, 40041, "payload error")
	}
	req := gorequest.New().Get(fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/%s/images", form.User, form.ImageName, form.Tag))
	req.Timeout(5 * time.Second)
	resp, body, _ := req.End()
	if resp == nil || resp.StatusCode != 200 {
		return utils.MakeErrJSON(500, 50028, "request dockerhub failed")
	}

	var imageLists []dockerImage
	if err := json.Unmarshal([]byte(body), &imageLists); err != nil {
		return utils.MakeErrJSON(500, 50029, "dockerhub json unmarshal error")
	}
	if len(imageLists) == 0 {
		return utils.MakeErrJSON(500, 50030, "dockerhub repo is empty")
	}

	imageInfo := imageLists[0]
	var ports []int
	reg := regexp.MustCompile(`EXPOSE\s+(\d+)`)
	// parse dockerfile.
	// TODO: maybe we can find a package to parse the dockerfile.
	for _, layer := range imageInfo.Layers {
		portStr := reg.FindStringSubmatch(layer.Instruction)
		for _, str := range portStr {
			port, err := strconv.Atoi(str)
			if err == nil {
				ports = append(ports, port)
			}
		}
	}

	return utils.MakeSuccessJSON(gin.H{
		"Name":         form.ImageName,
		"User":         form.User,
		"Tag":          form.Tag,
		"Image":        fmt.Sprintf("%s/%s:%s", form.User, form.ImageName, form.Tag),
		"Digest":       imageInfo.Digest,
		"Architecture": imageInfo.Architecture,
		"Ports":        ports,
	})
}

func PullImage(c *gin.Context) (int, interface{}) {
	type inputForm struct {
		Image string `binding:"required"`
	}

	var form inputForm
	err := c.BindJSON(&form)
	if err != nil {
		return utils.MakeErrJSON(400, 40099, "payload error")
	}

	livelogId, err := pullImage(DockerHubRegistry, form.Image)
	if err != nil {
		return utils.MakeErrJSON(500, 50099, err)
	}

	// return the livelog id.
	return utils.MakeSuccessJSON(livelogId)
}

// pull the image from the given registry.
func pullImage(registry string, image string) (string, error) {
	dockerCli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	// new livelog stream.
	livelogId := uuid.NewV4().String()
	err = livelog.Stream.Create(livelogId)
	if err != nil {
		return "", err
	}

	type Event struct {
		Status         string `json:"status"`
		Error          string `json:"error"`
		Progress       string `json:"progress"`
		ProgressDetail struct {
			Current int `json:"current"`
			Total   int `json:"total"`
		} `json:"progressDetail"`
	}

	go func() {
		var event *Event
		events, err := dockerCli.ImagePull(
			context.Background(),
			fmt.Sprintf("%s/%s", registry, image),
			types.ImagePullOptions{},
		)
		if err != nil {
			_ = livelog.Stream.Write(livelogId, livelog.NewLine("end", err))
		}
		d := json.NewDecoder(events)

		for {
			if err := d.Decode(&event); err != nil {
				if err == io.EOF {
					_ = livelog.Stream.Write(livelogId, livelog.NewLine("end", ""))
					break
				}
				break
			}
			_ = livelog.Stream.Write(livelogId, livelog.NewLine("progress", event))
		}
	}()

	return livelogId, nil
}

func DeleteImage(c *gin.Context) (int, interface{}) {
	imageId, ok := c.GetQuery("imageId")
	if !ok {
		return utils.MakeErrJSON(400, 40099, "payload error")
	}

	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}
	_, err = cli.ImageRemove(context.Background(), imageId, types.ImageRemoveOptions{})
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}

	return utils.MakeSuccessJSON("删除成功")
}
