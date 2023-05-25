package state

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/azurerm"
	"github.com/snyk/driftctl/enumeration/remote/github"
	"github.com/snyk/driftctl/enumeration/remote/google"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/output"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"
	resourceazure "github.com/snyk/driftctl/pkg/resource/azurerm"
	resourcegithub "github.com/snyk/driftctl/pkg/resource/github"
	resourcegoogle "github.com/snyk/driftctl/pkg/resource/google"
	testresource "github.com/snyk/driftctl/test/resource"

	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/snyk/driftctl/test/mocks"

	"github.com/r3labs/diff/v2"
)

func TestReadStateValid(t *testing.T) {
	reader, _ := os.Open("testdata/v4/valid.tfstate")
	_, err := readState("terraform.tfstate", reader)
	if err != nil {
		t.Errorf("Unable to read state, %s", err)
		return
	}
}

func TestReadStateInvalid(t *testing.T) {
	reader, _ := os.Open("testdata/v4/invalid.tfstate")
	state, err := readState("terraform.tfstate", reader)
	if err == nil || state != nil {
		t.Errorf("ReadFile invalid state should return error")
	}
}

// Check that resource sources are properly set
func TestTerraformStateReader_Source(t *testing.T) {
	progress := &output.MockProgress{}
	progress.On("Inc").Return().Times(1)
	progress.On("Stop").Return().Times(1)

	version := "3.19.0"

	provider := mocks.NewMockedGoldenTFProvider("source", terraform.AWS, version, nil, false)
	library := terraform.NewProviderLibrary()
	library.AddProvider(terraform.AWS, provider)

	repo := testresource.InitFakeSchemaRepository(terraform.AWS, version)
	resourceaws.InitResourcesMetadata(repo)

	factory := dctlresource.NewDriftctlResourceFactory(repo)

	r := &TerraformStateReader{
		config: config.SupplierConfig{
			Key:  "tfstate",
			Path: path.Join(goldenfile.GoldenFilePath, "source", "terraform.tfstate"),
		},
		library:      library,
		progress:     progress,
		deserializer: resource.NewDeserializer(factory),
	}

	got, err := r.Resources()
	assert.Nil(t, err)
	assert.Len(t, got, 2)
	for _, res := range got {
		if res.ResourceType() == resourceaws.AwsS3BucketResourceType {
			assert.Equal(t, &resource.TerraformStateSource{
				State:  "tfstate://test/source/terraform.tfstate",
				Module: "",
				Name:   "bucket",
			}, res.Source)
		}
		if res.ResourceType() == resourceaws.AwsIamUserResourceType {
			assert.Equal(t, &resource.TerraformStateSource{
				State:  "tfstate://test/source/terraform.tfstate",
				Module: "module.iam_iam-user",
				Name:   "this_no_pgp",
			}, res.Source)
		}
	}
}

