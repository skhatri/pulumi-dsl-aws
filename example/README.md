### test

#### setup
```
pulumi stack init aws-example
pulumi stack select $(whoami)/example/aws-example
pulumi config set aws:region ap-southeast-2
```

#### run
```
pulumi up
```

#### verify
An IP is printed, which you can visit.

```
EC2_IP=$(pulumi stack output ec2-0)
open http://${EC2_IP}/
```
executing above should print "server <internal ip> says hello to you"

#### cleanup
```
pulumi destroy -s $(whoami)/example/aws-example --yes
pulumi stack rm -s $(whoami)/example/aws-example --yes
pulumi stack rm aws-example
```


