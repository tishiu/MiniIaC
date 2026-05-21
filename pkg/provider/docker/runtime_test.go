package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeImageClient struct {
	imageInspectCalls int
	imagePullCalls    int
	inspectErr        error
	pullErr           error
	pullBody          io.ReadCloser
}

func (f *fakeImageClient) ImageInspect(context.Context, string, ...client.ImageInspectOption) (imagetypes.InspectResponse, error) {
	f.imageInspectCalls++
	return imagetypes.InspectResponse{}, f.inspectErr
}

func (f *fakeImageClient) ImagePull(context.Context, string, imagetypes.PullOptions) (io.ReadCloser, error) {
	f.imagePullCalls++
	if f.pullErr != nil {
		return nil, f.pullErr
	}
	if f.pullBody != nil {
		return f.pullBody, nil
	}
	return io.NopCloser(bytes.NewBufferString(`{}`)), nil
}

func TestEnsureImagePullsMissingImage(t *testing.T) {
	fake := &fakeImageClient{
		inspectErr: errdefs.NotFound(fmt.Errorf("missing image")),
	}

	err := ensureImage(context.Background(), fake, "nginx:alpine")
	require.NoError(t, err)
	assert.Equal(t, 1, fake.imageInspectCalls)
	assert.Equal(t, 1, fake.imagePullCalls)
}

func TestEnsureImageSkipsPullWhenImageExists(t *testing.T) {
	fake := &fakeImageClient{}

	err := ensureImage(context.Background(), fake, "nginx:alpine")
	require.NoError(t, err)
	assert.Equal(t, 1, fake.imageInspectCalls)
	assert.Equal(t, 0, fake.imagePullCalls)
}
