package conversion

import (
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	commonratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	policyv1alpha1 "github.com/zirain/limiter/api/policy/v1alpha1"
	"github.com/zirain/limiter/pkg/config"
)

type Converter interface {
	ToTokenBucket(policy *policyv1alpha1.RatelimitPolicy) *typev3.TokenBucket
	ToAction(match *policyv1alpha1.RateLimitMatch) *routev3.RateLimit_Action
	ToDescriptor(rule *policyv1alpha1.RateLimitRule) *commonratelimitv3.LocalRateLimitDescriptor
}

type LocalConverter struct {
	generators map[actionType]config.Generator
}

var (
	generators = map[actionType]config.Generator{
		Url:    &config.UrlMatchGenerator{},
		Header: &config.HeaderMatchGenerator{},
		Method: &config.MethodMatchGenerator{},
	}
)

func NewLocalConverter() Converter {
	conv := &LocalConverter{
		generators: generators,
	}

	return conv
}

func (conv *LocalConverter) ToTokenBucket(policy *policyv1alpha1.RatelimitPolicy) *typev3.TokenBucket {
	return &typev3.TokenBucket{
		MaxTokens: uint32(policy.Burst),
		TokensPerFill: &wrapperspb.UInt32Value{
			Value: uint32(policy.Burst),
		},
		FillInterval: durationpb.New(policy.Interval.Duration),
	}
}

func (conv *LocalConverter) ToAction(match *policyv1alpha1.RateLimitMatch) *routev3.RateLimit_Action {
	actionType := getActionType(match)
	action, _ := conv.generators[actionType].GeneratorAction(match)
	return action
}

func (conv *LocalConverter) ToDescriptor(rule *policyv1alpha1.RateLimitRule) *commonratelimitv3.LocalRateLimitDescriptor {
	descriptor := &commonratelimitv3.LocalRateLimitDescriptor{
		Entries:     make([]*commonratelimitv3.RateLimitDescriptor_Entry, 0),
		TokenBucket: conv.ToTokenBucket(rule.Policy),
	}

	for _, m := range rule.Match {
		actionType := getActionType(m)
		e, _ := conv.generators[actionType].GenerateDescriptor(m)

		descriptor.Entries = append(descriptor.Entries, e)
	}

	return descriptor
}

type actionType int32

const (
	Url actionType = iota
	Header
	Method
	Source
)

func getActionType(match *policyv1alpha1.RateLimitMatch) actionType {
	var actionType actionType
	if match.Url != nil {
		actionType = Url
	} else if match.Header != nil {
		actionType = Header
	} else if match.Method != nil {
		actionType = Method
	} else if match.Source != nil {
		actionType = Source
	}

	return actionType
}
