package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	k8s "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
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
	PublicIp     *bool             `json:"publicIp" yaml:"publicIp"`
}

type EksCluster struct {
	Name          string            `json:"name" yaml:"name"`
	Version       string            `json:"version" yaml:"version"`
	NodePublicKey string            `json:"nodePublicKey" yaml:"nodePublicKey"`
	VpcId         string            `json:"vpcId" yaml:"vpcId"`
	SubnetId      string            `json:"subnetId" yaml:"subnetId"`
	Tags          map[string]string `json:"tags" yaml:"tags"`
	NodeGroups    []NodeGroupParams `json:"nodeGroups" yaml:"nodeGroups"`
	Size          *int              `json:"size" yaml:"size"`
	InstanceType  *string           `json:"instanceType" yaml:"instanceType"`
}

func CreateEksCluster(ctx *pulumi.Context, clusterParams EksCluster) error {
	clusterArgs := eks.ClusterArgs{
		Name: pulumi.String(clusterParams.Name),
	}
	if clusterParams.Version != "" {
		clusterArgs.Version = pulumi.String(clusterParams.Version)
	}
	fmt.Println("total node groups", len(clusterParams.NodeGroups))
	if clusterParams.Size != nil && *clusterParams.Size > 0 {
		size := *clusterParams.Size
		clusterArgs.DesiredCapacity = pulumi.Int(size)
		clusterArgs.MinSize = pulumi.Int(size)
		clusterArgs.MaxSize = pulumi.Int(size)
		instanceType := "t3.medium"
		if clusterParams.InstanceType != nil {
			instanceType = *clusterParams.InstanceType
		}
		clusterArgs.InstanceType = pulumi.String(instanceType)
	}

	if clusterParams.VpcId != "" {
		clusterArgs.VpcId = pulumi.String(clusterParams.VpcId)
	}
	if clusterParams.SubnetId != "" {
		clusterArgs.SubnetIds = pulumi.ToStringArray(strings.Split(clusterParams.SubnetId, ","))
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

	eksProvider, err := k8s.NewProvider(ctx, "eksProvider", &k8s.ProviderArgs{
		Kubeconfig: cluster.Kubeconfig.ApplyT(
			func(config interface{}) (string, error) {
				b, err := json.Marshal(config)
				if err != nil {
					return "", err
				}
				return string(b), nil
			}).(pulumi.StringOutput),
	})
	if err != nil {
		return err
	}
	eksProviders := pulumi.ProviderMap(map[string]pulumi.ProviderResource{
		"kubernetes": eksProvider,
	})
	for _, nodeGroupParam := range clusterParams.NodeGroups {
		nodeGroupArgs := eks.NodeGroupArgs{
			InstanceType:       pulumi.String(nodeGroupParam.InstanceType),
			DesiredCapacity:    pulumi.Int(nodeGroupParam.Size),
			MinSize:            pulumi.Int(nodeGroupParam.Size),
			MaxSize:            pulumi.Int(nodeGroupParam.Size),
			Cluster:            cluster.Core,
			Version:            pulumi.String(clusterParams.Version),
			NodeRootVolumeSize: pulumi.Int(100),
			NodeSubnetIds:      pulumi.ToStringArray(strings.Split(clusterParams.SubnetId, ",")),
			Labels:             pulumi.StringMap{"topology.kubernetes.io/node-group-name": pulumi.String(nodeGroupParam.Name)},
		}

		if clusterParams.NodePublicKey != "" && nodeGroupParam.PublicKey {
			nodeGroupArgs.NodePublicKey = pulumi.String(clusterParams.NodePublicKey)
		}
		taintMap := make(map[string]eks.TaintInput, 0)

		for k, v := range nodeGroupParam.Taint {
			parts := strings.Split(v, ":")
			effect := "PreferNoSchedule"
			if len(parts) == 2 {
				effect = parts[1]
			}
			taintMap[k] = eks.TaintArgs{
				Value:  pulumi.String(k),
				Effect: pulumi.String(effect),
			}
		}
		nodeGroupArgs.Taints = eks.TaintMap(taintMap)
		publicIp := nodeGroupParam.PublicIp != nil && *nodeGroupParam.PublicIp
		nodeGroupArgs.NodeAssociatePublicIpAddress = pulumi.Bool(publicIp)
		_, nodeGroupErr := eks.NewNodeGroup(ctx, nodeGroupParam.Name, &nodeGroupArgs, eksProviders)
		if nodeGroupErr != nil {
			return nodeGroupErr
		}
	}

	ctx.Export("kubeconfig", cluster.Kubeconfig)

	return err
}
