package dckr

import (
	"alpr/models"
	"alpr/utils"
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"strings"
)

type AlprDockerManager struct {
	Client *client.Client
}

var imageName = "gokalpgoren/openalpr_local"

func (d *AlprDockerManager) getImage() (*types.ImageSummary, error) {
	results, err := d.Client.ImageList(context.Background(), types.ImageListOptions{All: false})
	if err != nil {
		log.Println("an error occurred while searching a docker image, err: ", err.Error())
		return nil, err
	}

	name := imageName + ":latest"
	for _, result := range results {
		for _, tag := range result.RepoTags {
			if tag == name {
				return &result, nil
			}
		}
	}

	return nil, nil
}

func (d *AlprDockerManager) pullImage() error {
	name := imageName + ":latest"
	out, err := d.Client.ImagePull(context.Background(), name, types.ImagePullOptions{})
	if err != nil {
		log.Println("an error occurred while pulling the image, err: ", err.Error())
		return err
	}
	defer func(out io.ReadCloser) {
		err := out.Close()
		if err != nil {
			log.Println("an error occurred while closing pull image reader, err: ", err.Error())
		}
	}(out)
	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		log.Println("an error occurred while copying reader, err: ", err.Error())
		return err
	}

	return nil
}

func (d *AlprDockerManager) InitImage() error {
	img, _ := d.getImage()
	if img == nil {
		err := d.pullImage()
		if err != nil {
			log.Println("an error occurred while pulling the image, err: ", err.Error())
			return err
		}
	}

	return nil
}

func (d *AlprDockerManager) RemoveContainers(all bool) (int, error) {
	ctx := context.Background()
	containers, err := d.Client.ContainerList(ctx, types.ContainerListOptions{All: all})
	if err != nil {
		log.Println("an error occurred while getting the container, err: ", err.Error())
		return 0, err
	}

	count := 0
	startWith := "/" + containerNamePrefix
	for _, cntr := range containers {
		for _, cname := range cntr.Names {
			if strings.HasPrefix(cname, startWith) {
				err := d.Client.ContainerRemove(ctx, cntr.ID, types.ContainerRemoveOptions{Force: true})
				if err != nil {
					log.Println("an error occurred while removing the container, err: ", err.Error())
				} else {
					count++
				}
			}
		}
	}

	return count, nil
}

func (d *AlprDockerManager) GetContainer(name string) (*types.Container, error) {
	containers, err := d.Client.ContainerList(context.Background(), types.ContainerListOptions{All: false})
	if err != nil {
		log.Println("an error occurred while getting the container, err: ", err.Error())
		return nil, err
	}

	name = "/" + name
	for _, cntr := range containers {
		for _, cname := range cntr.Names {
			if cname == name {
				return &cntr, nil
			}
		}
	}

	return nil, nil
}

func (d *AlprDockerManager) StartContainer(name string) (*types.Container, error) {
	ctx := context.Background()
	var cntr *types.Container
	containers, _ := d.Client.ContainerList(ctx, types.ContainerListOptions{All: true})
	for _, ctr := range containers {
		for _, cname := range ctr.Names {
			if cname == "/"+name {
				cntr = &ctr
			}
		}
	}
	if cntr != nil {
		if cntr.Status != "running" {
			err := d.Client.ContainerStart(ctx, cntr.ID, types.ContainerStartOptions{})
			if err != nil {
				log.Println("an error occurred during the starting the container, err: ", err.Error())
				return nil, err
			}
		}
		return cntr, nil
	} else {
		cc := container.Config{}
		cc.Image = imageName
		cc.OpenStdin = true

		hc := container.HostConfig{}
		m := mount.Mount{}
		m.Type = mount.TypeBind
		m.ReadOnly = true
		m.Source = utils.GetTempDir()
		m.Target = "/data"
		hc.Mounts = []mount.Mount{m}
		hc.RestartPolicy = container.RestartPolicy{Name: "unless-stopped"}

		resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, name)
		if err != nil {
			log.Println("an error occurred during the creating the container, err: ", err.Error())
			return nil, err
		}

		options := types.ContainerStartOptions{}
		err = d.Client.ContainerStart(ctx, resp.ID, options)

		return d.GetContainer(name)
	}
}

func (d *AlprDockerManager) ExecRun(ctr *types.Container, fileName string) (*models.AlprResult, error) {
	ctx := context.Background()
	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  true,
		Cmd:          []string{"alpr", "-c", "eu", fileName, "-j"},
	}

	resp, err := d.Client.ContainerExecCreate(ctx, ctr.ID, config)
	if err != nil {
		log.Println("an error occurred during the exec create, err: ", err.Error())
		return nil, err
	}
	ex, err := d.Client.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	defer func() {
		ex.Close()
	}()
	if err != nil {
		log.Println("an error occurred during the exec attach, err: ", err.Error())
		return nil, err
	}
	if b, err := io.ReadAll(ex.Reader); err == nil {
		r := &models.AlprResult{}
		err := json.Unmarshal(b[8:], r)
		if err != nil {
			log.Println("an error occurred during the deserializing json result, err: ", err.Error())
			return nil, err
		}
		r.FileName = fileName
		return r, nil
	} else {
		log.Println("an error occurred while reading the buffer, err: ", err.Error())
		return nil, err
	}
}
