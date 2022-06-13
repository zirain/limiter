package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	prometheusApi "github.com/prometheus/client_golang/api"
	prometheusApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/zirain/limiter/client-go/generated/clientset/versioned"
	"github.com/zirain/limiter/test/podexec"
	"github.com/zirain/limiter/test/util"
)

var (
	kubeClient    *kubernetes.Clientset
	limiterClient *clientset.Clientset
	istioClient   *istioclient.Clientset
	execClient    *podexec.Client
	prom          prometheusApiV1.API
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "E2E test")
}

func TestMain(m *testing.M) {
	home := util.HomeDir()
	configPath := util.KubeconfigPath(home)
	config, _ := clientcmd.BuildConfigFromFlags(util.MasterURL(), configPath)
	kubeClient = kubernetes.NewForConfigOrDie(config)
	istioClient = istioclient.NewForConfigOrDie(config)
	limiterClient = clientset.NewForConfigOrDie(config)
	execClient = podexec.NewForConfigOrDie(config)

	svc, _ := kubeClient.CoreV1().Services("istio-system").Get(context.Background(), "prometheus-elb", metav1.GetOptions{})
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		address := fmt.Sprintf("http://%s:9090", ingress.IP)
		c, _ := prometheusApi.NewClient(prometheusApi.Config{Address: address})
		prom = prometheusApiV1.NewAPI(c)
		break
	}

	os.Exit(m.Run())
}
