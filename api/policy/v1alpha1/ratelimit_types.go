/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RateLimitSpec defines the desired state of RateLimit
type RateLimitSpec struct {
	// Criteria used to select the specific set of pods/VMs on which this
	// `RateLimit` configuration should be applied. If omitted, the `RateLimit`
	// configuration will be applied to all workload instances in the same namespace.
	WorkloadSelector map[string]string `json:"workloadSelector"`
	// Traffic specifies the configuration of the sidecar for processing
	// inbound/outbound traffic to the attached workload instance.
	// +optional
	Traffic *TrafficSelector `json:"traffic,omitempty"`
	// HTTP Local rate limiting.
	// +optional
	HttpLocalRateLimit *HttpLocalRateLimit `json:"localRateLimit,omitempty"`
	// HTTP Global rate limiting.
	// +optional
	HttpGlobalRateLimit *HttpGlobalRateLimit `json:"globalRateLimit,omitempty"`
}

type TrafficSelector struct {
	// The network traffic direction to the attached workload instance(e.g. Inbound, Outbound, Gateway etc.).
	Direction TrafficDirection `json:"direction"`
	// The name of the outbound service to which should be rate limited.
	// The corresponding service can be a service in the service registry
	// (e.g., a Kubernetes or cloud foundry service) or a service specified
	// using a `ServiceEntry` or `VirtualService` configuration.
	// *Note for Kubernetes users*: When short names are used (e.g. "reviews"
	// instead of "reviews.default.svc.cluster.local"), Istio will interpret
	// the short name based on the namespace of the rule, not the service. A
	// rule in the "default" namespace containing a host "reviews" will be
	// interpreted as "reviews.default.svc.cluster.local", irrespective of
	// the actual namespace associated with the reviews service. To avoid
	// potential misconfigurations, it is recommended to always use fully
	// qualified domain names over short names.
	//
	// NOTE: THIS WILL BE IGNORED IF DIRECTION IS INBOUND
	//
	// +optional
	Host string `json:"host,omitempty"`
	// The inbound service port number, The outbound service port number
	Port uint32 `json:"port"`
}

// TrafficDirection allows selection of the network traffic direction to the attached workload instance.
type TrafficDirection string

const (
	// Inbound traffic of sidecar
	TrafficDirectionInbound TrafficDirection = "Inbound"

	// Outbound traffic of sidecar
	TrafficDirectionOutbound TrafficDirection = "Outbound"

	// Traffic of Gateway
	TrafficDirectionGateway TrafficDirection = "Gateway"
)

type HttpLocalRateLimit struct {
	// Rules of how the request will be performed.
	Rules []*RateLimitRule `json:"rules"`
}

type HttpGlobalRateLimit struct {
	Domain string `json:"domain"`
	// Specifies how the request will be matched.
	Match []*RateLimitMatch `json:"match"`
	// Configuration for an external rate limit service.
	Service *RateLimitService `json:"service"`
}

