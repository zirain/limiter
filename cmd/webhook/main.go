package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"

	"github.com/zirain/limiter/pkg/webhook"
)

func main() {
	var param webhook.WhSvrParam
	// webhook http server（tls）
	// 命令行参数
	flag.IntVar(&param.Port, "port", 443, "Webhook Server Port.")
	flag.StringVar(&param.CertFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "x509 certification file")
	flag.StringVar(&param.KeyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "x509 private key file")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(param.CertFile, param.KeyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %v", err)
		return
	}

	// 实例化一个Webhook Server
	s := webhook.Server{
		Server: &http.Server{
			Addr: fmt.Sprintf(":%d", param.Port),
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	// 定义 http server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", s.Handler)
	s.Server.Handler = mux

	// 在一个新的 goroutine 里面去启动 webhook server
	go func() {
		if err := s.Server.ListenAndServeTLS("", ""); err != nil {
			klog.Errorf("Failed to listen and serve webhook: %v", err)
		}
	}()

	klog.Info("Server started")

	// 监听 OS 的关闭信号
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	klog.Infof("Receive OS shutdown signal, gracefully shutting down...")
	if err := s.Server.Shutdown(context.Background()); err != nil {
		klog.Errorf("HTTP Server Shutdown error: %v", err)
	}

}
