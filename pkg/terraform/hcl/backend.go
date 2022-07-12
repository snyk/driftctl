package hcl

import (
	"fmt"
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
)

type BackendBlock struct {
	Name               string   `hcl:"name,label"`
	Path               string   `hcl:"path,optional"`
	WorkspaceDir       string   `hcl:"workspace_dir,optional"`
	Bucket             string   `hcl:"bucket,optional"`
	Key                string   `hcl:"key,optional"`
	Region             string   `hcl:"region,optional"`
	Prefix             string   `hcl:"prefix,optional"`
	ContainerName      string   `hcl:"container_name,optional"`
	WorkspaceKeyPrefix string   `hcl:"workspace_key_prefix,optional"`
	Remain             hcl.Body `hcl:",remain"`
}

func (b BackendBlock) SupplierConfig(workspace string) *config.SupplierConfig {
	switch b.Name {
	case "local":
		return b.parseLocalBackend()
	case "s3":
		return b.parseS3Backend(workspace)
	case "gcs":
		return b.parseGCSBackend(workspace)
	case "azurerm":
		return b.parseAzurermBackend()
	}
	return nil
}

func (b BackendBlock) parseLocalBackend() *config.SupplierConfig {
	if b.Path == "" {
		return nil
	}
	return &config.SupplierConfig{
		Key:     state.TerraformStateReaderSupplier,
		Backend: backend.BackendKeyFile,
		Path:    path.Join(b.WorkspaceDir, b.Path),
	}
}

func (b BackendBlock) parseS3Backend(ws string) *config.SupplierConfig {
	if b.Bucket == "" || b.Key == "" {
		return nil
	}

	keyPrefix := b.WorkspaceKeyPrefix
	if ws != DefaultStateName {
		if b.WorkspaceKeyPrefix == "" {
			b.WorkspaceKeyPrefix = "env:"
		}
		keyPrefix = path.Join(b.WorkspaceKeyPrefix, ws)
	}

	return &config.SupplierConfig{
		Key:     state.TerraformStateReaderSupplier,
		Backend: backend.BackendKeyS3,
		Path:    path.Join(b.Bucket, keyPrefix, b.Key),
	}
}

func (b BackendBlock) parseGCSBackend(ws string) *config.SupplierConfig {
	if b.Bucket == "" || b.Prefix == "" {
		return nil
	}
	return &config.SupplierConfig{
		Key:     state.TerraformStateReaderSupplier,
		Backend: backend.BackendKeyGS,
		Path:    fmt.Sprintf("%s.tfstate", path.Join(b.Bucket, b.Prefix, ws)),
	}
}

func (b BackendBlock) parseAzurermBackend() *config.SupplierConfig {
	if b.ContainerName == "" || b.Key == "" {
		return nil
	}
	return &config.SupplierConfig{
		Key:     state.TerraformStateReaderSupplier,
		Backend: backend.BackendKeyAzureRM,
		Path:    path.Join(b.ContainerName, b.Key),
	}
}
