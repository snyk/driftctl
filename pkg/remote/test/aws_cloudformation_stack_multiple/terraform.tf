provider "aws" {
  region  = "us-east-1"
  profile = "cloudskiff"
}

resource "aws_cloudformation_stack" "foo" {
  name = "foo-stack"

  parameters = {
    VPCCidr = "10.0.0.0/16"
  }

  template_body = <<STACK
{
  "Parameters" : {
    "VPCCidr" : {
      "Type" : "String",
      "Default" : "10.0.0.0/16",
      "Description" : "Enter the CIDR block for the VPC. Default is 10.0.0.0/16."
    }
  },
  "Resources" : {
    "myVpc": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : { "Ref" : "VPCCidr" },
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  }
}
STACK
}

resource "aws_cloudformation_stack" "bar" {
  name = "bar-stack"

  capabilities = [ "CAPABILITY_NAMED_IAM" ]

  template_body = file("./iam.yml")
}