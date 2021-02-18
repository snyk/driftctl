package alerter

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type AlerterInterface interface {
	SendAlert(key string, alert Alert)
}

type Alerter struct {
	alerts   Alerts
	alertsCh chan Alerts
	doneCh   chan bool
}

func NewAlerter() *Alerter {
	var alerter = &Alerter{
		alerts:   make(Alerts),
		alertsCh: make(chan Alerts),
		doneCh:   make(chan bool),
	}

	go alerter.run()

	return alerter
}

func (a *Alerter) run() {
	defer func() { a.doneCh <- true }()
	for alert := range a.alertsCh {
		for k, v := range alert {
			if val, ok := a.alerts[k]; ok {
				a.alerts[k] = append(val, v...)
			} else {
				a.alerts[k] = v
			}
		}
	}
}

func (a *Alerter) SetAlerts(alerts Alerts) {
	a.alerts = alerts
}

func (a *Alerter) Retrieve() Alerts {
	close(a.alertsCh)
	<-a.doneCh
	return a.alerts
}

func (a *Alerter) SendAlert(key string, alert Alert) {
	a.alertsCh <- Alerts{
		key: []Alert{alert},
	}
}

func (a *Alerter) IsResourceIgnored(res resource.Resource) bool {
	alert, alertExists := a.alerts[fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId())]
	wildcardAlert, wildcardAlertExists := a.alerts[res.TerraformType()]
	shouldIgnoreAlert := a.shouldBeIgnored(alert)
	shouldIgnoreWildcardAlert := a.shouldBeIgnored(wildcardAlert)
	return (alertExists && shouldIgnoreAlert) || (wildcardAlertExists && shouldIgnoreWildcardAlert)
}

func (a *Alerter) shouldBeIgnored(alert []Alert) bool {
	for _, a := range alert {
		if a.ShouldIgnoreResource() {
			return true
		}
	}
	return false
}
