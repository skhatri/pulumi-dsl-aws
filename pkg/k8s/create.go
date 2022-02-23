package k8s

import (
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strings"
)

type NodeGroupParams struct {
	Name         string            `json:"name" yaml:"name"`
	Size         int               `json:"size" yaml:"size"`
	Taint        map[string]string `json:"taint" yaml:"taint"`
	InstanceType string            `json:"instanceType" yaml:"instanceType"`
	SpotPrice    string            `json:"spotPrice" yaml:"spotPrice"`
	PublicKey    bool              `json:"publicKey" yaml:"publicKey"`
}

type EksCluster struct {
	Name          string            `json:"name" yaml:"name"`
	Version       string            `json:"version" yaml:"version"`
	NodePublicKey string            `json:"nodePublicKey" yaml:"nodePublicKey"`
	VpcId         string            `json:"vpcId" yaml:"vpcId"`
	SubnetId      string            `json:"subnetId" yaml:"subnetId"`
	Tags          map[string]string `json:"tags" yaml:"tags"`
	NodeGroups    []NodeGroupParams `json:"nodeGroups" yaml:"nodeGroups"`
}

func CreateEksCluster(ctx *pulumi.Context, clusterParams EksCluster) error {
	clusterArgs := eks.ClusterArgs{
		Name: pulumi.String(clusterParams.Name),
	}
	if clusterParams.Version != "" {
		clusterArgs.Version = pulumi.String(clusterParams.Version)
	}

	if len(clusterParams.NodeGroups) == 0 {
		clusterArgs.DesiredCapacity = pulumi.Int(1)
		clusterArgs.MinSize = pulumi.Int(1)
		clusterArgs.MaxSize = pulumi.Int(1)
		clusterArgs.InstanceType = pulumi.String("t3.medium")
	} else {
		clusterArgs.SkipDefaultNodeGroup = pulumi.Bool(true)
	}

	if clusterParams.VpcId != "" {
		clusterArgs.VpcId = pulumi.String(clusterParams.VpcId)
	}
	if clusterParams.SubnetId != "" {
		clusterArgs.SubnetIds = pulumi.StringArray{pulumi.String(clusterParams.SubnetId)}
	}

	if clusterParams.NodePublicKey != "" {
		clusterArgs.NodePublicKey = pulumi.String(clusterParams.NodePublicKey)
	}
	if len(clusterParams.Tags) > 0 {
		tags := make(pulumi.StringMap)
		for k, v := range clusterParams.Tags {
			tags[k] = pulumi.String(v)
		}
		clusterArgs.Tags = tags
	}

	if clusterArgs.DesiredCapacity != pulumi.Int(0) {
		clusterArgs.MaxSize = clusterArgs.DesiredCapacity
	}
	clusterArgs.NodeAssociatePublicIpAddress = pulumi.Bool(false)

	cluster, err := eks.NewCluster(ctx, clusterParams.Name, &clusterArgs)

	if err != nil {
		return err
	}

	for _, nodeGroupParam := range clusterParams.NodeGroups {
		nodeGroupArgs := eks.NodeGroupArgs{
			InstanceType:    pulumi.String(nodeGroupParam.InstanceType),
			DesiredCapacity: pulumi.Int(nodeGroupParam.Size),
			MinSize:         pulumi.Int(nodeGroupParam.Size),
			MaxSize:         pulumi.Int(nodeGroupParam.Size),
			Cluster:         cluster.Core,
		}
		if clusterParams.NodePublicKey != "" && nodeGroupParam.PublicKey {
			nodeGroupArgs.NodePublicKey = pulumi.String(clusterParams.NodePublicKey)
		}
		for k, v := range nodeGroupParam.Taint {
			parts := strings.Split(v, ":")
			effect := "PreferNoSchedule"
			if len(parts) == 2 {
				effect = parts[1]
			}
			nodeGroupArgs.Taints = eks.TaintMap{k: eks.TaintArgs{
				Value:  pulumi.String(k),
				Effect: pulumi.String(effect),
			}}
		}
		_, nodeGroupErr := eks.NewNodeGroup(ctx, nodeGroupParam.Name, &nodeGroupArgs)
		if nodeGroupErr != nil {
			return nodeGroupErr
		}
	}

	ctx.Export("kubeconfig", cluster.Kubeconfig)

	return err
}
