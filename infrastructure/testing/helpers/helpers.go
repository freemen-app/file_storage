package helpers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	MockCalls []struct {
		Method     string
		Args       []interface{}
		ReturnArgs []interface{}
	}
)

var DefaultCtx = context.Background()

func TimeoutCtx(t *testing.T, ctx context.Context, timeout time.Duration) context.Context {
	t.Helper()
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	t.Cleanup(cancel)
	return timeoutCtx
}

func OpenFile(t *testing.T, filename string) *os.File {
	t.Helper()
	file, err := os.Open(filename)
	assert.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, file.Close())
	})
	return file
}
