package controllers

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AppRestartReconciler struct {
	client.Client
}

func (r *AppRestartReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)
	/*
		Avoid conflict error due to optimistic concurrency control in Kubernetes in case:
		1. The controller fetched a version of the Deployment (i.e. v1)
		2. Before the controller could update it, something else updated the Deployment, making it v2
		3. The update is now stale, and Kubernetes rejects it to prevent overwriting newer changes

		With retry.RetryOnConflict, we are telling Kubernetes:
		If you hit a version conflict, just refetch the latest version and try again with the changes applied to the newer state.
		This is exactly what a well-behaved controller should do — especially in dynamic environments like Kubernetes where objects are in flux during startup.
	*/
	return ctrl.Result{}, retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var deploy appsv1.Deployment

		if err := r.Get(ctx, req.NamespacedName, &deploy); err != nil {
			return client.IgnoreNotFound(err)
		}

		labels := deploy.GetLabels()

		if val, ok := labels["restart"]; ok && val == "true" {
			logger.Info("Restart label detected", "name", deploy.Name)

			// Add dummy annotation to trigger rollout
			annotations := deploy.Spec.Template.Annotations

			if annotations == nil {
				annotations = make(map[string]string)
			}
			annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Local().Format(time.RFC3339)
			deploy.Spec.Template.Annotations = annotations

			// Optional: Remove the restart label to avoid infinite restart loop
			delete(deploy.Labels, "restart")

			if err := r.Update(ctx, &deploy); err != nil {
				logger.Error(err, "Failed to update deployment")
				return err
			}

			logger.Info("Deployment updated successfully", "Deployment", req.NamespacedName)
			RestartedDeployments.Inc()
		}
		return nil
	})
}

func SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}). // the controller has 1 goroutine actively processing Deployment events
		Complete(&AppRestartReconciler{
			Client: mgr.GetClient(),
		})
}
