package conversion

import (
	"testing"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"istio.io/istio/pkg/test/util/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
)

var (
	policy = &policyv1alpha1.RatelimitPolicy{
		Burst: 100,
		Interval: metav1.Duration{
			Duration: 60 * time.Second,
		},
	}
	bucket = &typev3.TokenBucket{
		MaxTokens: 100,
		TokensPerFill: &wrapperspb.UInt32Value{
			Value: 100,
		},
		FillInterval: durationpb.New(60 * time.Second),
	}
)

func TestToTokenBucket(t *testing.T) {
	cases := []struct {
		policy   *policyv1alpha1.RatelimitPolicy
		expected *typev3.TokenBucket
	}{
		{
			policy: policy,
			expected: &typev3.TokenBucket{
				MaxTokens: 100,
				TokensPerFill: &wrapperspb.UInt32Value{
					Value: 100,
				},
				FillInterval: durationpb.New(60 * time.Second),
			},
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := ToTokenBucket(tc.policy)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestToRateLimitAction(t *testing.T) {
	cases := []struct {
		name     string
		match    *policyv1alpha1.RateLimitMatch
		expected *routev3.RateLimit_Action
	}{
		{
			name: "request-header",
			match: &policyv1alpha1.RateLimitMatch{
				Header: &policyv1alpha1.HeaderMatch{
					MatchType: policyv1alpha1.StringMatchTypeExist,
					Key:       "x-real-user",
					Value:     "fake-user",
				},
			},
			expected: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "HEADER|Exist|x-real-user",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: "x-real-user",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_PresentMatch{
									PresentMatch: true,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "request-url",
			match: &policyv1alpha1.RateLimitMatch{
				Url: &policyv1alpha1.UrlMatch{
					MatchType: policyv1alpha1.StringMatchTypeRegex,
					Path:      "/ip.*",
				},
			},
			expected: &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
					HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
						DescriptorValue: "URL|Regex|/ip.*",
						Headers: []*routev3.HeaderMatcher{
							{
								Name: ":path",
								HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
									StringMatch: &matcherv3.StringMatcher{
										MatchPattern: &matcherv3.StringMatcher_SafeRegex{
											SafeRegex: &matcherv3.RegexMatcher{
												EngineType: &matcherv3.RegexMatcher_GoogleRe2{},
												Regex:      "/ip.*",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ToRateLimitAction(tc.match)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestToLocalRateLimitDescriptor(t *testing.T) {
	cases := []struct {
		name     string
		rule     *policyv1alpha1.RateLimitRule
		expected *commonratelimitv3.LocalRateLimitDescriptor
	}{
		{
			name: "request-url",
			rule: &policyv1alpha1.RateLimitRule{
				Match: []*policyv1alpha1.RateLimitMatch{
					{
						Url: &policyv1alpha1.UrlMatch{
							MatchType: policyv1alpha1.StringMatchTypeRegex,
							Path:      "/ip.*",
						},
					},
				},
				Policy: policy,
			},
			expected: &commonratelimitv3.LocalRateLimitDescriptor{
				Entries: []*commonratelimitv3.RateLimitDescriptor_Entry{
					{
						Key:   "header_match",
						Value: "URL|Regex|/ip.*",
					},
				},
				TokenBucket: bucket,
			},
		},
		{
			name: "request-header-exist",
			rule: &policyv1alpha1.RateLimitRule{
				Match: []*policyv1alpha1.RateLimitMatch{
					{
						Header: &policyv1alpha1.HeaderMatch{
							MatchType: policyv1alpha1.StringMatchTypeExist,
							Key:       "x-real-user",
							Value:     "fake-user",
						},
					},
				},
				Policy: policy,
			},
			expected: &commonratelimitv3.LocalRateLimitDescriptor{
				Entries: []*commonratelimitv3.RateLimitDescriptor_Entry{
					{
						Key:   "header_match",
						Value: "HEADER|Exist|x-real-user",
					},
				},
				TokenBucket: bucket,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ToLocalRateLimitDescriptor(tc.rule)
			assert.Equal(t, tc.expected, got)
		})
	}
}
