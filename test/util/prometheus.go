package util

import (
	"context"
	"fmt"
	"time"

	prometheusApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func QueryRateLimitedCount(prom prometheusApiV1.API, podName string) (int, error) {
	q := fmt.Sprintf("envoy_ratelimiter_http_local_rate_limit_rate_limited{pod=\"%s\"}", podName)
	v, _, err := prom.Query(context.Background(), q, time.Now())
	if err != nil {
		return 0, err
	}

	val, ok := v.(model.Vector)
	if !ok {
		return 0, fmt.Errorf("assert vector metric fail")
	}

	if len(val) == 0 {
		return 0, nil
	}

	return int(val[0].Value), nil
}
