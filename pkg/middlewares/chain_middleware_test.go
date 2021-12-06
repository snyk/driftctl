package middlewares

import (
	"errors"
	"testing"

	"github.com/snyk/driftctl/pkg/resource"
)

var callCounters map[string]int

type FakeMiddleware struct {
	Name string
	Err  error
}

func (m FakeMiddleware) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	callCounters[m.Name]++
	return m.Err
}

func TestChainMiddleware(t *testing.T) {

	callCounters = make(map[string]int)

	fakeMiddleware1 := FakeMiddleware{
		Name: "1",
	}
	fakeMiddleware2 := FakeMiddleware{
		Name: "2",
	}

	middleware := NewChain(fakeMiddleware1, fakeMiddleware2)
	remoteResources := []*resource.Resource{}
	stateResources := []*resource.Resource{}
	err := middleware.Execute(&remoteResources, &stateResources)

	if err != nil {
		t.Error("A middleware returned an error")
	}

	if callCounters["1"] != 1 {
		t.Error("Middleware 1 was not called correctly")
	}

	if callCounters["2"] != 1 {
		t.Error("Middleware 2 was not called correctly")
	}

}

func TestChainMiddlewareErrorShouldStopExecution(t *testing.T) {

	callCounters = make(map[string]int)

	fakeMiddleware1 := FakeMiddleware{
		Name: "1",
		Err:  errors.New("Test error"),
	}
	fakeMiddleware2 := FakeMiddleware{
		Name: "2",
	}

	middleware := NewChain(fakeMiddleware1, fakeMiddleware2)
	remoteResources := []*resource.Resource{}
	stateResources := []*resource.Resource{}
	err := middleware.Execute(&remoteResources, &stateResources)

	if err == nil {
		t.Error("No error were reported")
	}

	if err.Error() != "Test error" {
		t.Error("Unknown error reported")
	}

	if callCounters["1"] != 1 {
		t.Error("Middleware 1 was not called correctly")
	}

	if callCounters["2"] != 0 {
		t.Error("Middleware 2 was called after error happen in middleware 1")
	}

}
