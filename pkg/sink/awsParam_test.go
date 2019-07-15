package sink_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/chanzuckerberg/rotator/pkg/source"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	region      = "us-west-2"
	roleArn     = "arn:aws:iam::exampleAccount:role/poweruser"
	parName     = "name"
	parValue    = "value"
	fakeParName = "non-existing parameter"
)

type TestSuite struct {
	suite.Suite

	ctx context.Context

	// aws
	awsClient *cziAws.Client
	mockSSM   *cziAws.MockSSMSvc
	sink      *sink.AwsParamSink

	// cleanup
	server *httptest.Server
}

func (ts *TestSuite) TearDownTest() {
	ts.server.Close()
}

func (ts *TestSuite) SetupTest() {
	ts.ctx = context.Background()

	sess, server := cziAws.NewMockSession()
	ts.server = server

	ts.awsClient = cziAws.New(sess)
	ts.awsClient, ts.mockSSM = ts.awsClient.WithMockSSM()
	ts.sink = &sink.AwsParamSink{
		Client: ts.awsClient,
	}

	// mock PutParameterWithContext
	putOut := &ssm.PutParameterOutput{}
	ts.mockSSM.On("PutParameterWithContext", mock.Anything).Return(putOut, nil)
}

func (ts *TestSuite) TestWriteToAwsParamSinkFakeParam() {
	t := ts.T()
	r := require.New(t)

	// mock GetParameterWithContext for non-existing parameter
	in := &ssm.GetParameterInput{}
	in.SetName(fakeParName)
	ts.mockSSM.On("GetParameterWithContext", in).Return(&ssm.GetParameterOutput{}, awserr.New(ssm.ErrCodeParameterNotFound, "", nil))

	// test presence of log message
	logger, hook := test.NewNullLogger()
	errMsg := fmt.Sprintf("%s: parameter not found, skipping rotation", fakeParName)
	logger.Error(errMsg)

	err := ts.sink.Write(map[string]string{
		fakeParName: parValue,
	})
	r.Nil(err)
	r.Equal(1, len(hook.Entries))
	r.Equal(logrus.ErrorLevel, hook.LastEntry().Level)
	r.Equal(errMsg, hook.LastEntry().Message)
}

func (ts *TestSuite) TestWriteToAwsParamSink() {
	t := ts.T()
	r := require.New(t)

	// mock GetParameterWithContext for existing parameter
	in := &ssm.GetParameterInput{}
	in.SetName(parName)
	par := &ssm.Parameter{}
	par.SetName(parName).SetValue(parValue)
	getOut := &ssm.GetParameterOutput{}
	getOut.SetParameter(par)
	ts.mockSSM.On("GetParameterWithContext", in).Return(getOut, nil)

	err := ts.sink.Write(map[string]string{
		parName: parValue,
	})
	r.Nil(err)
}

func TestProviderSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func TestWriteToAwsParamSink_Integration(t *testing.T) {
	r := require.New(t)

	// Create a SSM client from a session.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region), // SSM functions require region configuration
	})
	r.Nil(err)
	sess.Config.Credentials = stscreds.NewCredentials(sess, roleArn) // the new Credentials object wraps the AssumeRoleProvider
	client := cziAws.New(sess).WithSSM(sess.Config)

	sink := sink.AwsParamSink{Client: client}
	svc := client.SSM.Svc

	// get the secret
	name := source.Secret
	in := &ssm.GetParameterInput{
		Name: &name,
	}
	in = in.SetWithDecryption(true)
	old, err := svc.GetParameterWithContext(context.Background(), in)
	r.Nil(err)

	// rotate the secret
	creds, err := (&source.DummySource{}).Read()
	r.Nil(err)
	err = sink.Write(creds)
	r.Nil(err)
	new, err := svc.GetParameterWithContext(context.Background(), in)
	r.Nil(err)

	// check new parameter value and other attributes
	r.Equal(creds[name], *new.Parameter.Value)
	r.NotEqual(*old.Parameter.Version, *new.Parameter.Version)
	r.Equal(*old.Parameter.Type, *new.Parameter.Type)
	r.Equal(*old.Parameter.ARN, *new.Parameter.ARN)
}