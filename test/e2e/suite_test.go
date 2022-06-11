package e2e

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zirain/limiter/test/util"
)

var kubeClient *kubernetes.Clientset

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E test")
}

func TestMain(m *testing.M) {
	home := util.HomeDir()
	configPath := util.KubeconfigPath(home)
	config, _ := clientcmd.BuildConfigFromFlags(util.MasterURL(), configPath)
	kubeClient = kubernetes.NewForConfigOrDie(config)
	os.Exit(m.Run())
}
