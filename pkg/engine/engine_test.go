package engine_test

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surik/k8s-image-warden/pkg/engine"
	helpers "github.com/surik/k8s-image-warden/pkg/repo/testing"
)

type fakeInspector struct{}

func newFakeInspector() *fakeInspector {
	return &fakeInspector{}
}

func (i *fakeInspector) GetDigest(_ context.Context, name string) (string, error) {
	if name == "docker://k8s-image-warden-controller:latest" {
		return helpers.Digest2, nil
	}
	return "", nil
}

func TestEngine_Validate(t *testing.T) {
	rules := []engine.Rule{
		{
			Name: "No Latest",
			ValidationRule: engine.ValidationRule{
				Type:  engine.ValidateTypeLatest,
				Allow: false,
			},
		},
		{
			Name: "Dev tag is allowed for our organisation",
			ValidationRule: engine.ValidationRule{
				Type:      engine.ValidateTypeLock,
				ImageName: `docker\.io/mycompany/.*`,
				ImageTag:  "dev",
				Allow:     true,
			},
		},
		{
			Name: "Alpine is any 3.x.y",
			ValidationRule: engine.ValidationRule{
				Type:      engine.ValidateTypeSemVer,
				ImageName: `docker\.io/alpine`,
				ImageTag:  ">= 3.0.0, < 4.0.0",
				Allow:     true,
			},
		},
		{
			Name: "nginx is newer than 1.0.0",
			ValidationRule: engine.ValidationRule{
				Type:      engine.ValidateTypeSemVer,
				ImageName: `docker\.io/nginx`,
				ImageTag:  ">= 1.0.0",
				Allow:     true,
			},
		},
	}

	ruleEngine, err := engine.NewEngine(nil, nil, rules)
	require.NoError(t, err)

	validate(t, ruleEngine, "docker.io/nginx", false, rules[0].Name)
	validate(t, ruleEngine, "docker.io/nginx:latest", false, rules[0].Name)
	validate(t, ruleEngine, "docker.io/nginx:0.9.1", false, "<No Rules>")
	validate(t, ruleEngine, "docker.io/nginx:1.9.1", true, rules[3].Name)
	validate(t, ruleEngine, "docker.io/nginx:5.0.0", true, rules[3].Name)
	validate(t, ruleEngine, "docker.io/alpine", false, rules[0].Name)
	validate(t, ruleEngine, "docker.io/alpine:2.9.1", false, "<No Rules>")
	validate(t, ruleEngine, "docker.io/alpine:3.9.1", true, rules[2].Name)
	validate(t, ruleEngine, "docker.io/alpine:4.9.1", false, "<No Rules>")
	validate(t, ruleEngine, "docker.io/mycompany/app1:dev", true, rules[1].Name)
	validate(t, ruleEngine, "docker.io/mycompany/app2:dev", true, rules[1].Name)
	validate(t, ruleEngine, "docker.io/alpine:dev", false, "<No Rules>")
}

func TestEngine_Mutate(t *testing.T) {
	rules := []engine.Rule{
		{
			Name: "docker.io is default",
			MutationRule: engine.MutationRule{
				Type:     engine.MutationTypeDefaultRegistry,
				Registry: "docker.io",
			},
		},
		{
			Name: "rewrite .com and .net to .io",
			MutationRule: engine.MutationRule{
				Type:        engine.MutationTypeRewriteRegistry,
				Registry:    "docker(.com|.net)",
				NewRegistry: "docker.io",
			},
		},
	}

	ruleEngine, err := engine.NewEngine(nil, nil, rules)
	require.NoError(t, err)
	mutate(t, ruleEngine, "nginx:latest", "docker.io/nginx:latest", []string{rules[0].Name})
	mutate(t, ruleEngine, "ghc.io/nginx:latest", "ghc.io/nginx:latest", nil)
	mutate(t, ruleEngine, "ghc.io/org/app:latest", "ghc.io/org/app:latest", nil)
	mutate(t, ruleEngine, "docker.com/nginx:latest", "docker.io/nginx:latest", []string{rules[1].Name})
	mutate(t, ruleEngine, "docker.net/nginx:latest", "docker.io/nginx:latest", []string{rules[1].Name})
}

