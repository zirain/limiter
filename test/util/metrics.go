package util

import "github.com/zirain/limiter/test/podexec"

func RateLimitedCount(c *podexec.Client, podName, podNamespace string) (int, error) {
	v, err := c.SidecarStats(podName, podNamespace, "envoy_ratelimiter_http_local_rate_limit_rate_limited")
	if err != nil {
		return 0, err
	}

	return int(v), nil
}
