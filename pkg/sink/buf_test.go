package sink_test

import (
	"context"
	"testing"

	"github.com/chanzuckerberg/rotator/pkg/sink"
	"github.com/stretchr/testify/require"
)

func TestWriteToBufSink(t *testing.T) {
	r := require.New(t)

	sink := sink.NewBufSink()
	secret := "EXAMPLESECRET"
	err := sink.Write(context.Background(), "secret", secret)
	r.Nil(err)
	written := sink.Read()
	r.Equal(secret, written)
}
