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

var (
	_   = appsv1.AddToScheme(scheme.Scheme)
	_   = corev1.AddToScheme(scheme.Scheme)
	ctx = context.TODO()
)

func TestReconcile_RestartsDeploymentWithEnvFromConfigMap(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my-config",
			Namespace: "default",
		},
		Data: map[string]string{"key": "value"},
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "restart-me",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx",
							EnvFrom: []corev1.EnvFromSource{
								{ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "my-config"},
								}},
							},
						},
					},
				},
			},
		},
	}

	client := fake.
		NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(cm, deploy).
		Build()

	r := &AppRestartReconciler{Client: client}

	_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: kclient.ObjectKeyFromObject(cm)})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var updated appsv1.Deployment
	err = client.Get(ctx, kclient.ObjectKeyFromObject(deploy), &updated)
	if err != nil {
		t.Fatalf("failed to get updated deployment: %v", err)
	}

	if _, ok := updated.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"]; !ok {
		t.Errorf("expected restartedAt annotation, but got none")
	}
}

func TestReconcile_IgnoresDeploymentsWithoutEnvFrom(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "unused-config",
			Namespace: "default",
		},
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "no-config",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "test",
							Image:   "nginx",
							EnvFrom: []corev1.EnvFromSource{}, // no ConfigMap ref
						},
					},
				},
			},
		},
	}

	client := fake.
		NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(cm, deploy).
		Build()

	r := &AppRestartReconciler{Client: client}

	_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: kclient.ObjectKeyFromObject(cm)})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var updated appsv1.Deployment
	_ = client.Get(ctx, kclient.ObjectKeyFromObject(deploy), &updated)

	if updated.Spec.Template.Annotations != nil {
		t.Errorf("expected no annotation, got: %v", updated.Spec.Template.Annotations)
	}
}

func TestReconcile_HandlesConflictDuringUpdate(t *testing.T) {
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "conflict-cm",
			Namespace: "default",
		},
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "conflict-deploy",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &v1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx",
							EnvFrom: []corev1.EnvFromSource{
								{ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "conflict-cm"},
								}},
							},
						},
					},
				},
			},
		},
	}

	client := fake.
		NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(cm, deploy).
		Build()

	r := &AppRestartReconciler{Client: client}

	// Simulate external mutation with goroutine closure
	go func() {
		time.Sleep(10 * time.Millisecond)
		var temp appsv1.Deployment
		_ = client.Get(ctx, kclient.ObjectKeyFromObject(deploy), &temp)
		temp.Labels = map[string]string{"injected": "true"}
		_ = client.Update(ctx, &temp)
	}()

	_, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: kclient.ObjectKeyFromObject(cm)})
	if err != nil {
		t.Fatalf("expected no error even with conflict, got: %v", err)
	}

	var updated appsv1.Deployment
	_ = client.Get(ctx, kclient.ObjectKeyFromObject(deploy), &updated)

	if _, ok := updated.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"]; !ok {
		t.Errorf("expected restartedAt annotation after reconcile retry")
	}
}
