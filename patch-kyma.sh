for i in {01..10}; do
  echo $i
  sed "s/{\$i}/$i/g" < patch-kyma.json > patch-kyma.tmp
  kubectl -n kyma-system patch kyma default --type json --patch-file patch-kyma.tmp
  sleep 5
done

