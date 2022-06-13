package podexec

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"

	v1 "k8s.io/api/core/v1"
	spdyStream "k8s.io/apimachinery/pkg/util/httpstream/spdy"
	"k8s.io/client-go/kubernetes"
	kubescheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/transport/spdy"
)

type Client struct {
	config *rest.Config
	client kubernetes.Interface
}

func NewForConfigOrDie(config *rest.Config) *Client {
	return &Client{
		config: config,
		client: kubernetes.NewForConfigOrDie(config),
	}
}

func (c *Client) PodExecCommands(podName, podNamespace, container string, commands []string) (stdout, stderr string, err error) {
	defer func() {
		if err != nil {
			if len(stderr) > 0 {
				err = fmt.Errorf("error exec'ing into %s/%s %s container: %v\n%s",
					podNamespace, podName, container, err, stderr)
			} else {
				err = fmt.Errorf("error exec'ing into %s/%s %s container: %v",
					podNamespace, podName, container, err)
			}
		}
	}()

	req := c.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(podNamespace).
		SubResource("exec").
		Param("container", container).
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   commands,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, kubescheme.ParameterCodec)

	wrapper, upgrader, err := roundTripperFor(c.config)
	if err != nil {
		return "", "", err
	}
	exec, err := remotecommand.NewSPDYExecutorForTransports(wrapper, upgrader, "POST", req.URL())
	if err != nil {
		return "", "", err
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
		Tty:    false,
	})

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	return
}

// roundTripperFor creates a SPDY upgrader that will work over custom transports.
func roundTripperFor(restConfig *rest.Config) (http.RoundTripper, spdy.Upgrader, error) {
	// Get the TLS config.
	tlsConfig, err := rest.TLSConfigFor(restConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed getting TLS config: %w", err)
	}
	if tlsConfig == nil && restConfig.Transport != nil {
		// If using a custom transport, skip server verification on the upgrade.
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	var upgrader *spdyStream.SpdyRoundTripper
	if restConfig.Proxy != nil {
		upgrader = spdyStream.NewRoundTripperWithProxy(tlsConfig, restConfig.Proxy)
	} else {
		upgrader = spdyStream.NewRoundTripper(tlsConfig)
	}
	wrapper, err := rest.HTTPWrappersForConfig(restConfig, upgrader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating SPDY upgrade wrapper: %w", err)
	}
	return wrapper, upgrader, nil
}
