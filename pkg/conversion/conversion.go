package conversion

import (
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
)

var local = NewLocalConverter()

func ToTokenBucket(policy *policyv1alpha1.RatelimitPolicy) *typev3.TokenBucket {
	return local.ToTokenBucket(policy)
}

func ToRateLimitAction(match *policyv1alpha1.RateLimitMatch) *routev3.RateLimit_Action {
	return local.ToAction(match)
}

func ToLocalRateLimitDescriptor(rule *policyv1alpha1.RateLimitRule) *commonratelimitv3.LocalRateLimitDescriptor {
	return local.ToDescriptor(rule)
}
