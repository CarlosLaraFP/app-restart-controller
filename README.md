# AppRestart Kubernetes Controller with Kubebuilder

A Kubernetes controller written in Go that watches Deployments labeled `restart: "true"` and triggers a rolling restart by patching their pod template — like an automated `kubectl rollout restart`.

---

## 🚀 Features

- ✅ Watches `apps/v1.Deployment` resources cluster-wide
- ✅ Detects `restart: "true"` label
- ✅ Patches with `restartedAt` annotation
- ✅ Removes label to avoid infinite loops
- ✅ Retries on Kubernetes object version conflicts
- ✅ Built with `controller-runtime`, deployed via Helm
- ✅ CI tested with GitHub Actions and KinD
- ✅ Prometheus metric: `apprestart_restarts_total`

---

## 🔐 RBAC

This controller runs as a ServiceAccount and requires:

- ClusterRole to list/watch/update Deployments
- ClusterRoleBinding to bind the permissions
- ServiceAccount (automatically created by the chart)

These solve 403 Forbidden errors like: "User system:serviceaccount:default:app-restart-controller cannot list deployments..."

---

## 📄 CI/CD (GitHub Actions)

- 🛠 KinD spins up locally
- 🐳 Docker image built + loaded into KinD
- 🚀 Helm installs the controller
- 🔁 A test Deployment is applied
- ✅ Controller reacts (with log verification)

---

## 📦 Try it out using the provided Makefile

```bash
make deploy
make apply-test-deployment
make destroy

