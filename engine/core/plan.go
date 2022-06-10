package core

import "context"

type plan interface {
	Execute(ctx context.Context) error
	IsSequential() bool
}

type pre interface {
	PreExecute(p any) error
}

type post interface {
	PostExecute(p any) error
}