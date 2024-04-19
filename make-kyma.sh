cat <<EOF | kubectl apply -f -
apiVersion: operator.kyma-project.io/v1beta2
kind: Kyma
metadata:
   annotations:
     skr-domain: "example.domain.com"
   name: kyma-sample
   namespace: kcp-system
spec:
   channel: regular
EOF
