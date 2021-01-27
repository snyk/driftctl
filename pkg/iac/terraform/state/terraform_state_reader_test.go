package state

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/iac"
	"github.com/cloudskiff/driftctl/pkg/iac/terraform/state/backend"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/r3labs/diff/v2"
)

func TestReadStateValid(t *testing.T) {
	reader, _ := os.Open("testdata/v4/valid.tfstate")
	_, err := readState(reader)
	if err != nil {
		t.Errorf("Unable to read state, %s", err)
		return
	}
}

func TestReadStateInvalid(t *testing.T) {
	reader, _ := os.Open("testdata/v4/invalid.tfstate")
	state, err := readState(reader)
	if err == nil || state != nil {
		t.Errorf("ReadFile invalid state should return error")
	}
}

func TestTerraformStateReader_Resources(t *testing.T) {
	tests := []struct {
		name    string
		dirName string
		wantErr bool
	}{
		{name: "IAM user module", dirName: "module.iam_iam-user", wantErr: false},
		{name: "Data source", dirName: "data_source", wantErr: false},
		{name: "Route 53 zone", dirName: "route53_zone", wantErr: false},
		{name: "Route 53 record - single record", dirName: "route53_record", wantErr: false},
		{name: "Route 53 record - multiples zones, multiples records", dirName: "route53_record_multiples", wantErr: false},
		{name: "s3 full", dirName: "s3_full", wantErr: false},
		{name: "RDS DB instance", dirName: "db_instance", wantErr: false},
		{name: "RDS DB Subnet group", dirName: "db_subnet_group", wantErr: false},
		{name: "Lambda function", dirName: "lambda_function", wantErr: false},
		{name: "unsupported attribute", dirName: "unsupported_attribute", wantErr: false},
		{name: "Unsupported provider", dirName: "unsupported_provider", wantErr: false},
		{name: "EC2 instance", dirName: "ec2_instance", wantErr: false},
		{name: "EC2 key pair", dirName: "ec2_key_pair", wantErr: false},
		{name: "EC2 ami", dirName: "ec2_ami", wantErr: false},
		{name: "EC2 eip", dirName: "ec2_eip", wantErr: false},
		{name: "EC2 eip with its association", dirName: "ec2_eip_association", wantErr: false},
		{name: "EC2 ebs volume", dirName: "ec2_ebs_volume", wantErr: false},
		{name: "EC2 ebs snapshot", dirName: "ec2_ebs_snapshot", wantErr: false},
		{name: "VPC security group", dirName: "vpc_security_group", wantErr: false},
		{name: "IAM Users", dirName: "iam_user_multiple", wantErr: false},
		{name: "IAM User Policy", dirName: "iam_user_policy_multiple", wantErr: false},
		{name: "IAM access keys", dirName: "iam_access_key_multiple", wantErr: false},
		{name: "IAM role", dirName: "iam_role_multiple", wantErr: false},
		{name: "IAM policy", dirName: "iam_policy_multiple", wantErr: false},
		{name: "IAM role policy", dirName: "iam_role_policy_multiple", wantErr: false},
		{name: "IAM role policy attachment", dirName: "iam_role_policy_attachment", wantErr: false},
		{name: "IAM user policy attachment", dirName: "iam_user_policy_attachment", wantErr: false},
		{name: "VPC security group rule", dirName: "vpc_security_group_rule", wantErr: false},
		{name: "route table", dirName: "route_table", wantErr: false},
		{name: "route", dirName: "route", wantErr: false},
		{name: "route table associations", dirName: "route_assoc", wantErr: false},
		{name: "NAT gateway", dirName: "aws_nat_gateway", wantErr: false},
		{name: "Internet Gateway", dirName: "internet_gateway", wantErr: false},
		{name: "SQS queue", dirName: "sqs_queue", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldUpdate := tt.dirName == *goldenfile.Update

			var realProvider terraform.TerraformProvider

			if shouldUpdate {
				var err error
				realProvider, err = aws.NewTerraFormProvider()
				if err != nil {
					t.Fatal(err)
				}
			}

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, realProvider, shouldUpdate)
			library := terraform.NewProviderLibrary()
			library.AddProvider(terraform.AWS, provider)

			b, _ := backend.NewFileReader(path.Join(goldenfile.GoldenFilePath, tt.dirName, "terraform.tfstate"))
			r := &TerraformStateReader{
				backend:       b,
				library:       library,
				deserializers: iac.Deserializers(),
			}

			got, err := r.Resources()
			resGoldenName := "result.golden.json"
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

func convert(got []resource.Resource) []interface{} {
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
