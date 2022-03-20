package config

import (
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"

	policyv1alpha1 "github.com/zirain/limiter/api/v1alpha1"
)

type Generator interface {
	GeneratorAction(match *policyv1alpha1.RateLimitMatch) (*routev3.RateLimit_Action, error)
	GenerateDescriptor(match *policyv1alpha1.RateLimitMatch) (*commonratelimitv3.RateLimitDescriptor_Entry, error)
}
