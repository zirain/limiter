package util

import (
	"context"
	"fmt"
	"os"
	"path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
	clientset "github.com/zirain/limiter/client-go/generated/clientset/versioned"
)

func ApplyRatelimit(client clientset.Interface, name string, namespace string) error {
	rl, err := readFromYaml(name)
	if err != nil {
		return err
	}
	rl.Namespace = namespace

	_, err = client.PolicyV1alpha1().RateLimits(rl.Namespace).Create(context.TODO(), rl, metav1.CreateOptions{})

	return err
}

func DeleteRatelimit(client clientset.Interface, name string, namespace string) error {
	rl, err := readFromYaml(name)
	if err != nil {
		return err
	}
	rl.Namespace = namespace

	return client.PolicyV1alpha1().RateLimits(rl.Namespace).Delete(context.TODO(), rl.Name, metav1.DeleteOptions{})
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
