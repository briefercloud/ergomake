package transformer

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitCompose_validateProjectComposeNotFound(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "validate_project_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	vErr, err := Validate(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, &projectValidationErrorComposeNotFound, vErr)
}

func TestGitCompose_validateProjectFindsComposeEverywhere(t *testing.T) {
	for _, composePath := range composePaths {
		t.Run(composePath, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "validate_project_test")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			err = os.Mkdir(path.Join(tmpDir, ".ergomake"), 0700)
			require.NoError(t, err)

			err = ioutil.WriteFile(
				path.Join(tmpDir, composePath),
				[]byte(`
version: '3'

services:
  web:
    image: nginx
`),
				0644,
			)
			require.NoError(t, err)

			vErr, err := Validate(tmpDir)
			require.NoError(t, err)

			assert.Nil(t, vErr)
		})
	}
}

func TestGitCompose_validateProjectInvalidYAMLSyntax(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "validate_project_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = ioutil.WriteFile(
		path.Join(tmpDir, "compose.yaml"),
		[]byte(`
version: '3'

services:
  web:
    image: nginx

 missindented:
   image: mongo
`),
		0644,
	)
	require.NoError(t, err)

	vErr, err := Validate(tmpDir)
	require.NoError(t, err)

	require.NotNil(t, vErr.ProjectValidationError)
	assert.Equal(t, "invalid-compose", vErr.ProjectValidationError.T)
}

func TestGitCompose_validateProjectInvalidCompose(t *testing.T) {
	tt := []struct {
		name    string
		compose string
	}{
		{
			name:    "compose is not a map",
			compose: "'this is just a string'",
		},
		{
			name: "invalid services",
			compose: `
version: '3'
services:
 - 'a list?'
 - 'that is invalid'
`,
		},
		{
			name: "invalid env_file",
			compose: `
version: '3'
services:
  web:
    env_file: 
      a_map: 'is_not_supported'
`,
		},
		{
			name: "non existing env_file",
			compose: `
version: '3'
services:
  web:
    env_file: 
      - 'this_file_does_not_exists.env'
`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "validate_project_test")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			err = ioutil.WriteFile(
				path.Join(tmpDir, "compose.yaml"),
				[]byte(tc.compose),
				0644,
			)
			require.NoError(t, err)

			vErr, err := Validate(tmpDir)
			require.NoError(t, err)
			require.NotNil(t, vErr.ProjectValidationError)

			assert.Equal(t, "invalid-compose", vErr.ProjectValidationError.T)
		})
	}
}

func TestGitCompose_validateProjectAcceptsGoodYAMLReferencingExistingEnvFiles(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "validate_project_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dir := path.Join(tmpDir, "inner")
	err = os.Mkdir(dir, 0700)
	require.NoError(t, err)

	err = os.Mkdir(path.Join(dir, ".ergomake"), 0700)
	require.NoError(t, err)

	err = ioutil.WriteFile(
		path.Join(dir, ".ergomake/compose.yaml"),
		[]byte(`
version: '3'
services:
  web:
    env_file: 
      - '../.env'
`),
		0644,
	)
	require.NoError(t, err)

	err = ioutil.WriteFile(
		path.Join(dir, ".env"),
		[]byte("MY_VAR=MY_VALUE"),
		0644,
	)
	require.NoError(t, err)

	vErr, err := Validate(tmpDir)
	require.NoError(t, err)

	assert.Nil(t, vErr.ProjectValidationError)
}
