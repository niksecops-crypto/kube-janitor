package janitor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Stats summarises a single cleanup cycle.
type Stats struct {
	PodsDeleted int
	JobsDeleted int
	Errors      int
}

// Janitor deletes stale Kubernetes pods and jobs.
type Janitor struct {
	client     kubernetes.Interface
	dryRun     bool
	maxAge     time.Duration
	namespaces []string
}

// NewJanitor creates a Janitor. Pass namespace names to restrict scope;
// omit them to operate across all namespaces.
func NewJanitor(client kubernetes.Interface, maxAge time.Duration, dryRun bool, namespaces ...string) *Janitor {
	return &Janitor{
		client:     client,
		maxAge:     maxAge,
		dryRun:     dryRun,
		namespaces: namespaces,
	}
}

// Cleanup runs one full cleanup cycle and returns aggregate statistics.
// Errors from individual namespace operations are counted in Stats.Errors
// but do not abort the cycle.
func (j *Janitor) Cleanup(ctx context.Context) (Stats, error) {
	slog.Info("cleanup cycle started", "dryRun", j.dryRun, "maxAge", j.maxAge)

	var stats Stats

	if err := j.cleanupPods(ctx, &stats); err != nil {
		return stats, fmt.Errorf("pod cleanup: %w", err)
	}
	if err := j.cleanupJobs(ctx, &stats); err != nil {
		return stats, fmt.Errorf("job cleanup: %w", err)
	}

	slog.Info("cleanup cycle finished",
		"podsDeleted", stats.PodsDeleted,
		"jobsDeleted", stats.JobsDeleted,
		"errors", stats.Errors,
	)
	return stats, nil
}

func (j *Janitor) listNamespaces(ctx context.Context) ([]string, error) {
	if len(j.namespaces) > 0 {
		return j.namespaces, nil
	}
	nsList, err := j.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		names = append(names, ns.Name)
	}
	return names, nil
}

func (j *Janitor) cleanupPods(ctx context.Context, stats *Stats) error {
	namespaces, err := j.listNamespaces(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, ns := range namespaces {
		pods, err := j.client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			slog.Error("failed to list pods", "namespace", ns, "error", err)
			stats.Errors++
			continue
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
				continue
			}
			if pod.Status.StartTime == nil {
				continue
			}
			if now.Sub(pod.Status.StartTime.Time) <= j.maxAge {
				continue
			}

			slog.Info("deleting pod",
				"name", pod.Name,
				"namespace", pod.Namespace,
				"phase", pod.Status.Phase,
				"age", now.Sub(pod.Status.StartTime.Time).Round(time.Second),
				"dryRun", j.dryRun,
			)
			if !j.dryRun {
				if err := j.client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
					slog.Error("failed to delete pod", "name", pod.Name, "error", err)
					stats.Errors++
					continue
				}
			}
			stats.PodsDeleted++
		}
	}
	return nil
}

func (j *Janitor) cleanupJobs(ctx context.Context, stats *Stats) error {
	namespaces, err := j.listNamespaces(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, ns := range namespaces {
		jobs, err := j.client.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			slog.Error("failed to list jobs", "namespace", ns, "error", err)
			stats.Errors++
			continue
		}

		for _, job := range jobs.Items {
			var finishTime time.Time
			switch {
			case job.Status.CompletionTime != nil:
				finishTime = job.Status.CompletionTime.Time
			case job.Status.Failed > 0:
				for _, cond := range job.Status.Conditions {
					if string(cond.Type) == "Failed" {
						finishTime = cond.LastTransitionTime.Time
					}
				}
			default:
				continue
			}

			if finishTime.IsZero() || now.Sub(finishTime) <= j.maxAge {
				continue
			}

			slog.Info("deleting job",
				"name", job.Name,
				"namespace", job.Namespace,
				"age", now.Sub(finishTime).Round(time.Second),
				"dryRun", j.dryRun,
			)
			if !j.dryRun {
				prop := metav1.DeletePropagationBackground
				if err := j.client.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{
					PropagationPolicy: &prop,
				}); err != nil {
					slog.Error("failed to delete job", "name", job.Name, "error", err)
					stats.Errors++
					continue
				}
			}
			stats.JobsDeleted++
		}
	}
	return nil
}
