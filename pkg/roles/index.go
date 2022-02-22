package roles

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateRolePolicy(ctx *pulumi.Context, name string, argName pulumi.StringOutput) (*iam.RolePolicy, error) {
	return iam.NewRolePolicy(ctx, name, &iam.RolePolicyArgs{
		Role: argName,
		Policy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Action": [
						"logs:CreateLogGroup",
						"logs:PutLogEvents",
						"logs:CreateLogStream"
					],
					"Resource": "arn:aws:logs:*:*:*"
				},
				{
					"Effect": "Allow",
					"Action": [
					  "ssm:Describe*",
					  "ssm:Get*",
					  "ssm:List*"
					],
					"Resource": "*"
				},
				{
					"Effect": "Allow",
					"Action": [
					  "lambda:InvokeFunction"
					],
					"Resource": "*"
				},
				{
					"Effect": "Allow",
					"Action": [
						"SQS:*"
					  ],
					  "Resource": "arn:aws:sqs:*:*:*"
				}]
			}`),
	})
}

func CreateLambdaRole(ctx *pulumi.Context, roleName string, argName string) (*iam.Role, error) {
	return iam.NewRole(ctx, roleName, &iam.RoleArgs{
		Name: pulumi.String(argName),
		AssumeRolePolicy: pulumi.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Sid": "",
						"Effect": "Allow",
						"Principal": {
							"Service": "lambda.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
					},
					{
						"Sid": "",
						"Effect": "Allow",
						"Principal": {
							"Service": "events.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
					},
					{
						"Sid": "",
						"Effect": "Allow",
						"Principal": {
							"Service": "sqs.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
					}
					]
				}`),
	})
}
