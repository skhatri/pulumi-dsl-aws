package dsl

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/pulumi-dsl-aws/pkg/dsl/core"
)

type CatalogHandler = func(*pulumi.Context, core.PipelineItem) error

var handlers = make(map[string]CatalogHandler)

func init() {
	handlers["ssm"] = core.SsmHandler
	handlers["eks"] = core.EksHandler
	handlers["s3"] = core.S3Handler
	handlers["sg"] = core.SecurityGroupHandler
	handlers["ec2"] = core.Ec2Handler
}

func init() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		return processRequirements(ctx)
	})
}
