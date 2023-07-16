package transformer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"
)

type ProjectValidationError struct {
	T       string `json:"type"`
	Message string `json:"message"`
}

var projectValidationErrorComposeNotFound = ProjectValidationError{
	T:       "compose-not-found",
	Message: "No compose file was present at the project.",
}

type ProjectValidationErrorInvalidCompose struct {
	T       string `json:"type"`
	Message string `json:"message"`
}

func (v *ProjectValidationErrorInvalidCompose) GetErrorMessage() string {
	return v.Message
}

func (v *ProjectValidationErrorInvalidCompose) Serialize() json.RawMessage {
	v.T = "invalid-compose"
	r, _ := json.Marshal(map[string]string{"type": "invalid-compose"})
	return r
}

type ProjectValidationResult struct {
	ProjectValidationError *ProjectValidationError
	IsCompose              bool
	ConfigFilePath         string
}

func Validate(projectPath string) (*ProjectValidationResult, error) {
	ergopackPath, err := findErgopackPath(projectPath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to find ergopack path")
	}

	if ergopackPath != "" {
		vErr, err := validateErgopack(projectPath, ergopackPath)
		return &ProjectValidationResult{
			ProjectValidationError: vErr,
			IsCompose:              false,
			ConfigFilePath:         ergopackPath,
		}, err
	}

	composePath, err := findComposePath(projectPath)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to find compose at %s", projectPath)
	}

	if composePath == "" {
		return &ProjectValidationResult{
			ProjectValidationError: &projectValidationErrorComposeNotFound,
		}, nil
	}

	vErr, err := validateCompose(projectPath, composePath)
	return &ProjectValidationResult{
		ProjectValidationError: vErr,
		IsCompose:              true,
		ConfigFilePath:         composePath,
	}, err
}

var ergopackPaths = []string{
	".ergomake/ergopack.yml",
	".ergomake/ergopack.yaml",
}

func findErgopackPath(projectPath string) (string, error) {
	var retErr error
	for _, candidate := range ergopackPaths {
		ergopackPath := path.Join(projectPath, candidate)
		if _, err := os.Stat(ergopackPath); err != nil {
			if !os.IsNotExist(err) {
				retErr = errors.Wrapf(err, "fail to stat %s", ergopackPath)
			}
			continue

		}

		return ergopackPath, nil
	}

	return "", retErr
}

func validateErgopack(projectPath string, ergopackPath string) (*ProjectValidationError, error) {
	// TODO
	return nil, nil
}

var composePaths = []string{
	".ergomake/compose.yml",
	".ergomake/compose.yaml",
	".ergomake/docker-compose.yml",
	".ergomake/docker-compose.yaml",
	"compose.yml",
	"compose.yaml",
	"docker-compose.yml",
	"docker-compose.yaml",
}

func findComposePath(projectPath string) (string, error) {
	var retErr error
	for _, candidate := range composePaths {
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

func validateCompose(projectPath, composePath string) (*ProjectValidationError, error) {
	relativePath, err := filepath.Rel(projectPath, composePath)
	if err != nil {
		relativePath = composePath
	}

	content, err := ioutil.ReadFile(composePath)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to read compose at %s", composePath)
	}

	var v interface{}
	err = yaml.Unmarshal(content, &v)
	if err != nil {
		return &ProjectValidationError{
			T:       "invalid-compose",
			Message: fmt.Sprintf("Compose file has syntax error\n```\n%s: %s\n```", relativePath, err.Error()),
		}, nil
	}

	asMap, ok := v.(map[string]interface{})
	if !ok {
		return &ProjectValidationError{T: "invalid-compose", Message: "Compose has no `services` defined."}, nil
	}

	rawSvcs, ok := asMap["services"]
	if !ok {
		return &ProjectValidationError{T: "invalid-compose", Message: "Compose has no `services` defined."}, nil
	}

	servicesMap, ok := rawSvcs.(map[string]interface{})
	if !ok {
		return &ProjectValidationError{T: "invalid-compose", Message: "Compose `services` has invalid type, expect a map."}, nil
	}

	services := make(map[string]map[string]interface{})
	for k, v := range servicesMap {
		service := v.(map[string]interface{})
		services[k] = service
	}

	validationErr, err := validateEnvFiles(composePath, services)

	return validationErr, errors.Wrap(err, "fail to validate env files")
}

func validateEnvFiles(composePath string, services map[string]map[string]interface{}) (*ProjectValidationError, error) {
	for name, svc := range services {
		rawEnvFile, ok := svc["env_file"]
		if !ok {
			continue
		}

		var envFiles []string
		asStr, ok := rawEnvFile.(string)
		if ok {
			envFiles = append(envFiles, asStr)
		} else {
			asArr, ok := rawEnvFile.([]interface{})
			if !ok {
				return &ProjectValidationError{
					T:       "invalid-compose",
					Message: fmt.Sprintf("`env_file` field of service `%s` has invalid type.", name),
				}, nil
			}

			for _, envFile := range asArr {
				asEnvFile, ok := envFile.(string)
				if !ok {
					return &ProjectValidationError{
						T:       "invalid-compose",
						Message: fmt.Sprintf("`env_file` field of service `%s` has invalid type", name),
					}, nil
				}

				envFiles = append(envFiles, asEnvFile)
			}
		}

		for _, envFile := range envFiles {
			envFilePath := path.Join(path.Dir(composePath), envFile)

			if _, err := os.Stat(envFilePath); err != nil {
				if os.IsNotExist(err) {
					return &ProjectValidationError{
						T:       "invalid-compose",
						Message: fmt.Sprintf("Env file `%s` for service `%s` not found.", envFile, name),
					}, nil
				}

				return nil, errors.Wrapf(err, fmt.Sprintf("fail to stat env file %s/%s", composePath, envFile))
			}
		}
	}

	return nil, nil
}
