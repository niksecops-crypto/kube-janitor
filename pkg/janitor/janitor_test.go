package janitor

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCleanup_SkipsRunningPods(t *testing.T) {
	client := fake.NewSimpleClientset()
	now := metav1.Now()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "running-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase:     corev1.PodRunning,
			StartTime: &now,
		},
	}
	if _, err := client.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	j := NewJanitor(client, 1*time.Hour, false, "default")
	if _, err := j.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}

	remaining, err := client.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining.Items) != 1 {
		t.Errorf("expected 1 running pod to remain, got %d", len(remaining.Items))
	}
}

func TestCleanup_DeletesOldSucceededPod(t *testing.T) {
	client := fake.NewSimpleClientset()
	old := metav1.NewTime(time.Now().Add(-48 * time.Hour))

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-succeeded-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase:     corev1.PodSucceeded,
			StartTime: &old,
		},
	}
	if _, err := client.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	j := NewJanitor(client, 24*time.Hour, false, "default")
	if _, err := j.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}

	remaining, err := client.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining.Items) != 0 {
		t.Errorf("expected old succeeded pod to be deleted, got %d remaining", len(remaining.Items))
	}
}

func TestCleanup_KeepsRecentSucceededPod(t *testing.T) {
	client := fake.NewSimpleClientset()
	recent := metav1.NewTime(time.Now().Add(-30 * time.Minute))

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "recent-succeeded-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase:     corev1.PodSucceeded,
			StartTime: &recent,
		},
	}
	if _, err := client.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	j := NewJanitor(client, 24*time.Hour, false, "default")
	if _, err := j.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}

	remaining, err := client.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining.Items) != 1 {
		t.Errorf("expected recent pod to survive, got %d remaining", len(remaining.Items))
	}
}

func TestCleanup_DryRunDoesNotDelete(t *testing.T) {
	client := fake.NewSimpleClientset()
	old := metav1.NewTime(time.Now().Add(-48 * time.Hour))

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-failed-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase:     corev1.PodFailed,
			StartTime: &old,
		},
	}
	if _, err := client.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	j := NewJanitor(client, 24*time.Hour, true, "default") // dry-run
	if _, err := j.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}

	remaining, err := client.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining.Items) != 1 {
		t.Errorf("dry-run: expected pod to remain, got %d", len(remaining.Items))
	}
}

func TestCleanup_DeletesOldCompletedJob(t *testing.T) {
	client := fake.NewSimpleClientset()
	old := metav1.NewTime(time.Now().Add(-48 * time.Hour))

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-job",
			Namespace: "default",
		},
		Status: batchv1.JobStatus{
			CompletionTime: &old,
		},
	}
	if _, err := client.BatchV1().Jobs("default").Create(context.Background(), job, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	j := NewJanitor(client, 24*time.Hour, false, "default")
	if _, err := j.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}

	remaining, err := client.BatchV1().Jobs("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining.Items) != 0 {
		t.Errorf("expected old job to be deleted, got %d remaining", len(remaining.Items))
	}
}
