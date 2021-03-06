package config

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetRateLimitConfig(t *testing.T) {
	scheme := testScheme()

	scenarios := []struct {
		Name        string
		InitialObjs []runtime.Object
		Namespace   string
		Assert      func(client.Client, *RateLimitConfig, error) error
	}{
		{
			Name:      "Success",
			Namespace: "redhat-test-operator",
			InitialObjs: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: v1.ObjectMeta{
						Name:      "sku-limits-managed-api-service",
						Namespace: "redhat-test-operator",
					},
					Data: map[string]string{
						"rate_limit": `
						{
							"RHOAM SERVICE SKU": {
								"unit": "minute",
								"requests_per_unit": 42
							}
						}
						`,
					},
				},
			},
			Assert: func(c client.Client, config *RateLimitConfig, err error) error {
				if err != nil {
					return fmt.Errorf("Unexpected error: %v", err)
				}

				expectedConfig := &RateLimitConfig{
					Unit:            "minute",
					RequestsPerUnit: 42,
				}

				if !reflect.DeepEqual(config, expectedConfig) {
					return fmt.Errorf("Obtained invalid config. Expected %v, but got %v", expectedConfig, config)
				}

				return nil
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			client := fake.NewFakeClientWithScheme(scheme, scenario.InitialObjs...)
			config, err := GetRateLimitConfig(context.TODO(), client, scenario.Namespace)

			if err := scenario.Assert(client, config, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGetAlertConfig(t *testing.T) {
	scheme := testScheme()

	scenarios := []struct {
		Name        string
		InitialObjs []runtime.Object
		Namespace   string
		Assert      func(client.Client, map[string]*AlertConfig, error) error
	}{
		{
			Name:      "Success",
			Namespace: "redhat-test-operator",
			InitialObjs: []runtime.Object{
				&corev1.ConfigMap{
					ObjectMeta: v1.ObjectMeta{
						Name:      "rate-limit-alerts",
						Namespace: "redhat-test-operator",
					},
					Data: map[string]string{
						"alerts": `
						{
							"alert-1": {
								"ruleName": "Rule1",
								"level": "warning",
								"minRate": "80%",
								"maxRate": "90%",
								"period": "2h"
							}
						}
						`,
					},
				},
			},
			Assert: func(c client.Client, config map[string]*AlertConfig, err error) error {
				if err != nil {
					return fmt.Errorf("Unexpected error: %v", err)
				}

				alertConfig, ok := config["alert-1"]
				if !ok {
					return fmt.Errorf("expected key alert-1 not found in resulting config")
				}

				maxRate := "90%"

				expectedConfig := &AlertConfig{
					RuleName: "Rule1",
					Level:    "warning",
					MaxRate:  &maxRate,
					MinRate:  "80%",
					Period:   "2h",
				}

				if !reflect.DeepEqual(alertConfig, expectedConfig) {
					return fmt.Errorf("Obtained invalid config. Expected %v, but got %v", expectedConfig, alertConfig)
				}

				return nil
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			client := fake.NewFakeClientWithScheme(scheme, scenario.InitialObjs...)
			config, err := GetAlertConfig(context.TODO(), client, scenario.Namespace)

			if err := scenario.Assert(client, config, err); err != nil {
				t.Error(err)
			}
		})
	}
}

func testScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	return scheme
}
