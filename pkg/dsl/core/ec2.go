package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/go-collections/pkg/maps"
)

type Ec2Spec struct {
	Ami            string   `json:"ami" yaml:"ami"`
	Tags           []Tag    `json:"tags" yaml:"tags"`
	Nodes          int      `json:"nodes" yaml:"nodes"`
	UserData       string   `json:"userData" yaml:"userData"`
	InstanceType   string   `json:"instanceType" yaml:"instanceType"`
	SecurityGroups []string `json:"securityGroups" yaml:"securityGroups"`
	PublicIp       *bool    `json:"publicIp" yaml:"publicIp"`
	PublicKey      string   `json:"publicKey" yaml:"publicKey"`
}

func Ec2Handler(ctx *pulumi.Context, pipelineItem PipelineItem) error {
	buff := bytes.Buffer{}
	m := maps.MapByStringKey(pipelineItem.Spec)
	encodeErr := json.NewEncoder(&buff).Encode(m)
	if encodeErr != nil {
		panic(encodeErr)
	}
	ec2Spec := Ec2Spec{}
	json.NewDecoder(&buff).Decode(&ec2Spec)
	tags := make(pulumi.StringMap, 0)
	for _, tag := range ec2Spec.Tags {
		tags[tag.Name] = pulumi.String(tag.Value)
	}

	var err error
	securityGroupIds := make([]string, 0)
	if len(ec2Spec.SecurityGroups) != 0 {
		res, err := ec2.GetSecurityGroups(ctx, &ec2.GetSecurityGroupsArgs{
			Filters: []ec2.GetSecurityGroupsFilter{
				{
					Name:   "group-name",
					Values: ec2Spec.SecurityGroups,
				},
			},
		})
		if err != nil {
			return err
		}
		securityGroupIds = res.Ids
	}

	for i := 0; i < ec2Spec.Nodes; i++ {
		name := fmt.Sprintf("%s-%s", pipelineItem.Name, pulumi.String(i))
		args := ec2.InstanceArgs{
			Ami:                 pulumi.String(ec2Spec.Ami),
			InstanceType:        pulumi.String(ec2Spec.InstanceType),
			Tags:                tags,
			VpcSecurityGroupIds: pulumi.ToStringArray(securityGroupIds),
		}
		if ec2Spec.UserData != "" {
			args.UserData = pulumi.String(ec2Spec.UserData)
		}
		publicIp := ec2Spec.PublicIp != nil && *ec2Spec.PublicIp
		args.AssociatePublicIpAddress = pulumi.Bool(publicIp)
		_, err = ec2.NewInstance(ctx, name, &args)
		if err != nil {
			return err
		}
	}
	return err
}
