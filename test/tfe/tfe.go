package test_tfe

import "github.com/hashicorp/go-tfe"

type Workspaces interface {
	tfe.Workspaces
}

type StateVersions interface {
	tfe.StateVersions
}
