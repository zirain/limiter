package config

import (
	"fmt"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"istio.io/istio/pkg/test/util/assert"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
)

func TestHeaderGenerator(t *testing.T) {
	g := &HeaderMatchGenerator{}

	cases := []struct {
		name               string
		match              *policyv1alpha1.RateLimitMatch
		expectedAction     *routev3.RateLimit_Action
		expectedDescriptor *commonratelimitv3.RateLimitDescriptor_Entry
	}{
		{
			name: "exist",
			match: &policyv1alpha1.RateLimitMatch{
				Header: &policyv1alpha1.HeaderMatch{
					MatchType: policyv1alpha1.StringMatchTypeExist,
					Key:       ":exist",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "HEADER|Exist|:exist",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: ":exist",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_PresentMatch{
									PresentMatch: true,
								},
							},
						},
					},
				},
			},
			expectedDescriptor: &commonratelimitv3.RateLimitDescriptor_Entry{
				Key:   "header_match",
				Value: "HEADER|Exist|:exist",
			},
		},
		{
			name: "exact",
			match: &policyv1alpha1.RateLimitMatch{
				Header: &policyv1alpha1.HeaderMatch{
					MatchType: policyv1alpha1.StringMatchTypeExact,
					Key:       ":exact",
					Value:     "exact_value",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "HEADER|Exact|:exact",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: ":exact",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_Exact{
											Exact: "exact_value",
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
				Value: "HEADER|Exact|:exact",
			},
		},
		{
			name: "prefix",
			match: &policyv1alpha1.RateLimitMatch{
				Header: &policyv1alpha1.HeaderMatch{
					MatchType: policyv1alpha1.StringMatchTypePrefix,
					Key:       ":prefix",
					Value:     "prefix_value",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "HEADER|Prefix|:prefix",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: ":prefix",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_Prefix{
											Prefix: "prefix_value",
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
				Value: "HEADER|Prefix|:prefix",
			},
		},
		{
			name: "suffix",
			match: &policyv1alpha1.RateLimitMatch{
				Header: &policyv1alpha1.HeaderMatch{
					MatchType: policyv1alpha1.StringMatchTypeSuffix,
					Key:       ":suffix",
					Value:     "suffix_value",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "HEADER|Suffix|:suffix",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: ":suffix",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_Suffix{
											Suffix: "suffix_value",
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
				Value: "HEADER|Suffix|:suffix",
			},
		},
		{
			name: "regex",
			match: &policyv1alpha1.RateLimitMatch{
				Header: &policyv1alpha1.HeaderMatch{
					MatchType: policyv1alpha1.StringMatchTypeRegex,
					Key:       ":regex",
					Value:     "regex_value",
				},
			},
			expectedAction: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "HEADER|Regex|:regex",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: ":regex",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_SafeRegex{
											SafeRegex: &matcherv3.RegexMatcher{
												EngineType: &matcherv3.RegexMatcher_GoogleRe2{},
												Regex:      "regex_value",
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
				Value: "HEADER|Regex|:regex",
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
