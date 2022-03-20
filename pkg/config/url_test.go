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

func TestUrlGenerator(t *testing.T) {
	url := &UrlMatchGenerator{}

	cases := []struct {
		name               string
		match              *policyv1alpha1.RateLimitMatch
		expectedAction     *routev3.RateLimit_Action
		expectedDescriptor *commonratelimitv3.RateLimitDescriptor_Entry
	}{
		{
			name: "exact",
			match: &policyv1alpha1.RateLimitMatch{
				Url: &policyv1alpha1.UrlMatch{
					MatchType: policyv1alpha1.StringMatchTypeExact,
					Path:      "/exact",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "URL|Exact|/exact",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: pathHeaderName,
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "/exact",
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
				Value: "URL|Exact|/exact",
			},
		},
		{
			name: "prefix",
			match: &policyv1alpha1.RateLimitMatch{
				Url: &policyv1alpha1.UrlMatch{
					MatchType: policyv1alpha1.StringMatchTypePrefix,
					Path:      "/prefix",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "URL|Prefix|/prefix",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: pathHeaderName,
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_Prefix{
											Prefix: "/prefix",
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
				Value: "URL|Prefix|/prefix",
			},
		},
		{
			name: "regex",
			match: &policyv1alpha1.RateLimitMatch{
				Url: &policyv1alpha1.UrlMatch{
					MatchType: policyv1alpha1.StringMatchTypeRegex,
					Path:      "/regex",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "URL|Regex|/regex",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: pathHeaderName,
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_SafeRegex{
											SafeRegex: &matcherv3.RegexMatcher{
												EngineType: &matcherv3.RegexMatcher_GoogleRe2{},
												Regex:      "/regex",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedDescriptor: &commonratelimitv3.RateLimitDescriptor_Entry{
				Key:   "header_match",
				Value: "URL|Regex|/regex",
			},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s-action", tc.name), func(t *testing.T) {
			action, err := url.GeneratorAction(tc.match)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedAction, action)
		})

		t.Run(fmt.Sprintf("%s-descriptor", tc.name), func(t *testing.T) {
			descriptor, err := url.GenerateDescriptor(tc.match)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedDescriptor, descriptor)
		})
	}
}
