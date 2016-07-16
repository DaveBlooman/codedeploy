package command

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/DaveBlooman/codedeploy/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

// CmdDeploy Deploys app
func CmdDeploy(c *cli.Context) {
	flags := fetchConfigGetFlags(c)
	checkOptionFlags(flags)
	banner("deploying", flags)

	file, err := os.Open(flags["filename"])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()

	buffer := make([]byte, size)

	file.Read(buffer)

	fileBytes := bytes.NewReader(buffer)

	fileType := "zip"
	path := flags["filename"]

	var config *aws.Config

	if c.String("awsprofile") == "" {
		config = &aws.Config{
			Region: aws.String(flags["region"]),
		}
	} else {
		stscreds := sts.New(session.New(&aws.Config{Region: aws.String(flags["region"])}))

		stsparams := &sts.AssumeRoleInput{
			RoleArn:         aws.String(c.String("awsprofile")),
			RoleSessionName: aws.String("roleSessionName"),
		}

		stsresp, err := stscreds.AssumeRole(stsparams)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		config = &aws.Config{
			Region:      aws.String(flags["region"]),
			Credentials: credentials.NewStaticCredentials(*stsresp.Credentials.AccessKeyId, *stsresp.Credentials.SecretAccessKey, *stsresp.Credentials.SessionToken),
		}
	}

	s3config := config

	_, putErr := storage.Put(flags["region"], flags["bucket"], path, fileType, fileBytes, size, s3config)

	if putErr != nil {
		fmt.Println("s3 put error")
		fmt.Println(putErr.Error())
		return
	}

	s3svc := s3.New(session.New(config))

	s3params := &s3.GetObjectInput{
		Bucket: aws.String(flags["bucket"]),
		Key:    aws.String(flags["filename"]),
	}

	s3resp, err := s3svc.GetObject(s3params)

	if err != nil {
		fmt.Println("s3 get error")
		fmt.Println(err.Error())
		return
	}

	etag := *s3resp.ETag

	svc := codedeploy.New(session.New(config))

	params := &codedeploy.CreateDeploymentInput{
		ApplicationName:               aws.String(flags["app-name"]), // Required
		DeploymentConfigName:          aws.String("CodeDeployDefault.OneAtATime"),
		DeploymentGroupName:           aws.String(flags["deployment-group"]),
		Description:                   aws.String("Testing"),
		IgnoreApplicationStopFailures: aws.Bool(true),
		Revision: &codedeploy.RevisionLocation{
			RevisionType: aws.String("S3"),
			S3Location: &codedeploy.S3Location{
				Bucket:     aws.String(flags["bucket"]),
				BundleType: aws.String("zip"),
				ETag:       aws.String(etag),
				Key:        aws.String(flags["filename"]),
			},
		},
	}
	cdresp, err := svc.CreateDeployment(params)

	if err != nil {
		fmt.Println("codedeploy error")
		fmt.Println(err.Error())
		return
	}

	fmt.Println(cdresp)
}

func fetchConfigGetFlags(c *cli.Context) map[string]string {
	return map[string]string{
		"region":           c.String("region"),
		"bucket":           c.String("bucket"),
		"filename":         c.String("filename"),
		"deployment-group": c.String("deployment-group"),
		"app-name":         c.String("app-name"),
	}
}

func checkOptionFlags(flags map[string]string) {
	for k, v := range flags {
		if len(v) == 0 {
			msg := fmt.Sprintf("Error: '%s' is a required flag", k)
			outputError(msg)
		}
	}
}

func banner(name string, flags map[string]string) {
	fmt.Printf("%s:\t\"%s\"\n", changeColor("Command", color.FgBlue), name)
	for key, val := range flags {
		if len(key) > 5 {
			fmt.Printf("%s:\t\"%s\"\n", changeColor(strings.Title(key), color.FgBlue), val)
		} else {
			fmt.Printf("%s:\t\t\"%s\"\n", changeColor(strings.Title(key), color.FgBlue), val)
		}
	}
}

func outputError(text string) {
	msg := changeColor(text, color.FgRed)
	fmt.Println(msg)
	os.Exit(1)
}

func changeColor(text string, code color.Attribute) string {
	c := color.New(code).SprintFunc()
	return c(text)
}
