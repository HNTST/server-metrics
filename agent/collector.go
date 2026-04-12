package main

import (
	md "agent/models"
	"context"
	"strings"
	"time"

	"github.com/moby/moby/client"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func getDockerContainersInfo(cli *client.Client) ([]md.ContainerInfo, error) {
	var res []md.ContainerInfo

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cont, err := cli.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return res, err
	}
	for _, c := range cont.Items {
		name := "unknown"
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		res = append(res, md.ContainerInfo{
			ContainerID: c.ID[:12],
			Name:        name,
			State:       string(c.State),
			Image:       c.Image,
		})
	}
	return res, nil
}

func getServerInfo(cfg *Config, conts []md.ContainerInfo) (md.Metrics, error) {
	var post md.Metrics
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return post, err
	}
	mm, err := mem.VirtualMemory()
	if err != nil {
		return post, err
	}
	dk, err := disk.Usage("/")
	if err != nil {
		return post, err
	}
	post = md.Metrics{
		Host:       cfg.HostName,
		Cpu:        percentages[0],
		MemTotal:   mm.Total,
		MemUsage:   mm.Used,
		SpaceTotal: dk.Total,
		SpaceUsage: dk.Used,
		Containers: conts,
		SentAt:     time.Now(),
	}
	return post, nil
}
