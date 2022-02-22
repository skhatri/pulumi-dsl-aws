package kms

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/pulumi-dsl-aws/pkg/core"
)

//CreateKms
//Usage CreateKms(365, "slack-lambda-kms-key", "Slack Lambda Kms Key")
func CreateKms(deletion int, name string, description string) core.PulumiFunc {

	var fn = func(ctx *pulumi.Context) error {
		_, err := kms.NewKey(ctx, name, &kms.KeyArgs{
			DeletionWindowInDays: pulumi.Int(deletion),
			Description:          pulumi.String(description),
		})
		if err != nil {
			return err
		}
		return nil
	}
	return fn
}
