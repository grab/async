package core

import (
	"context"
)

type plan interface {
	IsSequential() bool
}

type masterPlan interface {
	plan
	Execute(ctx context.Context) error
}

type pre interface {
	PreExecute(p any) error
}

type post interface {
	PostExecute(p any) error
}