func TestTerraformStateReader_AWS_Resources(t *testing.T) {
	tests := []struct {
		name            string
		dirName         string
		wantErr         bool
		providerVersion string
	}{
		{name: "IAM user module", dirName: "module.iam_iam-user", wantErr: false},
		{name: "Data source", dirName: "data_source", wantErr: false},
		{name: "Route 53 zone", dirName: "aws_route53_zone", wantErr: false},
		{name: "Route 53 record - single record", dirName: "aws_route53_record", wantErr: false},
		{name: "Route 53 record - multiples zones, multiples records", dirName: "aws_route53_record_multiples", wantErr: false},
		{name: "Route 53 record - empty records", dirName: "aws_route53_record_null_records", wantErr: false},
		{name: "s3 full", dirName: "aws_s3_full", wantErr: false},
		{name: "s3 bucket public access block", dirName: "aws_s3_bucket_public_access_block", wantErr: false},
		{name: "s3 account public access block", dirName: "aws_s3_account_public_access_block", wantErr: false},
		{name: "RDS DB instance", dirName: "aws_db_instance", wantErr: false},
		{name: "RDS DB Subnet group", dirName: "aws_db_subnet_group", wantErr: false},
		{name: "Lambda function", dirName: "aws_lambda_function", wantErr: false},
		{name: "unsupported attribute", dirName: "unsupported_attribute", wantErr: false},
		{name: "Unsupported provider", dirName: "unsupported_provider", wantErr: false},
		{name: "Unsupported resource", dirName: "unsupported_resource", wantErr: false},
		{name: "EC2 instance", dirName: "aws_ec2_instance", wantErr: false},
		{name: "EC2 key pair", dirName: "aws_ec2_key_pair", wantErr: false},
		{name: "EC2 ami", dirName: "aws_ec2_ami", wantErr: false},
		{name: "EC2 eip", dirName: "aws_ec2_eip", wantErr: false},
		{name: "EC2 eip with its association", dirName: "aws_ec2_eip_association", wantErr: false},
		{name: "EC2 ebs volume", dirName: "aws_ec2_ebs_volume", wantErr: false},
		{name: "EC2 ebs snapshot", dirName: "aws_ec2_ebs_snapshot", wantErr: false},
		{name: "VPC security group", dirName: "aws_vpc_security_group", wantErr: false},
		{name: "IAM Users", dirName: "aws_iam_user_multiple", wantErr: false},
		{name: "IAM User Policy", dirName: "aws_iam_user_policy_multiple", wantErr: false},
		{name: "IAM access keys", dirName: "aws_iam_access_key_multiple", wantErr: false},
		{name: "IAM role", dirName: "aws_iam_role_multiple", wantErr: false},
		{name: "IAM policy", dirName: "aws_iam_policy_multiple", wantErr: false},
		{name: "IAM role policy", dirName: "aws_iam_role_policy_multiple", wantErr: false},
		{name: "IAM role policy attachment", dirName: "aws_iam_role_policy_attachment", wantErr: false},
		{name: "IAM user policy attachment", dirName: "aws_iam_user_policy_attachment", wantErr: false},
		{name: "IAM group policy", dirName: "aws_iam_group_policy", wantErr: false},
		{name: "IAM group policy attachment", dirName: "aws_iam_group_policy_attachment", wantErr: false},
		{name: "VPC security group rule", dirName: "aws_vpc_security_group_rule", wantErr: false},
		{name: "default route table", dirName: "aws_default_route_table", wantErr: false, providerVersion: "3.62.0"},
		{name: "route table", dirName: "aws_route_table", wantErr: false, providerVersion: "3.62.0"},
		{name: "route table associations", dirName: "aws_route_assoc", wantErr: false},
		{name: "route", dirName: "aws_route", wantErr: false},
		{name: "NAT gateway", dirName: "aws_nat_gateway", wantErr: false},
		{name: "Internet Gateway", dirName: "aws_internet_gateway", wantErr: false},
		{name: "SQS queue", dirName: "aws_sqs_queue", wantErr: false},
		{name: "SQS queue policy", dirName: "aws_sqs_queue_policy", wantErr: false},
		{name: "SNS Topic", dirName: "aws_sns_topic", wantErr: false},
		{name: "SNS Topic Policy", dirName: "aws_sns_topic_policy", wantErr: false},
		{name: "SNS Topic Subscription", dirName: "aws_sns_topic_subscription", wantErr: false},
		{name: "DynamoDB table", dirName: "aws_dynamodb_table", wantErr: false},
		{name: "Route53 Health Check", dirName: "aws_route53_health_check", wantErr: false},
		{name: "Cloudfront distribution", dirName: "aws_cloudfront_distribution", wantErr: false},
		{name: "ECR Repository", dirName: "aws_ecr_repository", wantErr: false},
		{name: "KMS key", dirName: "aws_kms_key", wantErr: false},
		{name: "KMS alias", dirName: "aws_kms_alias", wantErr: false},
		{name: "lambda event source mapping", dirName: "aws_lambda_event_source_mapping", wantErr: false},
		{name: "VPC", dirName: "aws_vpc", wantErr: false},
		{name: "Subnet", dirName: "aws_subnet", wantErr: false},
		{name: "RDS cluster", dirName: "aws_rds_cluster", wantErr: false},
		{name: "Cloudformation stack", dirName: "aws_cloudformation_stack", wantErr: false},
		{name: "Api Gateway Rest Api", dirName: "aws_api_gateway_rest_api", wantErr: false},
		{name: "Api Gateway Account", dirName: "aws_api_gateway_account", wantErr: false},
		{name: "Api Gateway Api Key", dirName: "aws_api_gateway_api_key", wantErr: false},
		{name: "Api Gateway authorizer", dirName: "aws_api_gateway_authorizer", wantErr: false},
		{name: "Api Gateway stage", dirName: "aws_api_gateway_stage", wantErr: false},
		{name: "Api Gateway resource", dirName: "aws_api_gateway_resource", wantErr: false},
		{name: "Api Gateway domain name", dirName: "aws_api_gateway_domain_name", wantErr: false},
		{name: "Api Gateway vpc link", dirName: "aws_api_gateway_vpc_link", wantErr: false},
		{name: "Api Gateway V2 Api", dirName: "aws_apigatewayv2_api", wantErr: false},
		{name: "Api Gateway V2 Route", dirName: "aws_apigatewayv2_route", wantErr: false},
		{name: "Api Gateway V2 Deployment", dirName: "aws_apigatewayv2_deployment", wantErr: false},
		{name: "Api Gateway V2 stage", dirName: "aws_apigatewayv2_stage", wantErr: false},
		{name: "Api Gateway request validator", dirName: "aws_api_gateway_request_validator", wantErr: false},
		{name: "Api Gateway rest api policy", dirName: "aws_api_gateway_rest_api_policy", wantErr: false},
		{name: "Api Gateway base path mapping", dirName: "aws_api_gateway_base_path_mapping", wantErr: false},
		{name: "Api Gateway method", dirName: "aws_api_gateway_method", wantErr: false},
		{name: "Api Gateway model", dirName: "aws_api_gateway_model", wantErr: false},
		{name: "Api Gateway method response", dirName: "aws_api_gateway_method_response", wantErr: false},
		{name: "Api Gateway gateway response", dirName: "aws_api_gateway_gateway_response", wantErr: false},
		{name: "Api Gateway method settings", dirName: "aws_api_gateway_method_settings", wantErr: false},
		{name: "Api Gateway integration", dirName: "aws_api_gateway_integration", wantErr: false},
		{name: "Api Gateway integration response", dirName: "aws_api_gateway_integration_response", wantErr: false},
		{name: "Api Gateway V2 Api", dirName: "aws_apigatewayv2_api", wantErr: false},
		{name: "Api Gateway V2 Route", dirName: "aws_apigatewayv2_route", wantErr: false},
		{name: "Api Gateway V2 authorizer", dirName: "aws_apigatewayv2_authorizer", wantErr: false},
		{name: "Api Gateway V2 integration", dirName: "aws_apigatewayv2_integration", wantErr: false},
		{name: "Api Gateway V2 model", dirName: "aws_apigatewayv2_model", wantErr: false},
		{name: "Api Gateway V2 stage", dirName: "aws_apigatewayv2_stage", wantErr: false},
		{name: "App gateway v2 vpc link", dirName: "aws_apigatewayv2_vpc_link", wantErr: false},
		{name: "App gateway v2 route response", dirName: "aws_apigatewayv2_route_response", wantErr: false},
		{name: "Api Gateway V2 mapping", dirName: "aws_apigatewayv2_api_mapping", wantErr: false},
		{name: "App gateway v2 domain name", dirName: "aws_apigatewayv2_domain_name", wantErr: false},
		{name: "Api Gateway V2 integration response", dirName: "aws_apigatewayv2_integration_response", wantErr: false},
		{name: "AppAutoScaling Targets", dirName: "aws_appautoscaling_target", wantErr: false},
		{name: "network acl", dirName: "aws_network_acl", wantErr: false},
		{name: "network acl rule", dirName: "aws_network_acl_rule", wantErr: false},
		{name: "default network acl", dirName: "aws_default_network_acl", wantErr: false},
		{name: "App autoscaling policy", dirName: "aws_appautoscaling_policy", wantErr: false},
		{name: "App autoscaling scheduled action", dirName: "aws_appautoscaling_scheduled_action", wantErr: false},
		{name: "Launch template", dirName: "aws_launch_template", wantErr: false},
		{name: "Launch configuration", dirName: "aws_launch_configuration", wantErr: false},
		{name: "EBS encryption by default", dirName: "aws_ebs_encryption_by_default", wantErr: false},
		{name: "LoadBalancer", dirName: "aws_lb", wantErr: false},
		{name: "Load balancer listener", dirName: "aws_lb_listener", wantErr: false},
		{name: "Classic load balancer", dirName: "aws_elb", wantErr: false},
		{name: "ElastiCache Cluster", dirName: "aws_elasticache_cluster", wantErr: false},
		{name: "IAM Group", dirName: "aws_iam_group", wantErr: false},
		{name: "ECR Repository Policy", dirName: "aws_ecr_repository_policy", wantErr: false},
		{name: "cloudtrail", dirName: "aws_cloudtrail", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &output.MockProgress{}
			progress.On("Inc").Return().Times(1)
			progress.On("Stop").Return().Times(1)

			shouldUpdate := tt.dirName == *goldenfile.Update

			var realProvider *aws.AWSTerraformProvider
			if tt.providerVersion == "" {
				tt.providerVersion = "3.19.0"
			}

			if shouldUpdate {
				var err error
				realProvider, err = aws.NewAWSTerraformProvider(tt.providerVersion, progress, os.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				err = realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
			}

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.AWS, tt.providerVersion, realProvider, shouldUpdate)
			library := terraform.NewProviderLibrary()
			library.AddProvider(terraform.AWS, provider)

			repo := testresource.InitFakeSchemaRepository(terraform.AWS, tt.providerVersion)
			resourceaws.InitResourcesMetadata(repo)

			factory := dctlresource.NewDriftctlResourceFactory(repo)

			r := &TerraformStateReader{
				config: config.SupplierConfig{
					Path: path.Join(goldenfile.GoldenFilePath, tt.dirName, "terraform.tfstate"),
				},
				library:      library,
				progress:     progress,
				deserializer: resource.NewDeserializer(factory),
			}

			got, err := r.Resources()
			resGoldenName := goldenfile.ResultsFilename
			if shouldUpdate {
				unm, err := json.Marshal(got)
				if err != nil {
					panic(err)
				}
				goldenfile.WriteFile(tt.dirName, unm, resGoldenName)
			}

			file := goldenfile.ReadFile(tt.dirName, resGoldenName)
			var want []interface{}
			if err := json.Unmarshal(file, &want); err != nil {
				panic(err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotc := convert(got)
			changelog, err := diff.Diff(gotc, want)
			if err != nil {
				panic(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), change.From, change.To)
				}
			}
		})
	}
}

