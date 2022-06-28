package repository

import "context"

type pager interface {
	Err() error
	NextPage(ctx context.Context) bool
}
