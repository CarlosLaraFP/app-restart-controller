package controllers

import (
	"context"
	"testing"

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
