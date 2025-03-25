// main.go
package main

import (
	"flag"
	"os"

	"app-restart-controller/controllers"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Sets up:
	// a shared cache (for event-driven watches)
	// a client (for CRUD)
	// scheme (object types)
	// metrics and health endpoints
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr, // starts an HTTP server exposing Prometheus metrics at /metrics on port 8080
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "app-restart-controller.example.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	if err = controllers.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	// The manager initializes a shared informer cache that watches the API server for changes to the resources being watched (e.g. via a watch stream).
	// These events (add/update/delete) are pushed into a rate-limited queue per controller.
	// The controller’s Reconcile() is then called with the object key (namespace/name), and uses the Client (which reads from the cache) to get the actual resource state.
	// Events = via watch → enqueue reconcile.Request
	// Object data = read from cache (backed by API server)
	// The shared informer cache is a local, in-memory store of Kubernetes resources.
	// It’s populated and kept up-to-date by watching the API server using List + Watch. It exists to:
	// Avoid hammering the API server for every single read
	// Allow fast, consistent access to objects during Reconcile
	// Provide event notifications (add/update/delete) for controllers
	// In controller-runtime, it’s abstracted as cache.Cache, but under the hood it uses client-go's SharedInformers

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// go mod init app-restart-controller
// go mod tidy
