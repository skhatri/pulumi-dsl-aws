package dsl

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/pulumi-dsl-aws/pkg/dsl/core"
	"gopkg.in/yaml.v2"
	"os"
)

func processRequirements(ctx *pulumi.Context) error {
	var reader *os.File
	var ferr error
	fileName := "requirements.yaml"
	if _, localErr := os.Stat("requirements-local.yaml"); localErr == nil {
		fileName = "requirements-local.yaml"
	}
	reader, ferr = os.OpenFile(fileName, os.O_RDONLY, 0644)

	if ferr != nil {
		return fmt.Errorf("failed to open file: %v", ferr)
	}

	requirementPipeline := core.Requirement{}
	yaml.NewDecoder(reader).Decode(&requirementPipeline)

	for _, pipelineItem := range requirementPipeline.Pipeline {
		handler, ok := handlers[pipelineItem.CatalogName]
		if !ok {
			return fmt.Errorf("catalog item [%s] not found", pipelineItem.CatalogName)
		}
		err := handler(ctx, pipelineItem)
		if err != nil {
			return fmt.Errorf("error processing catalog [%s] with error [%v]", pipelineItem.CatalogName, err)
		}
	}
	return nil
}
