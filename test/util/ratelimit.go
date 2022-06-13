package util

import (
	"fmt"
	"os"
	"path"

	"sigs.k8s.io/yaml"

	policyv1alpha1 "github.com/zirain/limiter/api/v1alpha1"
)

func ApplyRatelimit(name string) error {
	_, err := readFromYaml(name)
	if err != nil {
		return err
	}

	return nil
}

func readFromYaml(name string) (*policyv1alpha1.RateLimit, error) {
	b, err := os.ReadFile(path.Join("../../config/samples/", fmt.Sprintf("%s.yaml", name)))

	if err != nil {
		return nil, err
	}
	var rl policyv1alpha1.RateLimit
	err = yaml.Unmarshal(b, &rl)
	return &rl, err
}
