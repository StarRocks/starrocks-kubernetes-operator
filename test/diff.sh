#! /bin/bash

# the script is used to compare resources created by different versions of starrocks-operator
# call it like this: ./diff.sh starrockscluster-sample v1.6.1 v1.7.0

# check the number of arguments
if [ $# -ne 3 ]; then
  echo "Usage: $0 <cluster> <release_tag1> <release_tag2>"
  exit 1
fi
cluster=$1
old_tag=$2
new_tag=$3

# compare fe
echo "compare fe related resources"
diff -u ${old_tag}/${cluster}-fe.yaml ${new_tag}/${cluster}-fe.yaml
diff -u ${old_tag}/${cluster}-fe-search.yaml ${new_tag}/${cluster}-fe-search.yaml
diff -u ${old_tag}/${cluster}-fe-service.yaml ${new_tag}/${cluster}-fe-service.yaml

echo "\n\n"
echo "compare be related resources"
diff -u ${old_tag}/${cluster}-be.yaml ${new_tag}/${cluster}-be.yaml
diff -u ${old_tag}/${cluster}-be-search.yaml ${new_tag}/${cluster}-be-search.yaml
diff -u ${old_tag}/${cluster}-be-service.yaml ${new_tag}/${cluster}-be-service.yaml

echo "\n\n"
echo "compare cn related resources"
diff -u ${old_tag}/${cluster}-cn.yaml ${new_tag}/${cluster}-cn.yaml
diff -u ${old_tag}/${cluster}-cn-search.yaml ${new_tag}/${cluster}-cn-search.yaml
diff -u ${old_tag}/${cluster}-cn-service.yaml ${new_tag}/${cluster}-cn-service.yaml