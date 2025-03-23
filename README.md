# AppRestart Kubernetes Controller with Kubebuilder

## What It Does:

* Watches Deployments with a label like restart: "true"
* When it sees one, it patches the deployment with a dummy annotation (kubectl rollout restart).
* Optionally resets the label after restart.
* This mimics a simplified kubectl rollout restart, but in an automated, reactive way.

## Bonus Features:

* Add a pause: "true" label to temporarily disable restarts
* Add a status field via CRD status subresource
* Helm chart to deploy the controller
* Integration tests with envtest