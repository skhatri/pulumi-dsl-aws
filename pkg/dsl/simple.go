package dsl

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/pulumi-dsl-aws/pkg/k8s"
	ssm2 "github.com/skhatri/pulumi-dsl-aws/pkg/ssm"
)

func ManualRun(nodes int) {
	pulumi.Run(func(ctx *pulumi.Context) error {

		var err error = nil
		err = ssm2.PutSsmParameter("eks-start-time", "19", "String")(ctx)
		if err != nil {
			return err
		}

		err = ssm2.PutSsmParameter("eks-end-time", "22", "String")(ctx)

		if err != nil {
			return err
		}

		err = k8s.CreateEksCluster(ctx, k8s.ClusterParams{
			Name:            "example-cluster",
			Version:         "1.20",
			DesiredCapacity: nodes,
			InstanceType:    "t3.medium",
			NodePublicKey:   "",
		})

		return err
	})
}

