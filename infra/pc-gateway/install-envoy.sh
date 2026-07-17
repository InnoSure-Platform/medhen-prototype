#!/usr/bin/env bash
set -e

echo "==> Installing Kubernetes Gateway API CRDs..."
kubectl get crd gateways.gateway.networking.k8s.io || \
  kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml

echo "==> Installing Envoy Gateway using kubectl..."
kubectl apply --server-side -f https://github.com/envoyproxy/gateway/releases/download/v1.1.0/install.yaml

echo "==> Waiting for Envoy Gateway to be ready..."
kubectl wait --timeout=3m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available

echo "==> Envoy Gateway Installed Successfully."