func TestTerraformStateReader_Github_Resources(t *testing.T) {
	tests := []struct {
		name    string
		dirName string
		wantErr bool
	}{
		{name: "github repository", dirName: "github_repository", wantErr: false},
		{name: "github team", dirName: "github_team", wantErr: false},
		{name: "github membership", dirName: "github_membership", wantErr: false},
		{name: "github team membership", dirName: "github_team_membership", wantErr: false},
		{name: "github branch protection", dirName: "github_branch_protection", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &output.MockProgress{}
			progress.On("Inc").Return().Times(1)
			progress.On("Stop").Return().Times(1)

			shouldUpdate := tt.dirName == *goldenfile.Update

			var realProvider *github.GithubTerraformProvider

			if shouldUpdate {
				var err error
				realProvider, err = github.NewGithubTerraformProvider("", progress, os.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				err = realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
			}

			version := "4.4.0"

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.GITHUB, version, realProvider, shouldUpdate)
			library := terraform.NewProviderLibrary()
			library.AddProvider(terraform.GITHUB, provider)

			repo := testresource.InitFakeSchemaRepository(terraform.GITHUB, version)
			resourcegithub.InitResourcesMetadata(repo)
			factory := dctlresource.NewDriftctlResourceFactory(repo)

			r := &TerraformStateReader{
				config: config.SupplierConfig{
					Path: path.Join(goldenfile.GoldenFilePath, tt.dirName, "terraform.tfstate"),
				},
				library:      library,
				progress:     progress,
				deserializer: resource.NewDeserializer(factory),
			}

			got, err := r.Resources()
			resGoldenName := goldenfile.ResultsFilename
			if shouldUpdate {
				unm, err := json.Marshal(got)
				if err != nil {
					panic(err)
				}
				goldenfile.WriteFile(tt.dirName, unm, resGoldenName)
			}

			file := goldenfile.ReadFile(tt.dirName, resGoldenName)
			var want []interface{}
			if err := json.Unmarshal(file, &want); err != nil {
				panic(err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			changelog, err := diff.Diff(convert(got), want)
			if err != nil {
				panic(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), change.From, change.To)
				}
			}
		})
	}
}

