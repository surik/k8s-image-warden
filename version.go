package k8simagewarden

import (
	"strings"
)

var (
	Tag     = "unknown"
	Version = func() string {
		if Tag == "" {
			return "unknown"
		}
		return strings.TrimLeft(Tag, "v")
	}()
)
