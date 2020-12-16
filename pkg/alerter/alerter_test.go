package alerter

import (
	"reflect"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resource2 "github.com/cloudskiff/driftctl/test/resource"
)

func TestAlerter_Alert(t *testing.T) {
	cases := []struct {
		name     string
		alerts   Alerts
		expected Alerts
	}{
		{
			name:     "TestNoAlerts",
			alerts:   nil,
			expected: Alerts{},
		},
		{
			name: "TestWithSingleAlert",
			alerts: Alerts{
				"fakeres.foobar": []Alert{
					{
						Message:              "This is an alert",
						ShouldIgnoreResource: false,
					},
				},
			},
			expected: Alerts{
				"fakeres.foobar": []Alert{
					{
						Message:              "This is an alert",
						ShouldIgnoreResource: false,
					},
				},
			},
		},
		{
			name: "TestWithMultipleAlerts",
			alerts: Alerts{
				"fakeres.foobar": []Alert{
					{
						Message:              "This is an alert",
						ShouldIgnoreResource: false,
					},
					{
						Message:              "This is a second alert",
						ShouldIgnoreResource: true,
					},
				},
				"fakeres.barfoo": []Alert{
					{
						Message:              "This is a third alert",
						ShouldIgnoreResource: true,
					},
				},
			},
			expected: Alerts{
				"fakeres.foobar": []Alert{
					{
						Message:              "This is an alert",
						ShouldIgnoreResource: false,
					},
					{
						Message:              "This is a second alert",
						ShouldIgnoreResource: true,
					},
				},
				"fakeres.barfoo": []Alert{
					{
						Message:              "This is a third alert",
						ShouldIgnoreResource: true,
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			alerter := NewAlerter()

			for k, v := range c.alerts {
				for _, a := range v {
					alerter.SendAlert(k, a)
				}
			}

			if eq := reflect.DeepEqual(alerter.GetAlerts(), c.expected); !eq {
				t.Errorf("Got %+v, expected %+v", alerter.GetAlerts(), c.expected)
			}
		})
	}
}

func TestAlerter_IgnoreResources(t *testing.T) {
	cases := []struct {
		name     string
		alerts   Alerts
		resource resource.Resource
		expected bool
	}{
		{
			name:   "TestNoAlerts",
			alerts: Alerts{},
			resource: &resource2.FakeResource{
				Type: "fakeres",
				Id:   "foobar",
			},
			expected: false,
		},
		{
			name: "TestShouldNotBeIgnoredWithAlerts",
			alerts: Alerts{
				"fakeres": {
					{
						Message: "Should not be ignored",
					},
				},
				"fakeres.foobar": {
					{
						Message: "Should not be ignored",
					},
				},
				"fakeres.barfoo": {
					{
						Message: "Should not be ignored",
					},
				},
				"other.resource": {
					{
						Message: "Should not be ignored",
					},
				},
			},
			resource: &resource2.FakeResource{
				Type: "fakeres",
				Id:   "foobar",
			},
			expected: false,
		},
		{
			name: "TestShouldBeIgnoredWithAlertsOnWildcard",
			alerts: Alerts{
				"fakeres": {
					{
						Message:              "Should be ignored",
						ShouldIgnoreResource: true,
					},
				},
				"other.foobaz": {
					{
						Message:              "Should be ignored",
						ShouldIgnoreResource: true,
					},
				},
				"other.resource": {
					{
						Message: "Should not be ignored",
					},
				},
			},
			resource: &resource2.FakeResource{
				Type: "fakeres",
				Id:   "foobar",
			},
			expected: true,
		},
		{
			name: "TestShouldBeIgnoredWithAlertsOnResource",
			alerts: Alerts{
				"fakeres": {
					{
						Message:              "Should be ignored",
						ShouldIgnoreResource: true,
					},
				},
				"other.foobaz": {
					{
						Message:              "Should be ignored",
						ShouldIgnoreResource: true,
					},
				},
				"other.resource": {
					{
						Message: "Should not be ignored",
					},
				},
			},
			resource: &resource2.FakeResource{
				Type: "other",
				Id:   "foobaz",
			},
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			alerter := NewAlerter()
			alerter.SetAlerts(c.alerts)
			if got := alerter.IsResourceIgnored(c.resource); got != c.expected {
				t.Errorf("Got %+v, expected %+v", got, c.expected)
			}
		})
	}
}
