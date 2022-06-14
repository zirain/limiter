package util

import (
	"context"
	"time"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	interval = time.Second
	timeout  = time.Minute
)

func WaitEnvoyFilterExists(client *istioclient.Clientset, name, namespace string) error {
	return wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		_, err = client.NetworkingV1alpha3().EnvoyFilters(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		return true, nil
	})
}
