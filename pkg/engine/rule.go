package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Masterminds/semver"
	"github.com/surik/k8s-image-warden/pkg/repo"
)

type ValidateType string

const (
	ValidateTypeLatest     ValidateType = "Latest"
	ValidateTypeSemVer     ValidateType = "SemVer"
	ValidateTypeLock       ValidateType = "Lock"
	ValidateTypeRollingTag ValidateType = "RollingTag"
)

type MutationType string

const (
	MutationTypeDefaultRegistry MutationType = "DefaultRegistry"
	MutationTypeRewriteRegistry MutationType = "RewriteRegistry"
)

var (
	ErrWrongRuleType = errors.New("wrong rule type")
)

type MutationRule struct {
	Type           MutationType   `yaml:"type"`
	Registry       string         `yaml:"registry,omitempty"`
	RegistryRegexp *regexp.Regexp `yaml:"-"`
	NewRegistry    string         `yaml:"newRegisty,omitempty"`
}

type ValidationRule struct {
	Type            ValidateType        `yaml:"type"`
	ImageName       string              `yaml:"imageName,omitempty"`
	ImageNameRegexp *regexp.Regexp      `yaml:"-"`
	ImageTag        string              `yaml:"imageTag,omitempty"`
	ImageTagSemVer  *semver.Constraints `yaml:"-"`
	Allow           bool                `yaml:"allow"`
	RollingTagAfter time.Time           `yaml:"after,omitempty"`
}

type Rule struct {
	Name           string         `yaml:"name"`
	MutationRule   MutationRule   `yaml:"mutate,omitempty"`
	ValidationRule ValidationRule `yaml:"validate,omitempty"`
}

type Rules struct {
	Rules []Rule `yaml:"rules"`
}

func (r Rule) compile() (Rule, error) {
	if r.ValidationRule.Type != "" && r.MutationRule.Type != "" {
		return r, fmt.Errorf("%w: should be either Validation or Mutation", ErrWrongRuleType)
	}

	if r.ValidationRule.Type != "" {
		return r.compileValidateRule()
	}

	return r.compileMutateRule()
}

func (r Rule) compileValidateRule() (Rule, error) {
	if r.ValidationRule.Type == ValidateTypeSemVer {
		compiled, err := semver.NewConstraint(r.ValidationRule.ImageTag)
		if err != nil {
			return r, err
		}
		r.ValidationRule.ImageTagSemVer = compiled
	}

	compiled, err := regexp.Compile(r.ValidationRule.ImageName)
	if err != nil {
		return r, err
	}
	r.ValidationRule.ImageNameRegexp = compiled

	return r, nil
}

func (r Rule) compileMutateRule() (Rule, error) {
	if r.MutationRule.Type == MutationTypeRewriteRegistry {
		compiled, err := regexp.Compile(r.MutationRule.Registry)
		if err != nil {
			return r, err
		}
		r.MutationRule.RegistryRegexp = compiled
	}

	return r, nil
}

func (r MutationRule) Mutate(domain string) (string, bool) {
	switch r.Type {
	case MutationTypeDefaultRegistry:
		if domain == "" {
			return r.Registry, true
		}
	case MutationTypeRewriteRegistry:
		if r.RegistryRegexp.MatchString(domain) {
			return r.NewRegistry, true
		}
	default:
	}
	return domain, false
}

func (r ValidationRule) Match(ctx context.Context, repo *repo.Repo, inspector ImageInspector, name, tag string) bool {
	switch r.Type {
	case ValidateTypeLatest:
		if r.matchName(name) && tag == "latest" {
			return true
		}
	case ValidateTypeLock:
		if r.matchName(name) && tag == r.ImageTag {
			return true
		}
	case ValidateTypeRollingTag:
		if r.matchName(name) {
			return r.validateRollingTag(ctx, repo, inspector, name, tag)
		}
	case ValidateTypeSemVer:
		if !r.matchName(name) {
			return false
		}
		version, err := semver.NewVersion(tag)
		if err != nil {
			return false
		}
		constraint := r.ImageTagSemVer
		if constraint.Check(version) {
			return true
		}
	default:
		return false
	}

	return false
}

func (r ValidationRule) matchName(name string) bool {
	if r.ImageName == "" { // name is not set means matches everything
		return true
	}

	if r.ImageNameRegexp != nil {
		return r.ImageNameRegexp.MatchString(name)
	}

	// regexp was not not properly compiled. this is abnormal case
	return false
}

func (r ValidationRule) validateRollingTag(parentCtx context.Context, repo *repo.Repo, inspector ImageInspector, name, tag string) bool {
	ids, err := repo.GetIDsByNameAndAfter(name+":"+tag, r.RollingTagAfter)
	if err != nil {
		log.Println(err)
		return false
	}

	if len(ids) > 1 {
		return true
	}

	if len(ids) == 1 {
		digests, err := repo.GetDigestsByNameAndAfter(name+":"+tag, r.RollingTagAfter)
		if err != nil {
			log.Println(err)
			return false
		}

		ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
		defer cancel()

		digest, err := inspector.GetDigest(ctx, "docker://"+name)
		if err != nil {
			log.Printf("error when inspecting image %s: %s", name, err)
			return false
		}

		// could there be any others?
		digestAlgorithm := "sha256"

		// in the registry we have another image that refers to the same tag. this is a rolling tag
		if digests[0] != digestAlgorithm+":"+digest {
			return true
		}
	}

	return false
}