func TestEngine_ParseYaml(t *testing.T) {
	e, err := engine.NewEngineFromFile(nil, nil, path.Join("..", "..", "testdata", "rules.yaml"))
	require.NoError(t, err)

	rules := e.GetRules()
	require.Len(t, rules, 4)

	require.Equal(t, engine.MutationTypeDefaultRegistry, rules[0].MutationRule.Type)
	require.Equal(t, "docker.io", rules[0].MutationRule.Registry)

	require.Equal(t, engine.ValidateTypeLatest, rules[1].ValidationRule.Type)
	require.Equal(t, false, rules[1].ValidationRule.Allow)

	require.Equal(t, engine.ValidateTypeSemVer, rules[2].ValidationRule.Type)
	require.Equal(t, true, rules[2].ValidationRule.Allow)
	require.Equal(t, "docker\\.io/nginx", rules[2].ValidationRule.ImageName)
	require.Equal(t, ">= 1.0.0", rules[2].ValidationRule.ImageTag)

	after, err := time.Parse("2006-01-02", "2023-07-01")
	require.NoError(t, err)
	require.Equal(t, engine.ValidateTypeRollingTag, rules[3].ValidationRule.Type)
	require.Equal(t, false, rules[3].ValidationRule.Allow)
	require.GreaterOrEqual(t, rules[3].ValidationRule.RollingTagAfter, after)
}

func TestEngine_ValidateRollingTags(t *testing.T) {
	repo := helpers.NewTestRepo(t)

	err := helpers.PrepareRollingTags(repo)
	require.NoError(t, err)

	defaultRules := []engine.Rule{
		{
			Name: "No Rolling tags",
			ValidationRule: engine.ValidationRule{
				Type:  engine.ValidateTypeRollingTag,
				Allow: false,
			},
		},
		{
			Name: "Allow Latest",
			ValidationRule: engine.ValidationRule{
				Type:  engine.ValidateTypeLatest,
				Allow: true,
			},
		},
	}

	t.Run("Disallow rolling tags", func(t *testing.T) {
		rules := defaultRules
		ruleEngine, err := engine.NewEngine(repo, newFakeInspector(), rules)
		require.NoError(t, err)
		validate(t, ruleEngine, "k8s-image-warden-agent:latest", false, rules[0].Name)
		validate(t, ruleEngine, "nginx:latest", true, rules[1].Name)
	})

	t.Run("Disalow rolling tag found by inspector", func(t *testing.T) {
		rules := defaultRules
		ruleEngine, err := engine.NewEngine(repo, newFakeInspector(), rules)
		require.NoError(t, err)
		validate(t, ruleEngine, "k8s-image-warden-controller:latest", false, rules[0].Name)
	})

	t.Run("Allow rolling tags", func(t *testing.T) {
		rules := defaultRules
		rules[0].Name = "Allow rolling Tags"
		rules[0].ValidationRule.Allow = true
		rules[1].Name = "No latests"
		rules[1].ValidationRule.Allow = false
		ruleEngine, err := engine.NewEngine(repo, newFakeInspector(), rules)
		require.NoError(t, err)
		validate(t, ruleEngine, "k8s-image-warden-agent:latest", true, rules[0].Name)
		validate(t, ruleEngine, "nginx:latest", false, rules[1].Name)
	})
}

func validate(t *testing.T, ruleEngine *engine.Engine, image string, expectedResult bool, expectedRule string) {
	t.Helper()

	result, rule := ruleEngine.Validate(context.Background(), image)

	if !assert.Equal(t, expectedResult, result) {
		t.FailNow()
	}

	if assert.Equal(t, expectedRule, rule) {
		return
	}

	t.FailNow()
}

func mutate(t *testing.T, ruleEngine *engine.Engine, image string, expectedReference string, expectedRules []string) {
	t.Helper()

	reference, rules := ruleEngine.Mutate(context.Background(), image)

	if !assert.Equal(t, expectedReference, reference) {
		t.FailNow()
	}

	if assert.Equal(t, expectedRules, rules) {
		return
	}

	t.FailNow()
}
