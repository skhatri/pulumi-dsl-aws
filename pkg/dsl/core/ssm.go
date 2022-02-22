package core

import (
	"bytes"
	"encoding/json"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/go-collections/pkg/maps"
	ssm2 "github.com/skhatri/pulumi-dsl-aws/pkg/ssm"
)

type SsmSpec struct {
	Add []AddOperation `json:"add" yaml:"add"`
}

type AddOperation struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

func SsmHandler(ctx *pulumi.Context, pipelineItem PipelineItem) error {
	buff := bytes.Buffer{}
	m := maps.MapByStringKey(pipelineItem.Spec)
	encodeErr := json.NewEncoder(&buff).Encode(m)
	if encodeErr != nil {
		panic(encodeErr)
	}
	ssmSpec := SsmSpec{}
	json.NewDecoder(&buff).Decode(&ssmSpec)

	var err error
	for _, addOp := range ssmSpec.Add {
		err = ssm2.PutSsmParameter(addOp.Key, addOp.Value, "String")(ctx)
		if err != nil {
			return err
		}
	}
	return err
}
