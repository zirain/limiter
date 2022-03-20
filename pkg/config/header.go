package config

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	policyv1alpha1 "github.com/zirain/limiter/api/v1alpha1"
)

type HeaderMatchGenerator struct {
}

func (header *HeaderMatchGenerator) GeneratorAction(match *policyv1alpha1.RateLimitMatch) (*routev3.RateLimit_Action, error) {
	if match.Header == nil {
		return nil, errors.New("requestHeader is nil")
	}

	rule := match.Header
	switch rule.MatchType {
	case policyv1alpha1.StringMatchTypeExist:
		return &routev3.RateLimit_Action{
			ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
				HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
					DescriptorValue: getHeaderMatchValue(rule),
					Headers: []*routev3.HeaderMatcher{
						{
							Name: rule.Key,
							HeaderMatchSpecifier: &routev3.HeaderMatcher_PresentMatch{
								PresentMatch: true,
							},
						},
					},
				},
			},
		}, nil
	case policyv1alpha1.StringMatchTypeExact:
		return &routev3.RateLimit_Action{
			ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
				HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
					DescriptorValue: getHeaderMatchValue(rule),
					Headers: []*routev3.HeaderMatcher{
						{
							Name: rule.Key,
							HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
								StringMatch: &matcherv3.StringMatcher{
									MatchPattern: &matcherv3.StringMatcher_Exact{
										Exact: rule.Value,
									},
									IgnoreCase: true,
								},
							},
						},
					},
				},
			},
		}, nil
	case policyv1alpha1.StringMatchTypePrefix:
		return &routev3.RateLimit_Action{
			ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
				HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
					DescriptorValue: getHeaderMatchValue(rule),
					Headers: []*routev3.HeaderMatcher{
						{
							Name: rule.Key,
							HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
								StringMatch: &matcherv3.StringMatcher{
									MatchPattern: &matcherv3.StringMatcher_Prefix{
										Prefix: rule.Value,
									},
									IgnoreCase: true,
								},
							},
						},
					},
				},
			},
		}, nil
	case policyv1alpha1.StringMatchTypeSuffix:
		return &routev3.RateLimit_Action{
			ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
				HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
					DescriptorValue: getHeaderMatchValue(rule),
					Headers: []*routev3.HeaderMatcher{
						{
							Name: rule.Key,
							HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
								StringMatch: &matcherv3.StringMatcher{
									MatchPattern: &matcherv3.StringMatcher_Suffix{
										Suffix: rule.Value,
									},
									IgnoreCase: true,
								},
							},
						},
					},
				},
			},
		}, nil
	case policyv1alpha1.StringMatchTypeRegex:
		return &routev3.RateLimit_Action{
			ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
				HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
					DescriptorValue: getHeaderMatchValue(rule),
					Headers: []*routev3.HeaderMatcher{
						{
							Name: rule.Key,
							HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
								StringMatch: &matcherv3.StringMatcher{
									MatchPattern: &matcherv3.StringMatcher_SafeRegex{
										SafeRegex: &matcherv3.RegexMatcher{
											EngineType: &matcherv3.RegexMatcher_GoogleRe2{},
											Regex:      rule.Value,
										},
									},
								},
							},
						},
					},
				},
			},
		}, nil
	default:
		return nil, errors.New("unsupport header match type")
	}
}

func (header *HeaderMatchGenerator) GenerateDescriptor(match *policyv1alpha1.RateLimitMatch) (*commonratelimitv3.RateLimitDescriptor_Entry, error) {
	if match.Header == nil {
		return nil, errors.New("urlrequestHeader is nil")
	}

	rule := match.Header
	return &commonratelimitv3.RateLimitDescriptor_Entry{
		Key:   headerMatchDescriptorKey,
		Value: getHeaderMatchValue(rule),
	}, nil
}

func getHeaderMatchValue(url *policyv1alpha1.HeaderMatch) string {
	return fmt.Sprintf("HEADER|%s|%s", url.MatchType, url.Key)
}
