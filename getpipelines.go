package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

// Roles struct which contains the roles.
// This is required for reading the config.json file.
type Roles struct {
	Roles []Role `json:"Roles"`
}

// Role struct which contains the role ARNs and regions.
// This is also required for reading the config.json file.
type Role struct {
	RoleArn string `json:"RoleArn"`
	Region  string `json:"Region"`
}

// Checker.
func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// Get the role ARNs and regions from the config.json
func parseConfJSON() Roles {
	data, err := ioutil.ReadFile("config.json")
	var r Roles

	check(err)
	json.Unmarshal(data, &r)
	return r
}

// Create a new session by assuming an IAM role. If IAM role doesn't exist,
// create a regular session by using the local AWS credentials.
func newSess(roleArn, region string) (*session.Session, *credentials.Credentials) {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))

	if roleArn != "" {
		creds := stscreds.NewCredentials(sess, roleArn)
		return sess, creds
	}

	return sess, nil
}

// Export data (slices) to CSV file.
func exportToCSV(data [][]string, output string) {
	file, err := os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	check(err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, val := range data {
		err = writer.Write(val)
		check(err)
	}
}

// Get active pipelines information like PipelineName, S3Bucket, and S3ObjectKey.
func getActivePipelinesInfo(cp *codepipeline.CodePipeline) {
	// List all pipelines first.
	listIn := &codepipeline.ListPipelinesInput{}
	listRes, listErr := cp.ListPipelines(listIn)
	check(listErr)

	// Then get each pipeline info.
	log.Println("getting pipelines information")
	pipelines := [][]string{}

	for i := range listRes.Pipelines {
		pipelineName := *listRes.Pipelines[i].Name
		getIn := &codepipeline.GetPipelineInput{Name: aws.String(pipelineName)}
		getRes, getErr := cp.GetPipeline(getIn)
		check(getErr)

		stageConfig := getRes.Pipeline.Stages[0].Actions[0].Configuration

		for k := range stageConfig {
			if k == "S3Bucket" || k == "S3ObjectKey" {
				bucket := *stageConfig["S3Bucket"]
				key := *stageConfig["S3ObjectKey"]
				row := []string{pipelineName, bucket, key}
				pipelines = append(pipelines, row)
			} else {
				log.Printf("[warn] %v is not using S3 as source\n", pipelineName)
			}

			break
		}
	}

	exportToCSV(pipelines, "getActivePipelinesInfoResults.csv")
	log.Println("saved to getActivePipelinesInfoResults.csv")
}

// Get approval logs from CloudTrail.
func getApprovalLogsInfo(ct *cloudtrail.CloudTrail) {
	previousMonth := time.Now().AddDate(0, -1, 0)
	currentDate := time.Now()
	in := &cloudtrail.LookupEventsInput{
		StartTime: aws.Time(previousMonth),
		EndTime:   aws.Time(currentDate),
		LookupAttributes: []*cloudtrail.LookupAttribute{
			{
				AttributeKey:   aws.String("EventName"),
				AttributeValue: aws.String("PutApprovalResult"),
			},
		},
	}
	res, err := ct.LookupEvents(in)
	check(err)

	// Get approval logs.
	log.Println("getting approval logs")
	data := [][]string{}

	for i := range res.Events {
		// Unmarshal JSON output from CloudTrailEvent.
		event := *res.Events[i].CloudTrailEvent
		var e map[string]interface{}
		json.Unmarshal([]byte(event), &e)

		userIdentity := e["userIdentity"].(map[string]interface{})
		requestParameters := e["requestParameters"].(map[string]interface{})
		result := requestParameters["result"].(map[string]interface{})
		responseElements := e["responseElements"].(map[string]interface{})

		// Things we need to export to CSV file.
		arn := userIdentity["arn"]
		awsRegion := e["awsRegion"]
		souceIPAddress := e["sourceIPAddress"]
		status := result["status"]
		summary := result["summary"]
		stageName := requestParameters["stageName"]
		pipelineName := *res.Events[i].Resources[0].ResourceName
		approvedAt := responseElements["approvedAt"]
		requestID := e["requestID"]
		eventID := e["eventID"]
		row := []string{
			arn.(string),
			awsRegion.(string),
			souceIPAddress.(string),
			status.(string),
			summary.(string),
			stageName.(string),
			pipelineName,
			approvedAt.(string),
			requestID.(string),
			eventID.(string),
		}
		data = append(data, row)
	}

	exportToCSV(data, "getApprovalLogsInfoResults.csv")
	log.Println("saved to getApprovalLogsInfoResults.csv")
}

// Main function.
func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// Write headers for getActivePipelinesInfo.
	headers := [][]string{{"PipelineName", "S3Bucket", "S3ObjectKey"}}
	exportToCSV(headers, "getActivePipelinesInfoResults.csv")

	// Write headers for getApprovalLogsInfo.
	headers = [][]string{{
		"UserIdentity",
		"AwsRegion",
		"SourceIPAddress",
		"Status",
		"Summary",
		"StageName",
		"PipelineName",
		"ApprovedAt",
		"RequestId",
		"EventId",
	}}
	exportToCSV(headers, "getApprovalLogsInfoResults.csv")

	l := parseConfJSON()

	for i := range l.Roles {
		roleArn := l.Roles[i].RoleArn
		region := l.Roles[i].Region
		sess, creds := newSess(roleArn, region)
		cp := codepipeline.New(sess, &aws.Config{Credentials: creds})
		ct := cloudtrail.New(sess, &aws.Config{Credentials: creds})

		getActivePipelinesInfo(cp)
		getApprovalLogsInfo(ct)
	}
}
