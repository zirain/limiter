package config

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
)

const (
	pathHeaderName = ":path"
	// HeaderValueMatch descriptor key is const before https://github.com/envoyproxy/envoy/pull/20321.
	headerMatchDescriptorKey = "header_match"
)

type UrlMatchGenerator struct {
}

func (url *UrlMatchGenerator) GeneratorAction(match *policyv1alpha1.RateLimitMatch) (*routev3.RateLimit_Action, error) {
	if match.Url == nil {
		return nil, errors.New("url is nil")
	}

	rule := match.Url
	switch rule.MatchType {
	case policyv1alpha1.StringMatchTypePrefix:
		return &routev3.RateLimit_Action{
			ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
				HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
					DescriptorValue: getUrlMatchValue(rule),
					Headers: []*routev3.HeaderMatcher{
						{
							Name: pathHeaderName,
							HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
								StringMatch: &matcherv3.StringMatcher{
									MatchPattern: &matcherv3.StringMatcher_Prefix{
										Prefix: rule.Path,
									},
									IgnoreCase: true,
								},
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
					DescriptorValue: getUrlMatchValue(rule),
					Headers: []*routev3.HeaderMatcher{
						{
							Name: pathHeaderName,
							HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
								StringMatch: &matcherv3.StringMatcher{
									MatchPattern: &matcherv3.StringMatcher_Exact{
										Exact: rule.Path,
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
					DescriptorValue: getUrlMatchValue(rule),
					Headers: []*routev3.HeaderMatcher{
						{
							Name: pathHeaderName,
							HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
								StringMatch: &matcherv3.StringMatcher{
									MatchPattern: &matcherv3.StringMatcher_SafeRegex{
										SafeRegex: &matcherv3.RegexMatcher{
											EngineType: &matcherv3.RegexMatcher_GoogleRe2{},
											Regex:      rule.Path,
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
		return nil, errors.New("unsupport match type")
	}
}

func (url *UrlMatchGenerator) GenerateDescriptor(match *policyv1alpha1.RateLimitMatch) (*commonratelimitv3.RateLimitDescriptor_Entry, error) {
	if match.Url == nil {
		return nil, errors.New("url is nil")
	}

	return &commonratelimitv3.RateLimitDescriptor_Entry{
		Key:   headerMatchDescriptorKey,
		Value: getUrlMatchValue(match.Url),
	}, nil
}

func getUrlMatchValue(url *policyv1alpha1.UrlMatch) string {
	return fmt.Sprintf("URL|%s|%s", url.MatchType, url.Path)
}
