package core

import (
	"bytes"
	"encoding/json"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/go-collections/pkg/maps"
	"github.com/skhatri/pulumi-dsl-aws/pkg/k8s"
)

type EksSpec struct {
	Cluster k8s.EksCluster `json:"cluster" yaml:"cluster"`
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
	clusterArgs := eksSpec.Cluster
	return k8s.CreateEksCluster(ctx, clusterArgs)
}
