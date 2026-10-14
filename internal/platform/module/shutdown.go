package module

import "context"

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}
