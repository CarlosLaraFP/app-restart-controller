package internal

import (
	"time"

	appsv1 "k8s.io/api/apps/v1"
)

func DeploymentReferencesConfigMap(deployment *appsv1.Deployment, configMapName string) bool {
	for _, c := range deployment.Spec.Template.Spec.Containers {
		for _, envFrom := range c.EnvFrom {
			if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == configMapName {
				// Add dummy annotation to trigger rollout
				annotations := deployment.Spec.Template.Annotations
				if annotations == nil {
					annotations = make(map[string]string)
				}
				annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Local().Format(time.RFC3339)
				deployment.Spec.Template.Annotations = annotations

				return true // first match is enough since the entire pod is patched
			}
		}
	}
	return false
}
