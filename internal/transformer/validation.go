package transformer

import (
	"encoding/json"
	"os"
	"path"

	"github.com/pkg/errors"
)

type ProjectValidationError interface {
	GetErrorMessage() string
	Serialize() json.RawMessage
}

type ProjectValidationResultComposeNotFound struct{}

func (v *ProjectValidationResultComposeNotFound) GetErrorMessage() string {
	return "No compose file was present at the project."
}

func (*ProjectValidationResultComposeNotFound) Serialize() json.RawMessage {
	r, _ := json.Marshal(map[string]string{"type": "compose-not-found"})
	return r
}

func (c *gitCompose) validateProject() (ProjectValidationError, error) {
	composePath, err := findComposePath(c.projectPath)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to find compose at %s", c.projectPath)
	}

	if composePath == "" {
		return &ProjectValidationResultComposeNotFound{}, nil
	}

	c.composePath = composePath

	return nil, nil
}

func findComposePath(projectPath string) (string, error) {
	candidates := []string{
		".ergomake/compose.yml",
		".ergomake/compose.yaml",
		".ergomake/docker-compose.yml",
		".ergomake/docker-compose.yaml",
		"compose.yml",
		"compose.yaml",
		"docker-compose.yml",
		"docker-compose.yaml",
	}

	var retErr error
	for _, candidate := range candidates {
		composePath := path.Join(projectPath, candidate)
		if _, err := os.Stat(composePath); err != nil {
			if !os.IsNotExist(err) {
				retErr = errors.Wrapf(err, "fail to stat %s", composePath)
			}
			continue

		}

		return composePath, nil
	}

	return "", retErr
}
