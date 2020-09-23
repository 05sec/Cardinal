package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

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