func TestTerraformStateReader_Google_Resources(t *testing.T) {
	tests := []struct {
		name    string
		dirName string
		wantErr bool
	}{
		{name: "compute firewall", dirName: "google_compute_firewall", wantErr: false},
		{name: "compute router", dirName: "google_compute_router", wantErr: false},
		{name: "compute instance", dirName: "google_compute_instance", wantErr: false},
		{name: "Bucket IAM Bindings", dirName: "google_bucket_iam_binding", wantErr: false},
		{name: "Bucket IAM members", dirName: "google_bucket_iam_member", wantErr: false},
		{name: "Bucket IAM Policy", dirName: "google_bucket_iam_policy", wantErr: false},
		{name: "DNS managed zone", dirName: "google_dns_managed_zone", wantErr: false},
		{name: "bigquery dataset", dirName: "google_bigquery_dataset", wantErr: false},
		{name: "bigquery table", dirName: "google_bigquery_table", wantErr: false},
		{name: "compute address", dirName: "google_compute_address", wantErr: false},
		{name: "compute global address", dirName: "google_compute_global_address", wantErr: false},
		{name: "cloudfunctions function", dirName: "google_cloudfunctions_function", wantErr: false},
		{name: "compute subnetwork", dirName: "google_compute_subnetwork", wantErr: false},
		{name: "compute disk", dirName: "google_compute_disk", wantErr: false},
		{name: "compute image", dirName: "google_compute_image", wantErr: false},
		{name: "bigtable instance", dirName: "google_bigtable_instance", wantErr: false},
		{name: "bigtable table", dirName: "google_bigtable_table", wantErr: false},
		{name: "sql database instance", dirName: "google_sql_database_instance", wantErr: false},
		{name: "health check", dirName: "google_compute_health_check", wantErr: false},
		{name: "cloudrun service", dirName: "google_cloudrun_service", wantErr: false},
		{name: "compute node group", dirName: "google_compute_node_group", wantErr: false},
		{name: "compute forwarding rule", dirName: "google_compute_forwarding_rule", wantErr: false},
		{name: "compute instance group manager", dirName: "google_compute_instance_group_manager", wantErr: false},
		{name: "compute global forwarding rule", dirName: "google_compute_global_forwarding_rule", wantErr: false},
		{name: "compute ssl certificate", dirName: "google_compute_ssl_certificate", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &output.MockProgress{}
			progress.On("Inc").Return().Times(1)
			progress.On("Stop").Return().Times(1)

			shouldUpdate := tt.dirName == *goldenfile.Update

			var realProvider *google.GCPTerraformProvider
			providerVersion := "3.78.0"
			var err error
			realProvider, err = google.NewGCPTerraformProvider(providerVersion, progress, os.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)

			if shouldUpdate {
				err = realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
			}

			library := terraform.NewProviderLibrary()
			library.AddProvider(terraform.GOOGLE, provider)

			repo := testresource.InitFakeSchemaRepository(terraform.GOOGLE, providerVersion)
			resourcegoogle.InitResourcesMetadata(repo)
			factory := dctlresource.NewDriftctlResourceFactory(repo)

			r := &TerraformStateReader{
				config: config.SupplierConfig{
					Path: path.Join(goldenfile.GoldenFilePath, tt.dirName, "terraform.tfstate"),
				},
				library:      library,
				progress:     progress,
				deserializer: resource.NewDeserializer(factory),
			}

			got, err := r.Resources()
			resGoldenName := goldenfile.ResultsFilename
			if shouldUpdate {
				unm, err := json.Marshal(got)
				if err != nil {
					panic(err)
				}
				goldenfile.WriteFile(tt.dirName, unm, resGoldenName)
			}

			file := goldenfile.ReadFile(tt.dirName, resGoldenName)
			var want []interface{}
			if err := json.Unmarshal(file, &want); err != nil {
				panic(err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			changelog, err := diff.Diff(convert(got), want)
			if err != nil {
				panic(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), change.From, change.To)
				}
			}
		})
	}
}

