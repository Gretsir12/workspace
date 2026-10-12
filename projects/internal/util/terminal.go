package util

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

const labelWorkspace = "ide.workspace"

type Terminal struct {
	d    *Docker
	id   string
	conn client.HijackedResponse
}

// StartTerminal создаёт контейнер из образа, подключается к его TTY и запускает.
// Attach делается ДО старта, чтобы не потерять начало вывода.
func (d *Docker) StartTerminal(ctx context.Context, image string, workspaceID uint32) (*Terminal, error) {
	created, err := d.api.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:        image,
			Cmd:          []string{"sh"},
			Tty:          true,
			OpenStdin:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Labels: map[string]string{
				labelManaged:   "true",
				labelWorkspace: strconv.FormatUint(uint64(workspaceID), 10),
			},
		},
		HostConfig: &container.HostConfig{
			Resources: container.Resources{
				Memory:    512 << 20,
				PidsLimit: ptr(int64(256)),
			},
			// NetworkMode: "none", // если интернет внутри не нужен
		},
	})
	if err != nil {
		return nil, fmt.Errorf("container create: %w", err)
	}

	t := &Terminal{d: d, id: created.ID}

	att, err := d.api.ContainerAttach(ctx, created.ID, client.ContainerAttachOptions{
		Stream: true, Stdin: true, Stdout: true, Stderr: true,
	})
	if err != nil {
		t.remove()
		return nil, fmt.Errorf("container attach: %w", err)
	}
	t.conn = att.HijackedResponse

	if _, err := d.api.ContainerStart(ctx, created.ID, client.ContainerStartOptions{}); err != nil {
		t.Close()
		return nil, fmt.Errorf("container start: %w", err)
	}
	return t, nil
}

func (t *Terminal) ID() string { return t.id }

func (t *Terminal) Read(p []byte) (int, error)  { return t.conn.Reader.Read(p) }
func (t *Terminal) Write(p []byte) (int, error) { return t.conn.Conn.Write(p) }

func (t *Terminal) Resize(ctx context.Context, cols, rows uint) error {
	_, err := t.d.api.ContainerResize(ctx, t.id, client.ContainerResizeOptions{Width: cols, Height: rows})
	return err
}

// Close закрывает attach и удаляет контейнер. Контекст свой, потому что
// к моменту вызова контекст запроса уже отменён.
func (t *Terminal) Close() {
	t.conn.Close()
	t.remove()
}

func (t *Terminal) remove() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, _ = t.d.api.ContainerRemove(ctx, t.id, client.ContainerRemoveOptions{Force: true})
}

func ptr[T any](v T) *T { return &v }

var _ io.ReadWriter = (*Terminal)(nil)
