package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"workspaces/projects/internal/models"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	nixBaseImage     = "nixos/nix:latest"
	containerWorkDir = "/app"
)

// main function
func StartProjectContainer(ctx context.Context, project *models.Project, workspace *models.Workspace) (string, error) {
	cli, err := newDockerClient()
	if err != nil {
		return "", err
	}
	defer cli.Close()

	if err := ensureImageExists(ctx, cli, nixBaseImage); err != nil {
		return "", fmt.Errorf("pull image: %w", err)
	}

	containerName := buildContainerName(project.ID)
	cmd := buildNixCommand(workspace.Dependencies)

	config := newContainerConfig(cmd)
	hostConfig := newHostConfig(project.Path)

	containerID, err := createContainer(ctx, cli, containerName, config, hostConfig)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}

	if err := startContainer(ctx, cli, containerID); err != nil {
		return "", fmt.Errorf("start container: %w", err)
	}

	return containerID, nil
}

// helper functions

func newDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("init docker client: %w", err)
	}
	return cli, nil
}

func ensureImageExists(ctx context.Context, cli *client.Client, imageName string) error {
	// В вашей версии используется types.ImagePullOptions
	reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Вычитываем поток, чтобы дождаться завершения скачивания
	_, _ = io.Copy(os.Stdout, reader)
	return nil
}

func buildContainerName(projectID uint32) string {
	return fmt.Sprintf("workspace-project-%d", projectID)
}

func buildNixCommand(dependencies []models.Dependency) []string {
	if len(dependencies) == 0 {
		return []string{"bash"}
	}

	cmd := []string{"nix-shell", "-p"}
	for _, dep := range dependencies {
		cmd = append(cmd, string(dep))
	}
	return cmd
}

// ИСПРАВЛЕНО: заместо containerimage.Config написано container.Config
func newContainerConfig(cmd []string) *container.Config {
	return &container.Config{
		Image:      nixBaseImage,
		Cmd:        cmd,
		WorkingDir: containerWorkDir,
		Tty:        true,
		OpenStdin:  true,
	}
}

func newHostConfig(hostPath string) *container.HostConfig {
	bindMount := fmt.Sprintf("%s:%s", hostPath, containerWorkDir)
	return &container.HostConfig{
		Binds: []string{bindMount},
	}
}

func createContainer(
	ctx context.Context,
	cli *client.Client,
	name string,
	config *container.Config,
	hostConfig *container.HostConfig,
) (string, error) {
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func startContainer(ctx context.Context, cli *client.Client, containerID string) error {
	return cli.ContainerStart(ctx, containerID, container.StartOptions{})
}
