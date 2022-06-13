package e2e

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/zirain/limiter/test/util"
)

const (
	httpbinURL   = "httpbin:8000/ip"
	requestCount = 15
	limitedCount = 5
)

var (
	namespace  = "default"
	fortioPod  = "fortio-deploy-6cf68cc4c-jglfs"
	httpbinPod = "httpbin-847f64cc8d-2kk9v"
)

var _ = ginkgo.Describe("RateLimit", func() {
	ginkgo.BeforeEach(func() {
		err := util.ApplyRatelimit(limiterClient, "basic", namespace)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})

	ginkgo.It("Basic demo", func() {
		preCount, err := util.QueryRateLimitedCount(prom, httpbinPod)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// do fortio load test
		err = util.FortioLoad(*execClient, fortioPod, requestCount, httpbinURL)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// wait prometheus scrape
		time.Sleep(15 * time.Second)

		postCount, err := util.QueryRateLimitedCount(prom, httpbinPod)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		expectedLimitedCount := postCount - preCount
		gomega.Expect(expectedLimitedCount).To(gomega.Equal(limitedCount))

	})

	ginkgo.BeforeEach(func() {
		err := util.DeleteRatelimit(limiterClient, "basic", namespace)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
})
