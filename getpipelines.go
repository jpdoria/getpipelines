package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
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
func exportToCSV(data [][]string) {
	file, err := os.OpenFile("results.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
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

	pipelines := [][]string{}

	// Then retrieve each pipeline info.
	for i := 0; i < len(listRes.Pipelines); i++ {
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

	exportToCSV(pipelines)
}

// Main function.
func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	headers := [][]string{{"PipelineName", "S3Bucket", "S3ObjectKey"}}
	exportToCSV(headers)

	l := parseConfJSON()

	for i := 0; i < len(l.Roles); i++ {
		roleArn := l.Roles[i].RoleArn
		region := l.Roles[i].Region
		sess, creds := newSess(roleArn, region)
		cp := codepipeline.New(sess, &aws.Config{Credentials: creds})

		getActivePipelinesInfo(cp)
	}
}
