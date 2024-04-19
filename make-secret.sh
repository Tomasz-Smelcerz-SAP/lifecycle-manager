cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
   name: kyma-sample
   namespace: kcp-system
   labels:
     "operator.kyma-project.io/kyma-name": "kyma-sample"
     "operator.kyma-project.io/managed-by": "lifecycle-manager"
data:
   config: $(k3d kubeconfig get skr-local | sed 's/0\.0\.0\.0/host.k3d.internal/' | base64 | tr -d '\n')
EOF
