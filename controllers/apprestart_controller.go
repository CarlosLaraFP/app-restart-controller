package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
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

	var deploy appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deploy); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	labels := deploy.GetLabels()
	if val, ok := labels["restart"]; ok && val == "true" {
		logger.Info("Restart label detected", "name", deploy.Name)

		// Add dummy annotation to trigger rollout
		annotations := deploy.Spec.Template.Annotations
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["kubectl.kubernetes.io/restartedAt"] = fmt.Sprintf("%v", ctx.Value("now"))
		deploy.Spec.Template.Annotations = annotations

		if err := r.Update(ctx, &deploy); err != nil {
			logger.Error(err, "Failed to update deployment")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(&AppRestartReconciler{
			Client: mgr.GetClient(),
		})
}
