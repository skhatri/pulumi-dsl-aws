package core

import (
	"bytes"
	"encoding/json"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/go-collections/pkg/maps"
	"strconv"
	"strings"
)

type SecurityGroupSpec struct {
	Tags  []Tag  `json:"tags" yaml:"tags"`
	Rules []Rule `json:"rules" yaml:"rules"`
	VpcId string `json:"vpcId" yaml:"vpcId"`
}

type Rule struct {
	Protocol    string  `json:"protocol" yaml:"protocol"`
	Port        string  `json:"port" yaml:"port"`
	Access      string  `json:"access" yaml:"access"`
	Description string  `json:"description" yaml:"description"`
	Outbound    *bool   `json:"outbound" yaml:"outbound"`
	Cidr        *string `json:"cidr" yaml:"cidr"`
}

const (
	MIN_PORT = 0
	MAX_PORT = 65535
)

func parsePort(portValue string) (int, int) {
	fromPort := MIN_PORT
	toPort := MAX_PORT
	if portValue != "" {
		switch portValue {
		case "all":
			fromPort = MIN_PORT
			toPort = MAX_PORT
		case "*":
			fromPort = MIN_PORT
			toPort = MAX_PORT
		default:
			port := strings.Split(portValue, "-")
			if len(port) == 2 {
				fromPort, _ = strconv.Atoi(port[0])
				toPort, _ = strconv.Atoi(port[1])
			} else {
				fromPort, _ = strconv.Atoi(port[0])
				toPort = fromPort
			}
		}
	}
	return fromPort, toPort
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
	rules := sgSpec.Rules

	tags := make(pulumi.StringMap, 0)
	for _, tag := range sgSpec.Tags {
		tags[tag.Name] = pulumi.String(tag.Value)
	}

	ingress := make(ec2.SecurityGroupIngressArray, 0)
	egress := make(ec2.SecurityGroupEgressArray, 0)
	cidr := "0.0.0.0/0"
	for _, rule := range rules {
		if rule.Cidr != nil && *rule.Cidr != "" {
			cidr = *rule.Cidr
		}
		fromPort, toPort := parsePort(rule.Port)
		isOutbound := rule.Outbound != nil && *rule.Outbound == true

		if rule.Access == "allow" {

			if isOutbound {
				egress = append(egress, ec2.SecurityGroupEgressArgs{
					Protocol: pulumi.String(rule.Protocol),
					FromPort: pulumi.Int(fromPort),
					ToPort:   pulumi.Int(toPort),
					CidrBlocks: pulumi.StringArray{
						pulumi.String(cidr),
					},
				})
			} else {
				ingress = append(ingress, ec2.SecurityGroupIngressArgs{
					Protocol: pulumi.String(rule.Protocol),
					FromPort: pulumi.Int(fromPort),
					ToPort:   pulumi.Int(toPort),
					CidrBlocks: pulumi.StringArray{
						pulumi.String(cidr),
					},
					Description: pulumi.String(rule.Description),
				})
			}
		}
	}

	securityGroupArgs := ec2.SecurityGroupArgs{
		Name:    pulumi.String(pipelineItem.Name),
		VpcId:   pulumi.String(sgSpec.VpcId),
		Tags:    tags,
		Ingress: ingress,
		Egress:  egress,
	}
	_, err := ec2.NewSecurityGroup(ctx, pipelineItem.Name, &securityGroupArgs)
	if err != nil {
		return err
	}
	return nil
}
