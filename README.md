# AppRestart Kubernetes Controller with Kubebuilder

A Kubernetes controller written in Go that watches ConfigMap resources and triggers a rolling restart of Deployment resources by patching their pod template — like an automated `kubectl rollout restart`.

---

## 🚀 Features

- ✅ Watches `v1.ConfigMap` resources cluster-wide
- ✅ Restarts `apps/v1.Deployment` resources cluster-wide
- ✅ Patches with `restartedAt` annotation
- ✅ Retries on Kubernetes object version conflicts
- ✅ Built with `controller-runtime` and deployed via Helm
- ✅ CI tested with GitHub Actions and KinD
- ✅ Prometheus metric: `apprestart_restarts_total`

---

## 🔐 RBAC

This controller runs as a ServiceAccount and requires:

- ClusterRole to list/get/watch ConfigMaps and list/update Deployments
- ClusterRoleBinding to bind the permissions
- ServiceAccount for the pods in the deployment

These solve 403 Forbidden errors like: "User system:serviceaccount:default:app-restart-controller cannot list deployments..."

---

## 📄 CI/CD (GitHub Actions)

- 🛠 KinD spins up locally
- 🐳 Docker image built + loaded into KinD
- 🚀 Helm installs the controller and test resources
- ✅ Controller reacts (with log verification)

---

## 📦 Try it out using the provided Makefile

```bash
make deploy
make log
make destroy

