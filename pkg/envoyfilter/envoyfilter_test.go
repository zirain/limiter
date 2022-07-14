package envoyfilter

import (
	"fmt"
	"os"
	"path"
	"testing"

	clientnetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pkg/test/util/assert"
	"sigs.k8s.io/yaml"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
)

func TestToEnvoyFilterGlobal(t *testing.T) {
	cases := []string{
		"global/basic",
	}

	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			rl, err := readInput(tc)
			assert.NoError(t, err)

			expected, err := readExpected(tc)
			assert.NoError(t, err)

			got := ToEnvoyFilter(rl)
			assert.Equal(t, expected, got)
		})
	}
}

func TestToEnvoyFilter(t *testing.T) {
	cases := []string{
		"basic",
		"egress",
		"ingress",
		"request-header",
		"request-method",
		"simple-url",
		"gateway",
	}

	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			rl, err := readInput(tc)
			assert.NoError(t, err)

			expected, err := readExpected(tc)
			assert.NoError(t, err)

			got := ToEnvoyFilter(rl)
			assert.Equal(t, expected, got)
		})
	}
}

func readFile(filename string) ([]byte, error) {
	return os.ReadFile(path.Join(".", "testdata", filename))
}

func readInput(caseName string) (*policyv1alpha1.RateLimit, error) {
	b, err := os.ReadFile(path.Join("../../config/samples/", fmt.Sprintf("%s.yaml", caseName)))
	if err != nil {
		return nil, err
	}
	var rl policyv1alpha1.RateLimit
	err = yaml.Unmarshal(b, &rl)
	return &rl, err
}

func readExpected(caseName string) (*clientnetworkingv1alpha3.EnvoyFilter, error) {
	b, err := readFile(fmt.Sprintf("%s.yaml", caseName))
	if err != nil {
		return nil, err
	}
	var ef clientnetworkingv1alpha3.EnvoyFilter
	err = yaml.Unmarshal(b, &ef)
	return &ef, err
}

func TestBuildPatchStruct(t *testing.T) {
	_, err := buildPatchStruct(localRatelimit)
	assert.NoError(t, err)
}
