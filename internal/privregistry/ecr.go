package privregistry

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/pkg/errors"
)

func GetECRToken(registry string, accessKeyID string, secretAccessKey string, region string) (string, error) {
	registryURLParts := strings.Split(registry, ".")
	if len(registryURLParts) < 1 {
		return "", errors.Errorf("fail to extract registryID from registry %s", registry)
	}
	registryID := registryURLParts[0]

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return "", errors.Wrap(err, "fail to create AWS session")
	}

	ecrClient := ecr.New(sess)
	input := &ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{
			aws.String(registryID),
		},
	}

	result, err := ecrClient.GetAuthorizationToken(input)
	if err != nil {
		return "", errors.Wrap(err, "fail to get ECR authorization token")
	}

	if len(result.AuthorizationData) == 0 {
		return "", errors.New("returned authorization data array is empty")
	}

	authorizationData := result.AuthorizationData[0]
	if authorizationData.AuthorizationToken == nil {
		return "", errors.New("returned authorization data token is nil")
	}

	token := *authorizationData.AuthorizationToken
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", errors.Wrap(err, "fail to decode authorization token")
	}

	return string(decodedToken), nil
}
