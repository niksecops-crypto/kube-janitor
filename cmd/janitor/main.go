package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/niksecops-crypto/kube-janitor/pkg/janitor"
	"github.com/niksecops-crypto/kube-janitor/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var version = "dev"

type nsFlag []string

func (n *nsFlag) String() string       { return strings.Join(*n, ",") }
func (n *nsFlag) Set(v string) error   { *n = append(*n, v); return nil }

func main() {
	var (
		maxAge      = flag.Duration("max-age", 24*time.Hour, "Maximum age for finished resources before deletion")
		dryRun      = flag.Bool("dry-run", false, "Preview mode — no deletions")
		kubeconfig  = flag.String("kubeconfig", "", "Path to kubeconfig (defaults to in-cluster config)")
		interval    = flag.Duration("interval", 1*time.Hour, "Cleanup interval")
		metricsAddr = flag.String("metrics-addr", ":9090", "Address to expose Prometheus metrics on /metrics")
		showVer     = flag.Bool("version", false, "Print version and exit")
		namespaces  nsFlag
	)
	flag.Var(&namespaces, "namespace", "Namespace to watch (repeatable; default: all)")
	flag.Parse()

	if *showVer {
		fmt.Println(version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	var (
		cfg *rest.Config
		err error
	)
	if *kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		slog.Error("failed to build kubeconfig", "error", err)
		os.Exit(1)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		slog.Error("failed to create clientset", "error", err)
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()
	m := metrics.New(reg)

	j := janitor.NewJanitor(client, *maxAge, *dryRun, namespaces...)

	// Start metrics HTTP server.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			slog.Warn("healthz write error", "error", err)
		}
	})
	metricsSrv := &http.Server{
		Addr:         *metricsAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		slog.Info("metrics server started", "addr", *metricsAddr)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("metrics server error", "error", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("kube-janitor started",
		"version", version,
		"maxAge", *maxAge,
		"dryRun", *dryRun,
		"interval", *interval,
		"namespaces", namespaces.String(),
		"metricsAddr", *metricsAddr,
	)

	runCleanup := func() {
		start := time.Now()
		stats, err := j.Cleanup(ctx)
		elapsed := time.Since(start).Seconds()

		m.CleanupSeconds.Observe(elapsed)
		m.PodsDeleted.Add(float64(stats.PodsDeleted))
		m.JobsDeleted.Add(float64(stats.JobsDeleted))
		m.Errors.Add(float64(stats.Errors))

		if err != nil {
			slog.Error("cleanup cycle failed", "error", err)
		}
	}

	runCleanup()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runCleanup()
		case <-ctx.Done():
			slog.Info("kube-janitor shutting down")
			shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutCancel()
			if err := metricsSrv.Shutdown(shutCtx); err != nil {
				slog.Warn("metrics server shutdown error", "error", err)
			}
			return
		}
	}
}
