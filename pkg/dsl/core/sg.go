package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/go-collections/pkg/maps"
)

type SecurityGroupSpec struct {
	Tags  []Tag  `json:"tags" yaml:"tags"`
	Rules []Rule `json:"rules" yaml:"rules"`
	VpcId string `json:"vpcId" yaml:"vpcId"`
}

type Rule struct {
	Protocol    string `json:"protocol" yaml:"protocol"`
	Port        int    `json:"port" yaml:"port"`
	Access      string `json:"access" yaml:"access"`
	Description string `json:"description" yaml:"description"`
}

func SecurityGroupHandler(ctx *pulumi.Context, pipelineItem PipelineItem) error {
	buff := bytes.Buffer{}
	m := maps.MapByStringKey(pipelineItem.Spec)
	encodeErr := json.NewEncoder(&buff).Encode(m)
	if encodeErr != nil {
		panic(encodeErr)
	}
	sgSpec := SecurityGroupSpec{}
	json.NewDecoder(&buff).Decode(&sgSpec)
	fmt.Println("securitygroup spec", sgSpec)
	rules := sgSpec.Rules

	tags := make(pulumi.StringMap, 0)
	for _, tag := range sgSpec.Tags {
		tags[tag.Name] = pulumi.String(tag.Value)
	}

	ingress := make(ec2.SecurityGroupIngressArray, 0)
	for _, rule := range rules {
		if rule.Access == "allow" {
			ingress = append(ingress, ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String(rule.Protocol),
				ToPort:   pulumi.Int(rule.Port),
				FromPort: pulumi.Int(rule.Port),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
				Description: pulumi.String(rule.Description),
			})
		}
	}

	securityGroupArgs := ec2.SecurityGroupArgs{
		Name:    pulumi.String(pipelineItem.Name),
		VpcId:   pulumi.String(sgSpec.VpcId),
		Tags:    tags,
		Ingress: ingress,
	}
	fmt.Println(securityGroupArgs)
	_, err := ec2.NewSecurityGroup(ctx, pipelineItem.Name, &securityGroupArgs)
	if err != nil {
		return err
	}
	return nil
}
