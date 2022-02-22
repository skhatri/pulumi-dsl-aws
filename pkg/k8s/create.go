package k8s

import (
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ClusterParams struct {
	Name            string
	Version         string
	DesiredCapacity int
	InstanceType    string
	NodePublicKey   string
	VpcId           string
	SubnetId        string
	Tags            map[string]string
}

func CreateEksCluster(ctx *pulumi.Context, clusterParams ClusterParams) error {
	clusterArgs := eks.ClusterArgs{
		Name: pulumi.String(clusterParams.Name),
	}
	if clusterParams.Version != "" {
		clusterArgs.Version = pulumi.String(clusterParams.Version)
	}
	if clusterParams.DesiredCapacity != 0 {
		clusterArgs.DesiredCapacity = pulumi.Int(clusterParams.DesiredCapacity)
	}
	if clusterParams.InstanceType != "" {
		clusterArgs.InstanceType = pulumi.String(clusterParams.InstanceType)
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

	ctx.Export("kubeconfig", cluster.Kubeconfig)

	return err
}
