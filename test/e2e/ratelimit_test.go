package e2e

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

		err = util.WaitEnvoyFilterExists(istioClient, "ratelimit-sample", namespace)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		podList, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		for _, pod := range podList.Items {
			if strings.HasPrefix(pod.Name, "fortio-deploy") {
				fortioPod = pod.Name
			}

			if strings.HasPrefix(pod.Name, "httpbin") {
				httpbinPod = pod.Name
			}
		}
	})

	ginkgo.It("Basic demo", func() {
		preCount, err := util.QueryRateLimitedCount(prom, httpbinPod)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		// TODO: find better way for waiting EnvoyFilter
		time.Sleep(5 * time.Second)

		// do fortio load test
		log.Printf("start request")
		err = util.FortioLoad(*execClient, fortioPod, requestCount, httpbinURL)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		postCount, err := util.RateLimitedCount(execClient, httpbinPod, namespace)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

		expectedLimitedCount := postCount - preCount
		gomega.Expect(expectedLimitedCount).To(gomega.Equal(limitedCount))

	})

	ginkgo.AfterEach(func() {
		err := util.DeleteRatelimit(limiterClient, "basic", namespace)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
})
