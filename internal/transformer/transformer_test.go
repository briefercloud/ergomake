package transformer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const rawCompose = `
# the string redis: appears first as a comment
version: "3"
# the string services: appears as comment also
volumes:
  redis:
services:
  # redis: also appears here
  mongo: # redis: also appears here
    image: mongo
    ports:
      - "27017:27017"
  redis: # make  a comment here to increase confusion
    image: redis
    ports:
      - "6379:6379"
    volumes:
      - redis:/data
volumes:
# make it show mongo: again to test better ServicesOrder
  mongo:
`

func TestTransformerCompose_FirstService(t *testing.T) {
	compose := Compose{
		Services: map[string]EnvironmentService{
			"redis": {Image: "redis", Index: 1},
			"mongo": {Image: "mongo", Index: 0},
		},
		RawCompose: rawCompose,
	}

	firstService := compose.FirstService()

	assert.Equal(t, compose.Services["mongo"], firstService)
}

func TestTransformerCompose_computeServicesIndexes(t *testing.T) {
	tt := []struct {
		name    string
		compose Compose
		want    map[string]EnvironmentService
	}{
		{
			name: "complicated compose",
			compose: Compose{
				Services: map[string]EnvironmentService{
					"redis": {Image: "redis"},
					"mongo": {Image: "mongo"},
				},
				RawCompose: rawCompose,
			},
			want: map[string]EnvironmentService{
				"redis": {Image: "redis", Index: 1},
				"mongo": {Image: "mongo", Index: 0},
			},
		},
		{
			name: "call service something that can be a key like 'ports' or 'image'",
			compose: Compose{
				Services: map[string]EnvironmentService{
					"ports": {Image: "mongo"},
					"image": {Image: "redis"},
				},
				RawCompose: `
version: '3.8'
services:
  image:
    image: redis
    ports:
      - '8000:8000'

  ports:
    image: mongo
    ports:
      - 27017`,
			},
			want: map[string]EnvironmentService{
				"ports": {Image: "mongo", Index: 1},
				"image": {Image: "redis", Index: 0},
			},
		},
		{
			name: "when first lines after services is are comments and empty lines",
			compose: Compose{
				Services: map[string]EnvironmentService{
					"ports": {Image: "mongo"},
					"image": {Image: "redis"},
				},
				RawCompose: `
version: '3.8'
services:

 # a comment. the line above is empty, the line below has one leading space
 
  image:
    image: redis
    ports:
      - '8000:8000'

  ports:
    image: mongo
    ports:
      - 27017`,
			},
			want: map[string]EnvironmentService{
				"ports": {Image: "mongo", Index: 1},
				"image": {Image: "redis", Index: 0},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.compose.computeServicesIndexes()
			assert.Equal(t, tc.want, tc.compose.Services)
		})
	}
}
