package events

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/sqs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/pulumi-dsl-aws/pkg/core"
	"github.com/skhatri/pulumi-dsl-aws/pkg/lmd"
	"github.com/skhatri/pulumi-dsl-aws/pkg/roles"
	"github.com/skhatri/pulumi-dsl-aws/pkg/sch"
)

func CloudWatchEventBridge(name string,
	description string, image string) core.PulumiFunc {
	fn := func(ctx *pulumi.Context) error {
		lambdaAssumeRole, err := roles.CreateLambdaRole(ctx, "lambdaExecRole", "executeLambdaRole")
		if err != nil {
			return err
		}

		lambdaLogging, err := roles.CreateRolePolicy(ctx, "lambdaAppRolePolicy", lambdaAssumeRole.Name)

		if err != nil {
			return err
		}

		lambdaFunction, err := lmd.CreateLambdaFunction(ctx, lmd.LambdaParams{
			Name: "api-notifier",
			Image: image,
			Description: "Invoke Model Y Availability Test",
			Memory: 128,
			Timeout: 4,
			FunctionArgName: "api-notifier-function",
			Role: lambdaAssumeRole.Arn,
			DependsOn: lambdaLogging,
		})

		if err != nil {
			return err
		}

		_, err = sqs.NewQueue(ctx, "result-sink-queue", &sqs.QueueArgs{})
		if err != nil {
			return err
		}
		sch.ScheduleLambdaTrigger(ctx, &sch.ScheduleParams{
			Name:        "check-modely-in-aus",
			Description: "Check Model Y in Australia",
			Cron:        `rate(240 minutes)`,
			Tag:         "check-au",
			TargetArn:   lambdaFunction.Arn,
			TargetName:  lambdaFunction.Name,
			Payload: `{
				"model": "modely", 
				"locale": "en_au", 
				"display_name": "Model Y"
			}`,
			Project: "car-purchase",
			Purpose: "notify",
		})

		sch.ScheduleLambdaTrigger(ctx, &sch.ScheduleParams{
			Name:        "check-model3-in-aus",
			Description: "Check Model 3 in Australia",
			Cron:        `rate(1440 minutes)`,
			Tag:         "check-au",
			TargetArn:   lambdaFunction.Arn,
			TargetName:  lambdaFunction.Name,
			Payload: `{
				"model": "model3", 
				"locale": "en_au", 
				"display_name": "Model 3"
			}`,
			Project: "car-purchase",
			Purpose: "notify",
		})

		sch.ScheduleLambdaTrigger(ctx, &sch.ScheduleParams{
			Name:        "check-modely-in-hk",
			Description: "Check Model Y in Hong Kong",
			Cron:        `rate(1440 minutes)`,
			Tag:         "check-hk",
			TargetArn:   lambdaFunction.Arn,
			TargetName:  lambdaFunction.Name,
			Payload: `{
				"model": "modely", 
				"locale": "en_hk", 
				"display_name": "Model Y"
			}`,
			Project: "car-purchase",
			Purpose: "notify",
		})

		return nil
	}
	return fn
}
