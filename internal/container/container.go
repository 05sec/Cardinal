package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/internal/utils"
	"time"
)

// GetContainerList returns all the docker containers.
func GetContainerList(c *gin.Context) (int, interface{}) {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}
	return utils.MakeSuccessJSON(containers)
}

// StartContainer starts the container.
func StartContainer(c *gin.Context) (int, interface{}) {
	var inputForm struct {
		ContainerId string `binding:"required"`
	}
	err := c.ShouldBind(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40099, "payload error")
	}

	containerID := inputForm.ContainerId
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}
	err = cli.ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}

	return utils.MakeSuccessJSON("启动成功")
}

// StopContainer starts the container.
func StopContainer(c *gin.Context) (int, interface{}) {
	var inputForm struct {
		ContainerId string `binding:"required"`
	}
	err := c.ShouldBind(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40099, "payload error")
	}

	containerID := inputForm.ContainerId
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}
	timeout := 15 * time.Second
	err = cli.ContainerStop(context.Background(), containerID, &timeout)
	if err != nil {
		return utils.MakeErrJSON(500, 59999, err)
	}

	return utils.MakeSuccessJSON("关闭成功")
}
