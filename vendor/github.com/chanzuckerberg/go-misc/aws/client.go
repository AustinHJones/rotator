package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Client is an aws client
type Client struct {
	session *session.Session

	// services
	EC2            *EC2
	IAM            *IAM
	KMS            *KMS
	Lambda         *Lambda
	S3             *S3
	Firehose       *Firehose
	SecretsManager *SecretsManager
	SSM            *SSM
	STS            *STS
	Support        *Support
}

// New returns a new aws client
func New(sess *session.Session) *Client {
	return &Client{session: sess}
}

// WithAllServices Convenience method that configures all services with the same aws.Config
func (c *Client) WithAllServices(conf *aws.Config) *Client {
	return c.
		WithEC2(conf).
		WithIAM(conf).
		WithKMS(conf).
		WithLambda(conf).
		WithS3(conf).
		WithFirehose(conf).
		WithSecretsManager(conf).
		WithSSM(conf).
		WithSTS(conf).
		WithSupport(conf)
}

// ------- SecretsManager -----------

// WithSecretsManager configures a Secrets Manager svc
func (c *Client) WithSecretsManager(conf *aws.Config) *Client {
	c.SecretsManager = NewSecretsManager(c.session, conf)
	return c
}

// WithMockSecretsManager mocks the Secrets Manager svc
func (c *Client) WithMockSecretsManager() (*Client, *MockSecretsManagerSvc) {
	mock := NewMockSecretsManager()
	c.SecretsManager = &SecretsManager{Svc: mock}
	return c, mock
}

// ------- Firehose -----------

// WithFirehose configures the firehose service
func (c *Client) WithFirehose(conf *aws.Config) *Client {
	c.Firehose = NewFirehose(c.session, conf)
	return c
}

// TODO(el): Firehose mock

// ------- S3 -----------

// WithS3 configures the s3 client
func (c *Client) WithS3(conf *aws.Config) *Client {
	c.S3 = NewS3(c.session, conf)
	return c
}

// WithMockS3 mocks s3 svc
func (c *Client) WithMockS3() (*Client, *MockS3Svc, *MockS3Manager) {
	mock := NewMockS3()
	mockDownloader := NewMockS3Manager()
	c.S3 = &S3{Svc: mock, Downloader: mockDownloader}
	return c, mock, mockDownloader
}

// ------- IAM -----------

// WithIAM configures the IAM SVC
func (c *Client) WithIAM(conf *aws.Config) *Client {
	c.IAM = NewIAM(c.session, conf)
	return c
}

// WithMockIAM mocks iam svc
func (c *Client) WithMockIAM() (*Client, *MockIAMSvc) {
	mock := NewMockIAM()
	c.IAM = &IAM{Svc: mock}
	return c, mock
}

// ------- SSM -----------

// WithSSM configures the SSM service
func (c *Client) WithSSM(conf *aws.Config) *Client {
	c.SSM = NewSSM(c.session, conf)
	return c
}

// WithMockSSM mocks the SSM service
func (c *Client) WithMockSSM() (*Client, *MockSSMSvc) {
	mock := NewMockSSM()
	c.SSM = &SSM{Svc: mock}
	return c, mock
}

// ------- STS -----------

// WithSTS configures the STS service
func (c *Client) WithSTS(conf *aws.Config) *Client {
	c.STS = NewSTS(c.session, conf)
	return c
}

// WithMockSTS mocks the STS service
func (c *Client) WithMockSTS() (*Client, *MockSTSSvc) {
	mock := NewMockSTS()
	c.STS = &STS{Svc: mock}
	return c, mock
}

// ------- Lambda -----------

// WithLambda configures the lambda service
func (c *Client) WithLambda(conf *aws.Config) *Client {
	c.Lambda = NewLambda(c.session, conf)
	return c
}

// WithMockLambda mocks the lambda service
func (c *Client) WithMockLambda() (*Client, *MockLambdaSvc) {
	mock := NewMockLambda()
	c.Lambda = &Lambda{Svc: mock}
	return c, mock
}

// ------- KMS -----------

// WithKMS configures the kms service
func (c *Client) WithKMS(conf *aws.Config) *Client {
	c.KMS = NewKMS(c.session, conf)
	return c
}

// WithMockKMS mocks the kms service
func (c *Client) WithMockKMS() (*Client, *MockKMSSvc) {
	mock := NewMockKMS()
	c.KMS = &KMS{Svc: mock}
	return c, mock
}

// ------- EC2 -----------

// WithEC2 configures an EC2 svc
func (c *Client) WithEC2(conf *aws.Config) *Client {
	c.EC2 = NewEC2(c.session, conf)
	return c
}

// ------- Support -----------

// WithSupport configures an Support svc
func (c *Client) WithSupport(conf *aws.Config) *Client {
	c.Support = NewSupport(c.session, conf)
	return c
}
