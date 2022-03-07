package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	cmderrors "github.com/snyk/driftctl/pkg/cmd/errors"
	"github.com/snyk/driftctl/pkg/cmd/scan/output"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/supplier"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
)

func parseFromFlag(from []string) ([]config.SupplierConfig, error) {

	configs := make([]config.SupplierConfig, 0, len(from))

	for _, flag := range from {
		schemePath := strings.Split(flag, "://")
		if len(schemePath) != 2 || schemePath[1] == "" || schemePath[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nAccepted schemes are: %s",
						strings.Join(supplier.GetSupportedSchemes(), ","),
					),
				),
				"Unable to parse from flag '%s'",
				flag,
			)
		}

		scheme := schemePath[0]
		path := schemePath[1]
		supplierBackend := strings.Split(scheme, "+")
		if len(supplierBackend) > 2 {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(fmt.Sprintf(
					"\nAccepted schemes are: %s",
					strings.Join(supplier.GetSupportedSchemes(), ","),
				),
				),
				"Unable to parse from scheme '%s'",
				scheme,
			)
		}

		supplierKey := supplierBackend[0]
		if !supplier.IsSupplierSupported(supplierKey) {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nAccepted values are: %s",
						strings.Join(supplier.GetSupportedSuppliers(), ","),
					),
				),
				"Unsupported IaC source '%s'",
				supplierKey,
			)
		}

		backendString := ""
		if len(supplierBackend) == 2 {
			backendString = supplierBackend[1]
			if !backend.IsSupported(backendString) {
				return nil, errors.Wrapf(
					cmderrors.NewUsageError(
						fmt.Sprintf(
							"\nAccepted values are: %s",
							strings.Join(backend.GetSupportedBackends(), ","),
						),
					),
					"Unsupported IaC backend '%s'",
					backendString,
				)
			}
		}

		configs = append(configs, config.SupplierConfig{
			Key:     supplierKey,
			Backend: backendString,
			Path:    path,
		})
	}

	return configs, nil
}

func parseOutputFlags(out []string) ([]output.OutputConfig, error) {
	result := make([]output.OutputConfig, 0, len(out))
	for _, v := range out {
		o, err := parseOutputFlag(v)
		if err != nil {
			return result, err
		}
		result = append(result, *o)
	}
	return result, nil
}

func parseOutputFlag(out string) (*output.OutputConfig, error) {
	schemeOpts := strings.Split(out, "://")
	if len(schemeOpts) < 2 || schemeOpts[0] == "" {
		return nil, errors.Wrapf(
			cmderrors.NewUsageError(
				fmt.Sprintf(
					"\nAccepted formats are: %s",
					strings.Join(output.SupportedOutputsExample(), ","),
				),
			),
			"Unable to parse output flag '%s'",
			out,
		)
	}

	o := &output.OutputConfig{
		Key: schemeOpts[0],
	}
	if !output.IsSupported(o.Key) {
		return nil, errors.Wrapf(
			cmderrors.NewUsageError(
				fmt.Sprintf(
					"\nValid formats are: %s",
					strings.Join(output.SupportedOutputsExample(), ","),
				),
			),
			"Unsupported output '%s'",
			o.Key,
		)
	}

	opts := schemeOpts[1:]

	switch o.Key {
	case output.JSONOutputType:
		if len(opts) != 1 || opts[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nMust be of kind: %s",
						output.Example(output.JSONOutputType),
					),
				),
				"Invalid json output '%s'",
				out,
			)
		}
		o.Path = opts[0]
	case output.HTMLOutputType:
		if len(opts) != 1 || opts[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nMust be of kind: %s",
						output.Example(output.HTMLOutputType),
					),
				),
				"Invalid html output '%s'",
				out,
			)
		}
		o.Path = opts[0]
	case output.PlanOutputType:
		if len(opts) != 1 || opts[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nMust be of kind: %s",
						output.Example(output.PlanOutputType),
					),
				),
				"Invalid plan output '%s'",
				out,
			)
		}
		o.Path = opts[0]
	}

	return o, nil
}
