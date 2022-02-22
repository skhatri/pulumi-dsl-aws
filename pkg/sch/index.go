package sch

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ScheduleParams struct {
	Name        string
	Description string
	Cron        string
	Tag         string
	TargetArn   pulumi.StringOutput
	TargetName  pulumi.StringOutput
	Payload     string
	Project     string
	Purpose     string
}

func ScheduleLambdaTrigger(ctx *pulumi.Context, scheduleParams *ScheduleParams) error {

	cronRule, err := cloudwatch.NewEventRule(ctx, fmt.Sprintf("%s-event-rule", scheduleParams.Name), &cloudwatch.EventRuleArgs{
		Description:        pulumi.String(scheduleParams.Description),
		ScheduleExpression: pulumi.String(scheduleParams.Cron),
		Tags: pulumi.ToStringMap(map[string]string{
			"purpose": scheduleParams.Purpose,
			"project": scheduleParams.Project,
			"task":    scheduleParams.Tag,
		}),
	})

	if err != nil {
		return err
	}

	_, err = cloudwatch.NewEventTarget(ctx, fmt.Sprintf("%s-target", scheduleParams.Name), &cloudwatch.EventTargetArgs{
		Input: pulumi.String(scheduleParams.Payload),
		Rule:  cronRule.Name,
		Arn:   scheduleParams.TargetArn,
	})
	if err != nil {
		return err
	}

	_, err = lambda.NewPermission(ctx, fmt.Sprintf("%s-lambda-permission", scheduleParams.Name), &lambda.PermissionArgs{
		Action:    pulumi.String("lambda:InvokeFunction"),
		Function:  scheduleParams.TargetName,
		SourceArn: cronRule.Arn,
		Principal: pulumi.String("events.amazonaws.com"),
	})
	if err != nil {
		return err
	}
	return nil
}