func TestTerraformStateReader_Azure_Resources(t *testing.T) {
	tests := []struct {
		name    string
		dirName string
		wantErr bool
	}{
		{name: "network security group", dirName: "azurerm_network_security_group", wantErr: false},
		{name: "load balancers", dirName: "azurerm_lb", wantErr: false},
		{name: "private dns zone", dirName: "azurerm_private_dns_zone", wantErr: false},
		{name: "private dns a record", dirName: "azurerm_private_dns_a_record", wantErr: false},
		{name: "private dns aaaa record", dirName: "azurerm_private_dns_aaaa_record", wantErr: false},
		{name: "private dns cname record", dirName: "azurerm_private_dns_cname_record", wantErr: false},
		{name: "private dns ptr record", dirName: "azurerm_private_dns_ptr_record", wantErr: false},
		{name: "private dns mx record", dirName: "azurerm_private_dns_mx_record", wantErr: false},
		{name: "private dns srv record", dirName: "azurerm_private_dns_srv_record", wantErr: false},
		{name: "private dns txt record", dirName: "azurerm_private_dns_txt_record", wantErr: false},
		{name: "images", dirName: "azurerm_image", wantErr: false},
		{name: "ssh public key", dirName: "azurerm_ssh_public_key", wantErr: false},
		{name: "load balancer rule", dirName: "azurerm_lb_rule", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &output.MockProgress{}
			progress.On("Inc").Return().Times(1)
			progress.On("Stop").Return().Times(1)

			shouldUpdate := tt.dirName == *goldenfile.Update

			var realProvider *azurerm.AzureTerraformProvider
			providerVersion := "2.71.0"
			var err error
			realProvider, err = azurerm.NewAzureTerraformProvider(providerVersion, progress, os.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)

			if shouldUpdate {
				err = realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
			}

			library := terraform.NewProviderLibrary()
			library.AddProvider(terraform.AZURE, provider)

			repo := testresource.InitFakeSchemaRepository(terraform.AZURE, providerVersion)
			resourceazure.InitResourcesMetadata(repo)
			factory := dctlresource.NewDriftctlResourceFactory(repo)

			r := &TerraformStateReader{
				config: config.SupplierConfig{
					Path: path.Join(goldenfile.GoldenFilePath, tt.dirName, "terraform.tfstate"),
				},
				library:      library,
				progress:     progress,
				deserializer: resource.NewDeserializer(factory),
			}

			got, err := r.Resources()
			resGoldenName := goldenfile.ResultsFilename
			if shouldUpdate {
				unm, err := json.Marshal(got)
				if err != nil {
					panic(err)
				}
				goldenfile.WriteFile(tt.dirName, unm, resGoldenName)
			}

			file := goldenfile.ReadFile(tt.dirName, resGoldenName)
			var want []interface{}
			if err := json.Unmarshal(file, &want); err != nil {
				panic(err)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			changelog, err := diff.Diff(convert(got), want)
			if err != nil {
				panic(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), change.From, change.To)
				}
			}
		})
	}
}

