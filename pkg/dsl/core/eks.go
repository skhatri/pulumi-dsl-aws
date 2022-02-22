package core

import (
	"bytes"
	"encoding/json"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/go-collections/pkg/maps"
	"github.com/skhatri/pulumi-dsl-aws/pkg/k8s"
)

type EksSpec struct {
	Cluster EksCluster `json:"cluster" yaml:"cluster"`
}

type EksCluster struct {
	Name            string `json:"name" yaml:"name"`
	Version         string `json:"version" yaml:"version"`
	DesiredCapacity int    `json:"desiredCapacity" yaml:"desiredCapacity"`
	InstanceType    string `json:"instanceType" yaml:"instanceType"`
	PublicKey       string `json:"publicKey" yaml:"publicKey"`
	VpcId           string `json:"vpcId" yaml:"vpcId"`
	SubnetId        string `json:"subnetId" yaml:"subnetId"`
	Tags            []Tag  `json:"tags" yaml:"tags"`
}

func EksHandler(ctx *pulumi.Context, pipelineItem PipelineItem) error {
	buff := bytes.Buffer{}
	m := maps.MapByStringKey(pipelineItem.Spec)
	encodeErr := json.NewEncoder(&buff).Encode(m)
	if encodeErr != nil {
		panic(encodeErr)
	}
	eksSpec := EksSpec{}
	json.NewDecoder(&buff).Decode(&eksSpec)
	cluster := eksSpec.Cluster
	clusterParams := k8s.ClusterParams{
		Name: cluster.Name,
	}
	if cluster.VpcId != "" {
		clusterParams.VpcId = cluster.VpcId
	}
	if cluster.SubnetId != "" {
		clusterParams.SubnetId = cluster.SubnetId
	}
	if cluster.DesiredCapacity != 0 {
		clusterParams.DesiredCapacity = cluster.DesiredCapacity
	}
	if cluster.InstanceType != "" {
		clusterParams.InstanceType = cluster.InstanceType
	}
	if cluster.Version != "" {
		clusterParams.Version = cluster.Version
	}
	if len(cluster.Tags) > 0 {
		tags := make(map[string]string)
		for _, tag := range cluster.Tags {
			tags[tag.Name] = tag.Value
		}
		clusterParams.Tags = tags
	}
	return k8s.CreateEksCluster(ctx, clusterParams)
}
