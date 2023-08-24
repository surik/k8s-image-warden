package engine

import (
	"github.com/docker/distribution/reference"
)

func ParseImageReference(imageRef string) (string, string) {
	ref, err := reference.Parse(imageRef)
	if err != nil {
		return "", ""
	}

	name := ref.String()
	named, ok := ref.(reference.Named)
	if ok {
		name = named.Name()
	}

	tag := "latest"
	tagged, ok := ref.(reference.NamedTagged)
	if ok {
		tag = tagged.Tag()
	} else {
		digested, ok := ref.(reference.Digested)
		if ok {
			tag = digested.Digest().String()
		}
	}

	return name, tag
}
