stack=$1/$2

echo 'name: '$1'
runtime: go
description: Infrastructure Example' > Pulumi.yaml

echo 'config:
  aws:region: ap-southeast-2
' > Pulumi.$2.yaml

pulumi stack select $(whoami)/$stack


