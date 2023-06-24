package dockerutils

import (
	"github.com/google/go-containerregistry/pkg/name"
)

func ExtractDockerRegistryURL(imageURL string) (string, error) {
	ref, err := name.ParseReference(imageURL)
	if err != nil {
		return "", err
	}

	return ref.Context().RegistryStr(), nil
}
