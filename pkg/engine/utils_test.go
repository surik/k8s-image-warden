package engine_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surik/k8s-image-warden/pkg/engine"
)

func TestUtils(t *testing.T) {
	alpineDigest := "sha256:7144f7bab3d4c2648d7e59409f15ec52a18006a128c733fcff20d3a4a54ba44a"
	alpine := "alpine@" + alpineDigest

	t.Run("Digest", func(t *testing.T) {
		name, tag := engine.ParseImageReference(alpine)
		require.Equal(t, "alpine", name)
		require.Equal(t, alpineDigest, tag)
	})

	t.Run("Latest", func(t *testing.T) {
		name, tag := engine.ParseImageReference("alpine")
		require.Equal(t, "alpine", name)
		require.Equal(t, "latest", tag)
	})

	t.Run("Tag", func(t *testing.T) {
		name, tag := engine.ParseImageReference("alpine:1.2.3")
		require.Equal(t, "alpine", name)
		require.Equal(t, "1.2.3", tag)
	})
}
