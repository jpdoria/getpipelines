# About

A simple script written in Go that retrieves the source and key of active pipelines that use S3 on CodePipeline.

# Usage

Update `config.json` with the correct region and IAM roles to assume.

```json
{
    "Roles": [
        {
            "RoleArn": "arn:aws:iam::000000000000:role/aws-a-role",
            "Region": "us-east-1"
        },
        {
            "RoleArn": "arn:aws:iam::111111111111:role/aws-b-role",
            "Region": "ap-southeast-1"
        },
        {
            "RoleArn": "arn:aws:iam::222222222222:role/aws-c-role",
            "Region": "eu-west-1"
        }
    ]
}
```

Run the script.

```bash
go run getpipelines.go
```

# Example

```bash
➜ go run getpipelines.go
2019/01/28 01:20:28.028246 [warn] pipeline-d is not using S3 as source
```

```bash
➜ cat results.csv
PipelineName,S3Bucket,S3ObjectKey
pipeline-a,bucket-a,owner-repo-dev-code.zip
pipeline-b,bucket-b,owner-repo-uat-code.zip
pipeline-c,bucket-c,owner-repo-prd-code.zip
```