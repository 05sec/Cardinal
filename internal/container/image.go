package container

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/internal/utils"
	"io"
)

// GetImageList returns the images list.
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

// FetchImage pull the image from the given registry.
func FetchImage(registry string, repo string, name string, tag string) error {
	dockerCli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	events, err := dockerCli.ImagePull(
		context.Background(),
		fmt.Sprintf("%s/%s/%s:%s", registry, repo, name, tag),
		types.ImagePullOptions{},
	)
	if err != nil {
		return err
	}
	d := json.NewDecoder(events)

	type Event struct {
		Status         string `json:"status"`
		Error          string `json:"error"`
		Progress       string `json:"progress"`
		ProgressDetail struct {
			Current int `json:"current"`
			Total   int `json:"total"`
		} `json:"progressDetail"`
	}

	var event *Event
	for {
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		fmt.Printf("EVENT: %+v\n", event)
	}
	return nil
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
