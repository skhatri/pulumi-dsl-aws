package k8s

import (
	"encoding/json"
	"fmt"
	eks2 "github.com/pulumi/pulumi-aws/sdk/v4/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
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
	Managed      *bool             `json:"managed" yaml:"managed"`
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
	err, role1, instanceProfile1 := createRoleAndInstanceProfile(ctx)
	if err != nil {
		return err
	}

	var clusterArgs *eks.ClusterArgs

	fmt.Println("total node groups", len(clusterParams.NodeGroups))
	clusterArgs = &eks.ClusterArgs{
		Name: pulumi.String(clusterParams.Name),
	}
	if clusterParams.Version != "" {
		clusterArgs.Version = pulumi.String(clusterParams.Version)
	}
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
		clusterArgs.InstanceRoles = iam.RoleArray{role1}
		clusterArgs.InstanceProfileName = instanceProfile1.Name
	} else {
		clusterArgs.SkipDefaultNodeGroup = pulumi.Bool(true)
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

	fmt.Println("cluster args", clusterArgs)
	cluster, err := eks.NewCluster(ctx, clusterParams.Name, clusterArgs)

	if err != nil {
		return err
	}

	eksProvider, err := k8s.NewProvider(ctx, "eksProvider", &k8s.ProviderArgs{
		Cluster: clusterArgs.Name,
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
		"provider": eksProvider,
	})

	err = createNodeGroups(ctx, clusterParams, cluster, instanceProfile1, eksProviders)
	if err != nil {
		return err
	}

	err = createManagedNodeGroups(ctx, clusterParams, cluster, role1, eksProviders)
	if err != nil {
		return err
	}

	ctx.Export("kubeconfig", cluster.Kubeconfig)
	return err
}

func createRoleAndInstanceProfile(ctx *pulumi.Context) (error, *iam.Role, *iam.InstanceProfile) {
	tmpJSON0, err := json.Marshal(map[string]interface{}{
		"Statement": []map[string]interface{}{
			map[string]interface{}{
				"Action": "sts:AssumeRole",
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"Service": "ec2.amazonaws.com",
				},
			},
		},
		"Version": "2012-10-17",
	})
	if err != nil {
		return err, nil, nil
	}
	json0 := string(tmpJSON0)
	role1, err := iam.NewRole(ctx, "example", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(json0),
	})
	if err != nil {
		return err, nil, nil
	}
	_, err = iam.NewRolePolicyAttachment(ctx, "example-AmazonEKSWorkerNodePolicy", &iam.RolePolicyAttachmentArgs{
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"),
		Role:      role1.Name,
	})
	if err != nil {
		return err, nil, nil
	}
	_, err = iam.NewRolePolicyAttachment(ctx, "example-AmazonEKSCNIPolicy", &iam.RolePolicyAttachmentArgs{
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"),
		Role:      role1.Name,
	})
	if err != nil {
		return err, nil, nil
	}
	_, err = iam.NewRolePolicyAttachment(ctx, "example-AmazonEC2ContainerRegistryReadOnly", &iam.RolePolicyAttachmentArgs{
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"),
		Role:      role1.Name,
	})
	if err != nil {
		return err, nil, nil
	}
	var instanceProfile1 *iam.InstanceProfile
	instanceProfile1, err = iam.NewInstanceProfile(ctx, "testProfile", &iam.InstanceProfileArgs{
		Role: role1.Name,
	})
	return err, role1, instanceProfile1
}

