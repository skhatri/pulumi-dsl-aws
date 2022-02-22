### Get started

Initialise stack. Import DSL package into your main func, create requirements.yaml.

Run from example folder or create the following:

### create a stack.go

```
import (
_ "github.com/skhatri/pulumi-dsl-aws/pkg/dsl"
)

func main() {
  //no-op
}
```

While the intention of this project is to create custom DSL for a specific use case, you can also use function call directly.
An example function ```dsl.ManualRun(2)``` will create a 2 node EKS cluster.  

#### initialise stack

```
brew install pulumi
pulumi stack init aws-example
pulumi stack select $(whoami)/example/aws-example
pulumi config set aws:region ap-southeast-2
```

### pulumi up
```
pulumi up 
# or
pulumi up --yes
```

### clean it up
```
pulumi destroy -s $(whoami)/example/aws-example --yes
pulumi stack rm -s $(whoami)/example/aws-example --yes
pulumi stack rm aws-example
```

