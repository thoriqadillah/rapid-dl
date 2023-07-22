package rapid

import "context"

type Job interface {
	Execute(ctx context.Context) error
	OnError(err error)
}
