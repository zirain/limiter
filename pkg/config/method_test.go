package config

import (
	"fmt"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"istio.io/istio/pkg/test/util/assert"

	policyv1alpha1 "github.com/zirain/limiter/api/v1alpha1"
)

func TestMethodGenerator(t *testing.T) {
	g := &MethodMatchGenerator{}

	cases := []struct {
		name               string
		match              *policyv1alpha1.RateLimitMatch
		expectedAction     *routev3.RateLimit_Action
		expectedDescriptor *commonratelimitv3.RateLimitDescriptor_Entry
	}{
		{
			name: "GET",
			match: &policyv1alpha1.RateLimitMatch{
				Method: &policyv1alpha1.MethodMatch{
					Verb: "GET",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "METHOD|GET",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: ":method",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "GET",
										},
										IgnoreCase: true,
									},
								},
							},
						},
					},
				},
			},
			expectedDescriptor: &commonratelimitv3.RateLimitDescriptor_Entry{
				Key:   "header_match",
				Value: "METHOD|GET",
			},
		},
		{
			name: "POST",
			match: &policyv1alpha1.RateLimitMatch{
				Method: &policyv1alpha1.MethodMatch{
					Verb: "POST",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "METHOD|POST",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: ":method",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "POST",
										},
										IgnoreCase: true,
									},
								},
							},
						},
					},
				},
			},
			expectedDescriptor: &commonratelimitv3.RateLimitDescriptor_Entry{
				Key:   "header_match",
				Value: "METHOD|POST",
			},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s-action", tc.name), func(t *testing.T) {
			action, err := g.GeneratorAction(tc.match)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAction, action)
		})

		t.Run(fmt.Sprintf("%s-descriptor", tc.name), func(t *testing.T) {
			descriptor, err := g.GenerateDescriptor(tc.match)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedDescriptor, descriptor)
		})
	}
}
