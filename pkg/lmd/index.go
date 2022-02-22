package lmd

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type LambdaParams struct {
	Name string
	Image string
	Description string
	Memory int
	Timeout int
	FunctionArgName string
	Role pulumi.StringInput
	DependsOn pulumi.Resource
}

func CreateLambdaFunction(ctx *pulumi.Context, lambdaParams LambdaParams) (*lambda.Function, error) {
	return lambda.NewFunction(ctx, lambdaParams.Name, &lambda.FunctionArgs{
		ImageUri:    pulumi.String(lambdaParams.Image),
		Description: pulumi.String(lambdaParams.Description),
		MemorySize:  pulumi.Int(lambdaParams.Memory),
		Timeout:     pulumi.Int(lambdaParams.Timeout),
		Name:        pulumi.String(lambdaParams.FunctionArgName),
		PackageType: pulumi.String("Image"),
		ImageConfig: lambda.FunctionImageConfigArgs{
			Commands: pulumi.ToStringArray([]string{"app"}),
		},
		Environment: lambda.FunctionEnvironmentArgs{
			Variables: pulumi.ToStringMap(map[string]string{
				"NOTIFY_IF_NOT_FOUND": "true",
			}),
		},
		Role: lambdaParams.Role,
	}, pulumi.DependsOn([]pulumi.Resource{lambdaParams.DependsOn}))
}
