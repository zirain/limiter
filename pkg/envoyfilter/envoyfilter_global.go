package envoyfilter

import (
	"fmt"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"istio.io/api/networking/v1alpha3"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
	"github.com/zirain/limiter/pkg/conversion"
)

var (
	defaultTimeout          = time.Millisecond * 20
	globalRatelimitTemplate = `{
	"name": "envoy.filters.http.ratelimit",
	"typed_config": {
		"@type": "type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit",
		"domain": "%s",
		"failure_mode_deny": %v,
		"rate_limit_service": {
			"grpc_service": {
				"envoy_grpc": {
					"authority": "%s",
					"cluster_name": "%s"
				},
				"timeout": "%s"
			},
			"transport_api_version": "V3"
		}
	}
}
`
)

func globalConfigPatches(ratelimit *policyv1alpha1.RateLimit) []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch {

	r := buildRouteComponent(ratelimit.Spec.HttpGlobalRateLimit.Match)
	routeCfg, _ := generateValue(r)

	vHostName := vhostName(ratelimit)
	ratelimitFilterCfg, _ := buildPatchStruct(ratelimitTypedConfig(ratelimit.Spec.HttpGlobalRateLimit.Domain, ratelimit.Spec.HttpGlobalRateLimit.Service))
	return []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
		{
			ApplyTo: v1alpha3.EnvoyFilter_HTTP_FILTER,
			Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
				Context: matchContext(ratelimit),
				ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
					Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
						FilterChain: &v1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
							Filter: &v1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
								Name: hcmFilter,
							},
						},
					},
				},
			},
			Patch: &v1alpha3.EnvoyFilter_Patch{
				Operation: v1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
				Value:     ratelimitFilterCfg,
			},
		},
		{
			ApplyTo: v1alpha3.EnvoyFilter_HTTP_ROUTE,
			Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
				Context: matchContext(ratelimit),
				ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_RouteConfiguration{
					RouteConfiguration: &v1alpha3.EnvoyFilter_RouteConfigurationMatch{
						Vhost: &v1alpha3.EnvoyFilter_RouteConfigurationMatch_VirtualHostMatch{
							Name: vHostName,
							Route: &v1alpha3.EnvoyFilter_RouteConfigurationMatch_RouteMatch{
								Action: v1alpha3.EnvoyFilter_RouteConfigurationMatch_RouteMatch_ROUTE,
							},
						},
					},
				},
			},
			Patch: &v1alpha3.EnvoyFilter_Patch{
				Operation: v1alpha3.EnvoyFilter_Patch_MERGE,
				Value:     routeCfg,
			},
		},
	}
}

func ratelimitTypedConfig(domain string, service *policyv1alpha1.RateLimitService) string {
	authority := service.Host
	clusterName := fmt.Sprintf("outbound|%d||%s", service.Port, service.Host)
	timeout := defaultTimeout
	if service.Timeout != nil {
		timeout = service.Timeout.Duration
	}
	return fmt.Sprintf(globalRatelimitTemplate, domain, service.DenyOnFailed, authority, clusterName, timeout)
}

func buildRouteComponent(match []*policyv1alpha1.RateLimitMatch) *routev3.Route {
	ratelimitActions := []*routev3.RateLimit{}

	actions := make([]*routev3.RateLimit_Action, 0)

	for _, match := range match {
		actions = append(actions, conversion.ToRateLimitAction(match))
	}

	ratelimitActions = append(ratelimitActions, &routev3.RateLimit{
		Actions: actions,
	})

	return &routev3.Route{
		Action: &routev3.Route_Route{
			Route: &routev3.RouteAction{
				RateLimits: ratelimitActions,
			},
		},
	}
}
