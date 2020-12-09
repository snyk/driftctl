# Output format

Driftctl supports multiple kinds of output formats and by default uses the standard output (console).

## Console

Environment: `DCTL_OUTPUT`

### Usage 

```
$ driftctl scan
$ driftctl scan --output console://
$ DCTL_OUTPUT=console:// driftctl scan
```

### Structure

```
Found deleted resources:
  aws_s3_bucket:
    - driftctl-bucket-test-2
Found unmanaged resources:
  aws_s3_bucket:
    - driftctl-bucket-test-3
Found drifted resources:
  - driftctl-bucket-test-1 (aws_s3_bucket):
    ~ Versioning.0.Enabled: false => true
Found 3 resource(s)
 - 33% coverage
 - 1 covered by IaC
 - 1 not covered by IaC
 - 1 deleted on cloud provider
 - 1/1 drifted from IaC
```

## JSON

### Usage

```
$ driftctl scan --output json:///tmp/result.json # Will output results to /tmp/result.json
$ driftctl scan --output json://result.json # Will output results to ./result.json
$ DCTL_OUTPUT=json://result.json driftctl scan
```

### Structure

```json5
{
	"summary": {
		"total_resources": 3,
		"total_drifted": 1,
		"total_unmanaged": 1,
		"total_deleted": 1,
		"total_managed": 1
	},
	"managed": [ // list of resources found in IaC and in sync with remote 
		{
			"id": "driftctl-bucket-test-1",
			"type": "aws_s3_bucket"
		}
	],
	"unmanaged": [ // list of resources found in remote but not in IaC
		{
			"id": "driftctl-bucket-test-3",
			"type": "aws_s3_bucket"
		}
	],
	"deleted": [ // list of resources found in IaC but not on remote
		{
			"id": "driftctl-bucket-test-2",
			"type": "aws_s3_bucket"
		}
	],
	"differences": [ // A list of changes on managed resources
		{
			"res": {
				"id": "driftctl-bucket-test-1",
				"type": "aws_s3_bucket"
			},
			"changelog": [
				{
					"type": "update", // Kind of change, could be one of update, create, delete
					"path": [ // Path of the change, sorted from root to leaf
						"Versioning",
						"0",
						"Enabled"
					],
					"from": false, // Mixed type
					"to": true // Mixed type
				}
			]
		}
	],
	"coverage": 33
}
```