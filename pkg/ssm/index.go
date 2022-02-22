package ssm

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/pulumi-dsl-aws/pkg/core"
)

func PutSsmParameter(paramName string, paramValue string, paramType string) core.PulumiFunc {
	var fn = func(ctx *pulumi.Context) error {
		param, err := ssm.NewParameter(ctx, paramName, &ssm.ParameterArgs{
			Type:  pulumi.String(paramType),
			Value: pulumi.String(paramValue),
			Name:  pulumi.String(fmt.Sprintf("%s-parameter", paramName)),
		})
		if err != nil {
			return err
		}
		ctx.Export(paramName, param.ID())
		return nil
	}
	return fn
}
