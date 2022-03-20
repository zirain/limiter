package config

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	policyv1alpha1 "github.com/zirain/limiter/api/v1alpha1"
)

const (
	methodHeaderName = ":method"
)

type MethodMatchGenerator struct {
}

func (method *MethodMatchGenerator) GeneratorAction(match *policyv1alpha1.RateLimitMatch) (*routev3.RateLimit_Action, error) {
	if match.Method.Verb == "" {
		return nil, errors.New("request method is nil")
	}

	rule := match.Method
	return &routev3.RateLimit_Action{
		ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
			HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
				DescriptorValue: getMethodMatchValue(rule),
				Headers: []*routev3.HeaderMatcher{
					{
						Name: methodHeaderName,
						HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
							StringMatch: &matcherv3.StringMatcher{
								MatchPattern: &matcherv3.StringMatcher_Exact{
									Exact: rule.Verb,
								},
								IgnoreCase: true,
							},
						},
					},
				},
			},
		},
	}, nil
}

func (method *MethodMatchGenerator) GenerateDescriptor(match *policyv1alpha1.RateLimitMatch) (*commonratelimitv3.RateLimitDescriptor_Entry, error) {
	if match.Method.Verb == "" {
		return nil, errors.New("urlrequestHeader is nil")
	}

	return &commonratelimitv3.RateLimitDescriptor_Entry{
		Key:   headerMatchDescriptorKey,
		Value: getMethodMatchValue(match.Method),
	}, nil
}

func getMethodMatchValue(method *policyv1alpha1.MethodMatch) string {
	return fmt.Sprintf("METHOD|%s", method.Verb)
}
