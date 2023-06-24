package dockerutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDockerRegistryURL(t *testing.T) {
	t.Parallel()

	tt := []struct {
		imageURL      string
		expectedURL   string
		expectedError error
	}{
		{
			imageURL:      "gcr.io/project/image:tag",
			expectedURL:   "gcr.io",
			expectedError: nil,
		},
		{
			imageURL:      "postgres:13-alpine",
			expectedURL:   "index.docker.io",
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.imageURL, func(t *testing.T) {
			t.Parallel()

			url, err := ExtractDockerRegistryURL(tc.imageURL)

			assert.Equal(t, tc.expectedURL, url)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
