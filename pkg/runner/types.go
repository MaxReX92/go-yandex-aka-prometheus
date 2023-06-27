package runner

import "context"

type Runner interface {
	Start(ctx context.Context) error
}

type Stopper interface {
	Stop(ctx context.Context) error
}
