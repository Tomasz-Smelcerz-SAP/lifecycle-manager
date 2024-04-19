#!/usr/bin/bash

echo "Deleting cluster kcp-local"
sleep 3

k3d cluster delete kcp-local

echo "Creating cluster kcp-local"
sleep 3

k3d cluster create kcp-local --port 9443:443@loadbalancer --registry-create k3d-registry.localhost:0.0.0.0:5111 --k3s-arg '--disable=traefik@server:0'

sleep 3
k3d kubeconfig get kcp-local > kcp.kubeconfig.yaml

echo "Installing 3rd party components"
sleep 3

$HOME/Downloads/istio-1.20.2/bin/istioctl install --set profile=demo -y

sleep 3
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml

#kubectl create namespace kcp-system
