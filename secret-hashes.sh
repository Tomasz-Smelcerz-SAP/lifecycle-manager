echo -n "tls.crt: "
kubectl -n $1 get secret $2 -o jsonpath='{.data.tls\.crt}' | md5sum
echo -n "tls.key: "
kubectl -n $1 get secret $2 -o jsonpath='{.data.tls\.key}' | md5sum
echo -n "ca.crt : "
kubectl -n $1 get secret $2 -o jsonpath='{.data.ca\.crt}' | md5sum