type RateLimitService struct {
	// Specifies the service that implements rate limit service.
	// The format is "<Hostname>". The <Hostname> is the full qualified host name in the Istio service
	// registry defined by the Kubernetes service or ServiceEntry. The <Namespace> is the namespace of the Kubernetes
	// service or ServiceEntry object, and can be omitted if the <Hostname> alone can decide the service unambiguously
	// (normally this means there is only 1 such host name in the service registry).
	//
	// Example: "rls.foo.svc.cluster.local" or "rls.example.com".
	Host string `json:"host"`
	// The behaviour in case the rate limiting service does not respond back.
	// When it is set to true, the proxy will not allow traffic in case of
	// communication failure between rate limiting service and the proxy.
	// +optional
	DenyOnFailed bool `json:"denyOnFailed,omitempty"`
	// Specifies the port of the service.
	Port uint32 `json:"port"`
	// The timeout in milliseconds for the rate limit service RPC. If not
	// set, this defaults to 20ms.
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

type RateLimitRule struct {
	// Specifies how the request will be matched.
	// If match is empty, policy will be used as default policy.
	// +optional
	Match []*RateLimitMatch `json:"match,omitempty"`
	// Policy will be used when the request matched.
	Policy *RatelimitPolicy `json:"policy"`
}

type RateLimitMatch struct {
	// Rate limit on request path.
	// +optional
	Url *UrlMatch `json:"url,omitempty"`
	// Rate limit on request source.
	// +optional
	Source *SourceMatch `json:"source,omitempty"`
	// Rate limit on request method.
	// +optional
	Method *MethodMatch `json:"method,omitempty"`
	// Rate limit on request header.
	// +optional
	Header *HeaderMatch `json:"header,omitempty"`
}

type UrlMatch struct {
	// Specifies how the path match will be performed(e.g. prefix, suffix, exact, regex, etc.).
	// Exist match type is not support.
	MatchType StringMatchType `json:"matchType"`
	// Path of the request.
	Path string `json:"path"`
	// If false, indicates the exact/prefix/suffix matching should be case insensitive.
	// This has no effect for the regex match.
	// For example, the matcher *data* will match both input string *Data* and *data* if set to false.
	// +optional
	CaseSensitive bool `json:"caseSensitive,omitempty"`
}

type StringMatchType string

const (
	// If specified, match will be performed whether the content is in the request.
	StringMatchTypeExist StringMatchType = "Exist"
	// If specified, match will be performed based on the value.
	StringMatchTypeExact StringMatchType = "Exact"
	// If specified, match will be performed based on the prefix of the content.
	StringMatchTypePrefix StringMatchType = "Prefix"
	// If specified, match will be performed based on the suffix of the content.
	StringMatchTypeSuffix StringMatchType = "Suffix"
	// If specified, this regex string is a regular expression in [RE2](https://github.com/google/re2) format.
	StringMatchTypeRegex StringMatchType = "Regex"
)

type SourceMatch struct {
	// IP will be performed in Classless Inter-Domain Routing format(e.g. 10.10.0.10/32, 10.10.0.0/16, etc.).
	// When RateLimitType is Local, the mask must be 32.
	Cidr string `json:"cidr"`
}

type MethodMatch struct {
	// Specifies the name of method in a request (e.g. GET, POST, etc.).
	Verb string `json:"verb"`
}

type HeaderMatch struct {
	// Specifies how the header match will be performed to route the request.
	MatchType StringMatchType `json:"matchType"`
	// Specifies the name of the header in the request.
	Key string `json:"key"`
	// If specified, header match will be performed based on the value of the header.
	// This should be empty when type is Exist.
	// +optional
	Value string `json:"value,omitempty"`
	// If false, indicates the exact/prefix/suffix/contains matching should be case insensitive.
	// This has no effect for the regex match.
	// For example, the matcher *data* will match both input string *Data* and *data* if set to false.
	// +optional
	CaseSensitive bool `json:"caseSensitive,omitempty"`
}

type RatelimitPolicy struct {
	// Burst is the maximum number of requests allowed to go through in the same arbitrarily small period of time.
	Burst int32 `json:"brust"`
	// The number of tokens added to the bucket during each fill interval.
	TokensPerFill int32 `json:"tokensPerFill"`
	// The fill interval that tokens are added to the bucket. During each fill interval
	// `TokensPerFill` are added to the bucket. The bucket will never contain more than
	// `Burst` tokens.
	Interval metav1.Duration `json:"interval"`
}

// RateLimitStatus defines the observed state of RateLimit
type RateLimitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope="Namespaced"
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RateLimit is the Schema for the ratelimits API
type RateLimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RateLimitSpec   `json:"spec,omitempty"`
	Status RateLimitStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RateLimitList contains a list of RateLimit
type RateLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimit `json:"items"`
}