func createNodeGroups(ctx *pulumi.Context, clusterParams EksCluster, cluster *eks.Cluster, instanceProfile1 *iam.InstanceProfile, eksProviders pulumi.ResourceOption) error {
	for _, nodeGroupParam := range clusterParams.NodeGroups {
		if nodeGroupParam.Managed != nil && *nodeGroupParam.Managed {
			continue
		}
		nodeGroupsArgs := eks.NodeGroupArgs{
			Cluster:         cluster.Core,
			Labels:          pulumi.StringMap{"topology.kubernetes.io/node-group-name": pulumi.String(nodeGroupParam.Name)},
			InstanceType:    pulumi.String(nodeGroupParam.InstanceType),
			DesiredCapacity: pulumi.Int(nodeGroupParam.Size),
			MinSize:         pulumi.Int(0),
			MaxSize:         pulumi.Int(nodeGroupParam.Size),

			InstanceProfile:    instanceProfile1,
			NodeSubnetIds:      pulumi.ToStringArray(strings.Split(clusterParams.SubnetId, ",")),
			NodeRootVolumeSize: pulumi.Int(100),
			NodePublicKey:      pulumi.String(clusterParams.NodePublicKey),
			Version:            pulumi.String(clusterParams.Version),
		}

		if clusterParams.NodePublicKey != "" && nodeGroupParam.PublicKey {
			nodeGroupsArgs.NodePublicKey = pulumi.String(clusterParams.NodePublicKey)
		}

		taintMap := make(map[string]eks.TaintInput, 0)
		taintMap["node.kubernetes.io/not-ready"] = eks.TaintArgs{
			Effect: pulumi.String("NO_SCHEDULE"),
		}
		taintMap["node.kubernetes.io/unreachable"] = eks.TaintArgs{
			Effect: pulumi.String("NO_SCHEDULE"),
		}
		for k, v := range nodeGroupParam.Taint {
			parts := strings.Split(v, ":")
			effect := "PREFER_NO_SCHEDULE"
			if len(parts) == 2 {
				effect = parts[1]
			}
			taintMap[k] = eks.TaintArgs{
				Value:  pulumi.String(k),
				Effect: pulumi.String(effect),
			}
		}
		nodeGroupsArgs.Taints = eks.TaintMap(taintMap)

		_, nodeGroupErr := eks.NewNodeGroup(ctx, nodeGroupParam.Name, &nodeGroupsArgs, eksProviders)
		if nodeGroupErr != nil {
			return nodeGroupErr
		}
	}
	return nil
}

func createManagedNodeGroups(ctx *pulumi.Context, clusterParams EksCluster, cluster *eks.Cluster, role1 *iam.Role, eksProviders pulumi.ResourceOption) error {
	for _, nodeGroupParam := range clusterParams.NodeGroups {
		if nodeGroupParam.Managed == nil || !*nodeGroupParam.Managed {
			continue
		}
		managedNodeGroupArgs := eks.ManagedNodeGroupArgs{
			Cluster:       cluster.Core,
			Version:       pulumi.String(clusterParams.Version),
			Labels:        pulumi.StringMap{"topology.kubernetes.io/node-group-name": pulumi.String(nodeGroupParam.Name)},
			InstanceTypes: pulumi.StringArray{pulumi.String(nodeGroupParam.InstanceType)},
			ScalingConfig: eks2.NodeGroupScalingConfigArgs{
				DesiredSize: pulumi.Int(nodeGroupParam.Size),
				MinSize:     pulumi.Int(0),
				MaxSize:     pulumi.Int(nodeGroupParam.Size),
			},
			NodeGroupName: pulumi.String(nodeGroupParam.Name),
			DiskSize:      pulumi.Int(100),
			SubnetIds:     pulumi.ToStringArray(strings.Split(clusterParams.SubnetId, ",")),
			NodeRole:      role1,
		}

		if clusterParams.NodePublicKey != "" && nodeGroupParam.PublicKey {
			managedNodeGroupArgs.RemoteAccess = eks2.NodeGroupRemoteAccessArgs{
				Ec2SshKey: pulumi.String(clusterParams.NodePublicKey),
			}
		}
		taintArray := eks2.NodeGroupTaintArray{
			eks2.NodeGroupTaintArgs{
				Key:    pulumi.String("node.kubernetes.io/not-ready"),
				Effect: pulumi.String("NO_SCHEDULE"),
			},
			eks2.NodeGroupTaintArgs{
				Key:    pulumi.String("node.kubernetes.io/unreachable"),
				Effect: pulumi.String("NO_SCHEDULE"),
			},
		}

		for k, v := range nodeGroupParam.Taint {
			parts := strings.Split(v, ":")
			effect := "PREFER_NO_SCHEDULE"
			if len(parts) == 2 {
				effect = parts[1]
			}
			taintArray = append(taintArray, eks2.NodeGroupTaintArgs{
				Key:    pulumi.String(k),
				Value:  pulumi.String(parts[0]),
				Effect: pulumi.String(effect),
			})
		}
		managedNodeGroupArgs.Taints = taintArray

		_, nodeGroupErr := eks.NewManagedNodeGroup(ctx, nodeGroupParam.Name, &managedNodeGroupArgs, eksProviders)
		if nodeGroupErr != nil {
			return nodeGroupErr
		}
	}
	return nil
}
