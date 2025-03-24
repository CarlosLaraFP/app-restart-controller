package controllers

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcile_AddsRestartAnnotation(t *testing.T) {
	_ = appsv1.AddToScheme(scheme.Scheme)

	ctx := context.TODO()

	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "restart-me",
			Namespace: "default",
			Labels: map[string]string{
				"restart": "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(deploy).Build()
	r := &AppRestartReconciler{Client: client}

	_, err := r.Reconcile(context.TODO(), ctrl.Request{
		NamespacedName: kclient.ObjectKeyFromObject(deploy),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var updated appsv1.Deployment
	if err := client.Get(ctx, kclient.ObjectKeyFromObject(deploy), &updated); err != nil {
		t.Fatalf("failed to get updated deployment: %v", err)
	}

	val, ok := updated.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"]
	if !ok || val == "" {
		t.Errorf("expected restartedAt annotation, got: %v", val)
	}

	if _, found := updated.Labels["restart"]; found {
		t.Errorf("expected restart label to be removed")
	}
}

// Deployment without restart: "true" label — should not mutate
func TestReconcile_NoRestartLabel(t *testing.T) {
	_ = appsv1.AddToScheme(scheme.Scheme)

	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "no-restart",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(deploy).Build()
	r := &AppRestartReconciler{Client: client}

	_, err := r.Reconcile(context.TODO(), ctrl.Request{
		NamespacedName: kclient.ObjectKeyFromObject(deploy),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var updated appsv1.Deployment
	_ = client.Get(context.TODO(), kclient.ObjectKeyFromObject(deploy), &updated)

	if updated.Spec.Template.Annotations != nil {
		t.Errorf("expected no annotation, got: %v", updated.Spec.Template.Annotations)
	}
}

// Controller should still remove the restart label, but not re-patch unnecessarily
func TestReconcile_AlreadyAnnotated(t *testing.T) {
	_ = appsv1.AddToScheme(scheme.Scheme)

	alreadyAnnotated := map[string]string{
		"kubectl.kubernetes.io/restartedAt": "2025-03-24T17:22:58Z",
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "already-annotated",
			Namespace: "default",
			Labels:    map[string]string{"restart": "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Annotations: alreadyAnnotated,
					Labels:      map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(deploy).Build()
	r := &AppRestartReconciler{Client: client}

	_, err := r.Reconcile(context.TODO(), ctrl.Request{
		NamespacedName: kclient.ObjectKeyFromObject(deploy),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var updated appsv1.Deployment
	_ = client.Get(context.TODO(), kclient.ObjectKeyFromObject(deploy), &updated)

	if updated.Labels["restart"] != "" {
		t.Errorf("expected restart label to be removed")
	}

	if updated.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] == "" {
		t.Errorf("expected restartedAt annotation to still exist")
	}
}

// Simulates conflict by updating the object mid-reconcile
func TestReconcile_ConflictHandled(t *testing.T) {
	_ = appsv1.AddToScheme(scheme.Scheme)

	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "conflict-test",
			Namespace: "default",
			Labels:    map[string]string{"restart": "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(deploy).Build()
	r := &AppRestartReconciler{Client: client}

	// Simulate external update right after controller fetches the object
	go func() {
		time.Sleep(50 * time.Millisecond) // brief delay to cause version mismatch
		var tmp appsv1.Deployment
		_ = client.Get(context.TODO(), kclient.ObjectKeyFromObject(deploy), &tmp)
		tmp.Labels["injected"] = "true"
		_ = client.Update(context.TODO(), &tmp)
	}()

	_, err := r.Reconcile(context.TODO(), ctrl.Request{
		NamespacedName: kclient.ObjectKeyFromObject(deploy),
	})
	if err != nil {
		t.Fatalf("expected no error despite conflict, got: %v", err)
	}

	var updated appsv1.Deployment
	_ = client.Get(context.TODO(), kclient.ObjectKeyFromObject(deploy), &updated)

	if updated.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] == "" {
		t.Errorf("expected restartedAt annotation to exist after retry")
	}

	if updated.Labels["restart"] != "" {
		t.Errorf("expected restart label to be removed after retry")
	}
}
