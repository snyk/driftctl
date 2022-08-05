package diagnostic

import (
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/resource"
)

type Diagnostic interface {
	Code() string
	Message() string
	ResourceType() string
	Resource() *resource.Resource
}

type diagnosticImpl struct {
	alert alerter.Alert
}

func (d *diagnosticImpl) Code() string {
	if _, ok := d.alert.(*alerts.RemoteAccessDeniedAlert); ok {
		return "ACCESS_DENIED"
	}
	return "UNKNOWN_ERROR"
}

func (d *diagnosticImpl) Message() string {
	return d.alert.Message()
}

func (d *diagnosticImpl) ResourceType() string {
	ty := ""
	if d.Resource() != nil {
		ty = d.Resource().ResourceType()
	}
	return ty
}

func (d *diagnosticImpl) Resource() *resource.Resource {
	return d.alert.Resource()
}

type Diagnostics []Diagnostic

func FromAlerts(alertMap alerter.Alerts) Diagnostics {
	var results Diagnostics
	for _, v := range alertMap {
		for _, alert := range v {
			diag := &diagnosticImpl{alert}
			results = append(results, diag)
		}
	}
	return results
}
