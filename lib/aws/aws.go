package aws

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/iam"
)

// We seperate the AWS client so that we
// can mock it in the unit testing also
// reduce the overhead of creating it every time
var (
	AwsSess = session.Must(
		session.NewSession(
			&aws.Config{
				Credentials: credentials.NewEnvCredentials(),
				Region:      aws.String(os.Getenv("AWS_DEFAULT_REGION")),
			},
		),
	)
)

// Retrieving ECR auths
func GetEcrAuths() ([]*ecr.AuthorizationData, error) {
	svc := ecr.New(AwsSess)

	result, err := svc.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		handleErr(err)
		return nil, err
	}

	return result.AuthorizationData, nil
}

// Get IAM user details
func GetIamUser() (*iam.User, error) {
	svc := iam.New(AwsSess)

	input := &iam.GetUserInput{
		UserName: aws.String("Bob"),
	}

	result, err := svc.GetUser(input)
	if err != nil {
		handleErr(err)
		return nil, err
	}

	return result.User, nil
}

// Handling AWS errors
func handleErr(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case ecr.ErrCodeServerException:
			log.Println(ecr.ErrCodeServerException, aerr.Error())
		case ecr.ErrCodeInvalidParameterException:
			log.Println(ecr.ErrCodeInvalidParameterException, aerr.Error())
		default:
			log.Println(aerr.Error())
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Println(err.Error())
	}
}
