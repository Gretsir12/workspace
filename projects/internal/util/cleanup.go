package util

import (
	"context"
	"log"

	"github.com/moby/moby/client"
)

// RemoveManagedContainers принудительно удаляет все контейнеры workspace
// (имеющие метку ide.workspace). Возвращает число удалённых.
func (d *Docker) RemoveManagedContainers(ctx context.Context) (int, error) {
	res, err := d.api.ContainerList(ctx, client.ContainerListOptions{
		All:     true,
		Filters: make(client.Filters).Add("label", labelWorkspace).Add("status", "created", "exited", "dead"),
	})
	if err != nil {
		return 0, err
	}

	removed := 0
	for _, c := range res.Items {
		if _, err := d.api.ContainerRemove(ctx, c.ID, client.ContainerRemoveOptions{Force: true}); err != nil {
			log.Printf("remove container %s: %v", c.ID, err)
			continue
		}
		removed++
	}
	return removed, nil
}
