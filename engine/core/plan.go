package core

import "context"

type plan interface {
	Execute(ctx context.Context) error
	ExecuteWithin(ctx context.Context, masterPlan any) error
	IsSequential() bool
}

type prePlan interface {
	PreExecute(ctx context.Context, masterPlan any) error
}

type postPlan interface {
	PostExecute(ctx context.Context, masterPlan any) error
}