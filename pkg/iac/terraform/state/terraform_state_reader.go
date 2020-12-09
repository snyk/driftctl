package state

import (
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/hashicorp/terraform/addrs"

	"github.com/cloudskiff/driftctl/pkg/iac"
	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/cloudskiff/driftctl/pkg/iac/terraform/state/backend"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/hashicorp/terraform/states/statefile"

	"github.com/hashicorp/terraform/states"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

const TerraformStateReaderSupplier = "tfstate"

type TerraformStateReader struct {
	config        config.SupplierConfig
	backend       backend.Backend
	deserializers []deserializer.CTYDeserializer
}

func (r *TerraformStateReader) initReader() error {
	b, err := backend.GetBackend(r.config)
	r.backend = b
	if err != nil {
		return err
	}
	return nil
}

func NewReader(config config.SupplierConfig) (*TerraformStateReader, error) {
	reader := TerraformStateReader{config: config, deserializers: iac.Deserializers()}
	err := reader.initReader()
	if err != nil {
		return nil, err
	}
	return &reader, nil
}

func (r *TerraformStateReader) retrieve() (map[string][]cty.Value, error) {

	state, err := read(r.backend)
	defer r.backend.Close()
	if err != nil {
		return nil, err
	}

	stateResources := state.RootModule().Resources
	resMap := make(map[string][]cty.Value)
	for _, stateRes := range stateResources {
		if stateRes.Addr.Resource.Mode != addrs.ManagedResourceMode {
			logrus.WithFields(logrus.Fields{
				"mode": stateRes.Addr.Resource.Mode,
				"name": stateRes.Addr.Resource.Name,
				"type": stateRes.Addr.Resource.Type,
			}).Debug("Skipping state entry as it is not a managed resource")
			continue
		}
		providerType := stateRes.ProviderConfig.Provider.Type
		provider := terraform.Provider(providerType)
		if provider == nil {
			logrus.WithFields(logrus.Fields{
				"providerKey": providerType,
			}).Debug("Unsupported provider found in state")
			continue
		}
		schema := provider.Schema()[stateRes.Addr.Resource.Type]
		for _, instance := range stateRes.Instances {
			decodedVal, err := instance.Current.Decode(schema.Block.ImpliedType())
			if err != nil {
				logrus.Error(err)
				continue
			}
			_, exists := resMap[stateRes.Addr.Resource.Type]
			if !exists {
				resMap[stateRes.Addr.Resource.Type] = []cty.Value{
					decodedVal.Value,
				}
			} else {
				resMap[stateRes.Addr.Resource.Type] = append(resMap[stateRes.Addr.Resource.Type], decodedVal.Value)
			}
		}
	}

	return resMap, nil
}

func (r *TerraformStateReader) decode(values map[string][]cty.Value) ([]resource.Resource, error) {
	results := make([]resource.Resource, 0)
	for _, deserializer := range r.deserializers {

		typ := deserializer.HandledType().String()
		vals, exists := values[typ]
		if !exists {
			logrus.Debugf("No resource of type %s found in state", typ)
			continue
		}
		decodedResources, err := deserializer.Deserialize(vals)
		if err != nil {
			logrus.Warnf("Could not read from decoder for %s: %+v", typ, err)
			continue
		}
		for _, res := range decodedResources {
			logrus.WithFields(logrus.Fields{
				"id":   res.TerraformId(),
				"type": res.TerraformType(),
			}).Debug("Found IAC resource")
			normalisable, ok := res.(resource.NormalizedResource)
			if ok {
				normalizedRes, err := normalisable.NormalizeForState()
				if err != nil {
					logrus.Errorf("Could not normalize state for res %s: %+v", res.TerraformId(), err)
					results = append(results, res)
				}

				if err == nil {
					results = append(results, normalizedRes)
				}
			}
			if !ok {
				results = append(results, res)
			}
		}
	}

	return results, nil
}

func (r *TerraformStateReader) Resources() ([]resource.Resource, error) {
	values, err := r.retrieve()
	if err != nil {
		return nil, err
	}
	return r.decode(values)
}

func read(reader backend.Backend) (*states.State, error) {
	state, err := readState(reader)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func readState(reader backend.Backend) (*states.State, error) {
	state, err := statefile.Read(reader)
	if err != nil {
		return nil, err
	}
	return state.State, nil
}
