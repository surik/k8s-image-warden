package engine

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/docker/distribution/reference"
	"github.com/surik/k8s-image-warden/pkg/repo"
	"gopkg.in/yaml.v3"
)

type Engine struct {
	rules     []Rule
	repo      *repo.Repo
	inspector ImageInspector
}

var (
	ErrBadImageReference = errors.New("bad image reference")
)

func NewEngine(repo *repo.Repo, inspector ImageInspector, rules []Rule) (*Engine, error) {
	compiledRules := make([]Rule, len(rules))
	for i, rule := range rules {
		compiled, err := rule.compile()
		if err != nil {
			return nil, err
		}
		compiledRules[i] = compiled
	}

	return &Engine{
		repo:      repo,
		rules:     compiledRules,
		inspector: inspector,
	}, nil
}

func NewEngineFromFile(repo *repo.Repo, inspector ImageInspector, file string) (*Engine, error) {
	var rules Rules

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &rules)
	if err != nil {
		return nil, err
	}

	return NewEngine(repo, inspector, rules.Rules)
}

func (e Engine) GetRules() []Rule {
	return e.rules
}

func (e Engine) Validate(ctx context.Context, imageRef string) (bool, string) {
	name, tag := ParseImageReference(imageRef)

	for _, rule := range e.rules {
		if rule.ValidationRule.Match(ctx, e.repo, e.inspector, name, tag) {
			return rule.ValidationRule.Allow, rule.Name
		}
	}

	return false, "<No Rules>"
}

func (e Engine) Mutate(_ context.Context, imageRef string) (string, []string) {
	ref, err := reference.Parse(imageRef)
	if err != nil {
		return imageRef, []string{err.Error()}
	}

	named, ok := ref.(reference.NamedTagged)
	if !ok {
		return imageRef, []string{fmt.Errorf("%w: could not cast to reference.NamedTagged", ErrBadImageReference).Error()}
	}

	domain := reference.Domain(named)
	var rules []string

	for _, rule := range e.rules {
		newDomain, mutated := rule.MutationRule.Mutate(domain)
		if mutated {
			domain = newDomain
			rules = append(rules, rule.Name)
		}
	}

	if len(rules) > 0 {
		return domain + "/" + reference.Path(named) + ":" + named.Tag(), rules
	}

	return imageRef, rules
}
