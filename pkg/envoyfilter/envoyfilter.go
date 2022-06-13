package envoyfilter

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	ratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	gogojsonpb "github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"istio.io/api/networking/v1alpha3"
	clientnetworkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/networking/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
	"github.com/zirain/limiter/pkg/conversion"
)

const (
	hcmFilter      = "envoy.filters.network.http_connection_manager"
	statPrefix     = "ratelimiter"
	localRatelimit = `{"name": "envoy.filters.http.local_ratelimit", 
	"typed_config": { 
		"@type": "type.googleapis.com/udpa.type.v1.TypedStruct", 
		"type_url": "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit",
		"value": {
			"stat_prefix": "ratelimiter"
		}}}`
)

var infinitePolicy = &policyv1alpha1.RatelimitPolicy{
	Burst:         math.MaxInt32,
	TokensPerFill: math.MaxInt32,
	Interval: metav1.Duration{
		Duration: 60 * time.Second,
	},
}

func ToEnvoyFilter(ratelimit *policyv1alpha1.RateLimit) *clientnetworkingv1alpha3.EnvoyFilter {
	r := buildRouteComponent(ratelimit.Spec.HttpLocalRateLimit.Rules)
	val, _ := generateValue(r)

	vHostName := vhostName(ratelimit)
	insertval, _ := buildPatchStruct(localRatelimit)
	ef := &clientnetworkingv1alpha3.EnvoyFilter{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "EnvoyFilter",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ratelimit.Name,
			Namespace: ratelimit.Namespace,
		},
		Spec: v1alpha3.EnvoyFilter{
			WorkloadSelector: &v1alpha3.WorkloadSelector{
				Labels: ratelimit.Spec.WorkloadSelector,
			},
			ConfigPatches: []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
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
						Value:     insertval,
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
						Value:     val,
					},
				},
			},
		},
	}
	return ef
}

func generateValue(message proto.Message) (*types.Struct, error) {
	var (
		buf []byte
		err error
	)

	if buf, err = protojson.Marshal(message); err != nil {
		return nil, err
	}

	var value = &types.Struct{}
	if err = (&gogojsonpb.Unmarshaler{AllowUnknownFields: false}).Unmarshal(bytes.NewBuffer(buf), value); err != nil {
		return nil, err
	}

	return value, nil
}

func matchContext(ratelimit *policyv1alpha1.RateLimit) v1alpha3.EnvoyFilter_PatchContext {
	if ratelimit.Spec.Egress != nil {
		return v1alpha3.EnvoyFilter_SIDECAR_OUTBOUND

	}

	return v1alpha3.EnvoyFilter_SIDECAR_INBOUND
}

func vhostName(ratelimit *policyv1alpha1.RateLimit) string {
	if ratelimit.Spec.Ingress != nil && ratelimit.Spec.Ingress.Port != nil {
		return fmt.Sprintf("inbound|http|%d", *ratelimit.Spec.Ingress.Port)
	}

	egress := ratelimit.Spec.Egress
	if egress != nil {
		return fmt.Sprintf("%s:%d", egress.Host, egress.Port)
	}

	return ""
}

func buildRouteComponent(rules []*policyv1alpha1.RateLimitRule) *routev3.Route {
	return &routev3.Route{
		Action: &routev3.Route_Route{
			Route: &routev3.RouteAction{
				RateLimits: buildRateLimitActions(rules),
			},
		},
		TypedPerFilterConfig: map[string]*any.Any{
			"envoy.filters.http.local_ratelimit": util.MessageToAny(buildLocalRateLimit(rules)),
		},
	}
}

func buildRateLimitActions(rules []*policyv1alpha1.RateLimitRule) []*routev3.RateLimit {
	ratelimitActions := []*routev3.RateLimit{}
	for _, r := range rules {
		if len(r.Match) == 0 {
			continue
		}

		actions := make([]*routev3.RateLimit_Action, 0)

		for _, match := range r.Match {
			actions = append(actions, conversion.ToRateLimitAction(match))
		}

		ratelimitActions = append(ratelimitActions, &routev3.RateLimit{
			Actions: actions,
		})
	}

	return ratelimitActions
}

func buildLocalRateLimit(rules []*policyv1alpha1.RateLimitRule) *ratelimitv3.LocalRateLimit {
	p := getDefaultPolicy(rules)

	return &ratelimitv3.LocalRateLimit{
		StatPrefix:  statPrefix,
		TokenBucket: conversion.ToTokenBucket(p),
		FilterEnabled: &corev3.RuntimeFractionalPercent{
			RuntimeKey: "filter_enabled",
			DefaultValue: &typev3.FractionalPercent{
				Numerator:   100,
				Denominator: typev3.FractionalPercent_HUNDRED,
			},
		},
		FilterEnforced: &corev3.RuntimeFractionalPercent{
			RuntimeKey: "filter_enforced",
			DefaultValue: &typev3.FractionalPercent{
				Numerator:   100,
				Denominator: typev3.FractionalPercent_HUNDRED,
			},
		},
		ResponseHeadersToAdd: []*corev3.HeaderValueOption{
			{
				Append: wrapperspb.Bool(false),
				Header: &corev3.HeaderValue{
					Key:   "x-local-rate-limit",
					Value: "true",
				},
			},
		},
		Descriptors: buildLocalRateLimitDescriptors(rules),
	}
}

func getDefaultPolicy(rules []*policyv1alpha1.RateLimitRule) *policyv1alpha1.RatelimitPolicy {
	for _, r := range rules {
		if len(r.Match) == 0 {
			return r.Policy
		}
	}

	return infinitePolicy
}

func buildLocalRateLimitDescriptors(rules []*policyv1alpha1.RateLimitRule) []*commonratelimitv3.LocalRateLimitDescriptor {
	descriptors := make([]*commonratelimitv3.LocalRateLimitDescriptor, 0, len(rules))
	for _, r := range rules {
		if len(r.Match) == 0 {
			continue
		}
		descriptors = append(descriptors, conversion.ToLocalRateLimitDescriptor(r))
	}
	return descriptors
}

func buildPatchStruct(config string) (*types.Struct, error) {
	val := &types.Struct{}
	err := gogojsonpb.Unmarshal(strings.NewReader(config), val)
	return val, err
}
