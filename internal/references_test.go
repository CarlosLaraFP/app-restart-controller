package internal

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentReferencesConfigMap_Positive(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{Name: "test"},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test",
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

	if !DeploymentReferencesConfigMap(deploy, "my-config") {
		t.Errorf("Expected true but got false")
	}
}

func TestDeploymentReferencesConfigMap_Negative(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{Name: "test"},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "test",
							EnvFrom: []corev1.EnvFromSource{}, // empty
						},
					},
				},
			},
		},
	}

	if DeploymentReferencesConfigMap(deploy, "not-used") {
		t.Errorf("Expected false but got true")
	}
}