func convert(got []*resource.Resource) []interface{} {
	unm, err := json.Marshal(got)
	if err != nil {
		panic(err)
	}
	var want []interface{}
	if err := json.Unmarshal(unm, &want); err != nil {
		panic(err)
	}
	return want
}

func TestTerraformStateReader_VersionSupported(t *testing.T) {
	tests := []struct {
		name      string
		statePath string
		err       error
	}{
		{
			name:      "should detect unsupported version",
			statePath: "testdata/v4/unsupported_version.tfstate",
			err:       errors.New("terraform.tfstate was generated using Terraform 0.10.26 which is currently not supported by driftctl. Please read documentation at https://docs.driftctl.com/limitations"),
		},
		{
			name:      "should detect supported version",
			statePath: "testdata/v4/supported_version.tfstate",
			err:       nil,
		},
		{
			name:      "should return invalid version error",
			statePath: "testdata/v4/invalid_version.tfstate",
			err:       errors.New("Invalid Terraform version string: State file claims to have been written by Terraform version \"invalid\", which is not a valid version string."),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader, err := os.Open(test.statePath)
			assert.NoError(t, err)

			_, err = readState("terraform.tfstate", reader)
			if test.err != nil {
				assert.EqualError(t, err, test.err.Error())
			} else {
				assert.Equal(t, test.err, err)
			}
		})
	}
}

func TestTerraformStateReader_WithIgnoredResource(t *testing.T) {
	progress := &output.MockProgress{}
	progress.On("Inc").Return().Times(1)
	progress.On("Stop").Return().Times(1)

	provider := mocks.NewMockedGoldenTFProvider("ignored_resources", terraform.AWS, "3.19.0", nil, false)
	library := terraform.NewProviderLibrary()
	library.AddProvider(terraform.AWS, provider)

	filter := &filter.MockFilter{}
	filter.On("IsTypeIgnored", resource.ResourceType("aws_s3_bucket")).Return(true)

	r := &TerraformStateReader{
		config: config.SupplierConfig{
			Path: path.Join(goldenfile.GoldenFilePath, "ignored_resources", "terraform.tfstate"),
		},
		library:  library,
		progress: progress,
		filter:   filter,
	}

	got, err := r.Resources()
	filter.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Len(t, got, 0)
}
