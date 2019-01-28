# About

A simple script written in Go that retrieves the source and key of active pipelines that use S3 on CodePipeline.

# Installing

```bash
go get -u github.com/jpdoria/getpipelines
```

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
getpipelines.go -config config.json -destdir /tmp
```

# Example

```bash
➜ getcicdinfo.go -config config.json -destdir /tmp
2019/01/28 20:58:18.871780 getcicdinfo.go:217: [info] getting all the information you need
2019/01/28 20:58:55.376848 getcicdinfo.go:108: [warn] pipeline-d is not using S3 as source
2019/01/28 20:58:56.220959 getcicdinfo.go:108: [warn] pipeline-e is not using S3 as source
2019/01/28 20:59:22.152755 getcicdinfo.go:230: [info] saved to /tmp/getActivePipelinesInfoResults.csv
2019/01/28 20:59:22.152785 getcicdinfo.go:231: [info] saved to /tmp/getApprovalLogsInfoResults.csv
```

```bash
➜ cat /tmp/getActivePipelinesInfoResults.csv
PipelineName,S3Bucket,S3ObjectKey
pipeline-a,bucket-a,owner-repo-dev-code.zip
pipeline-b,bucket-b,owner-repo-uat-code.zip
pipeline-c,bucket-c,owner-repo-prd-code.zip
```

```bash
➜ cat /tmp/getApprovalLogsInfoResults.csv
UserIdentity,AwsRegion,SourceIPAddress,Status,Summary,StageName,PipelineName,ApprovedAt,RequestId,EventId
arn:aws:sts::000000000000:assumed-role/aws-a-role/iam.user,us-east-1,1.2.3.4,Approved,Code from master branch,Approval,pipeline-a,"Jan 01, 2019 00:00:00 AM",f8e4e2d3-4b09-467f-b797-d8120a34aeb6,0a622855-835f-4031-9cd9-aadb18c1b533
arn:aws:sts::111111111111:assumed-role/aws-b-role/iam.user,ap-southeast-1,1.2.3.4,Approved,Code from master branch,Approval,pipeline-b,"Jan 01, 2019 00:00:00 AM",2a7964fb-2788-48d8-bddc-243a1ababbf2,51c8805d-0201-45ce-b3c9-c5ce59179c10
arn:aws:sts::222222222222:assumed-role/aws-c-role/iam.user,eu-west-1,1.2.3.4,Approved,Code from master branch,Approval,pipeline-c,"Jan 01, 2019 00:00:00 AM",463324df-5f34-4f96-a314-f0ed5c32891e,d5de5dee-9cce-4904-8cb3-2691062788a7
```
