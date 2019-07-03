package source

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/pkg/errors"
)

const (
	// Keys for the map returned by Read()
	AwsAccessKeyID     string = "accessKeyId"
	AwsSecretAccessKey string = "secretAccessKey"

	// Default values
	DefaultMaxAge time.Duration = 100 * time.Minute
)

type AwsIamSource struct {
	UserName string         `yaml:"username"`
	RoleArn  string         `yaml:"role_arn"`
	Client   *cziAws.Client `yaml:"client"`
	MaxAge   time.Duration  `yaml:"max_age"`
}

func NewAwsIamSource() *AwsIamSource {
	return &AwsIamSource{
		MaxAge: DefaultMaxAge,
	}
}

func (src *AwsIamSource) WithUserName(userName string) *AwsIamSource {
	src.UserName = userName
	return src
}

func (src *AwsIamSource) WithRoleArn(roleArn string) *AwsIamSource {
	src.RoleArn = roleArn
	return src
}

func (src *AwsIamSource) WithAwsClient(client *cziAws.Client) *AwsIamSource {
	src.Client = client
	return src
}

func (src *AwsIamSource) WithMaxAge(maxAge time.Duration) *AwsIamSource {
	src.MaxAge = maxAge
	return src
}

// RotateKeys rotates the AWS IAM keys for the user specified in src.
// It returns any new key created and any error encountered.
// If the user has less than two keys, RotateKeys creates a new key.
// If the user has two keys, RotateKeys checks if the older key is older
// than the MaxAge specified in src. If so, RotateKeys deletes that key
// and returns a new key, else it does nothing and returns a nil key.
func (src *AwsIamSource) RotateKeys(ctx context.Context) (*iam.AccessKey, error) {
	svc := src.Client.IAM.Svc

	// list a user's access keys
	keys, err := svc.ListAccessKeysWithContext(ctx, &iam.ListAccessKeysInput{
		UserName: aws.String(src.UserName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list access keys")
	}

	if len(keys.AccessKeyMetadata) == 2 {
		// identify older access key
		var olderKey *iam.AccessKeyMetadata
		if keys.AccessKeyMetadata[1].CreateDate.After(*keys.AccessKeyMetadata[0].CreateDate) {
			olderKey = keys.AccessKeyMetadata[0]
		} else {
			olderKey = keys.AccessKeyMetadata[1]
		}

		// delete older access key if expired, else nothing to do
		if time.Since(*olderKey.CreateDate) > src.MaxAge {
			_, err = svc.DeleteAccessKeyWithContext(ctx, &iam.DeleteAccessKeyInput{
				AccessKeyId: aws.String(*olderKey.AccessKeyId),
				UserName:    aws.String(src.UserName),
			})
			if err != nil {
				return nil, errors.Wrap(err, "unable to delete older access key")
			}
		} else {
			return nil, nil
		}
	}

	// create a new IAM access key
	result, err := svc.CreateAccessKeyWithContext(ctx, &iam.CreateAccessKeyInput{
		UserName: aws.String(src.UserName),
	})
	return result.AccessKey, errors.Wrap(err, "unable to create new access key")
}

func (src *AwsIamSource) Read() (map[string]string, error) {
	newKey, err := src.RotateKeys(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to rotate keys")
	}
	if newKey == nil {
		return nil, nil
	}
	creds := map[string]string{
		AwsAccessKeyID:     *newKey.AccessKeyId,
		AwsSecretAccessKey: *newKey.SecretAccessKey,
	}
	return creds, nil
}

func (src *AwsIamSource) Kind() Kind {
	return KindAws
}
