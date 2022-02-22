package core

import (
	"bytes"
	"encoding/json"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/skhatri/go-collections/pkg/maps"
)

type S3Spec struct {
	Bucket []BucketSpec `json:"buckets" yaml:"buckets"`
}

type BucketSpec struct {
	BucketName string    `json:"bucketName" yaml:"bucketName"`
	Region     string    `json:"region" yaml:"region"`
	Tags       []Tag `json:"tags" yaml:"tags"`
}

func S3Handler(ctx *pulumi.Context, pipelineItem PipelineItem) error {
	buff := bytes.Buffer{}
	m := maps.MapByStringKey(pipelineItem.Spec)
	encodeErr := json.NewEncoder(&buff).Encode(m)
	if encodeErr != nil {
		panic(encodeErr)
	}
	s3Spec := S3Spec{}
	json.NewDecoder(&buff).Decode(&s3Spec)
	bucketSpecs := s3Spec.Bucket

	var err error
	for _, bucketSpec := range bucketSpecs {
		tags := make(pulumi.StringMap, 0)
		for _, tag := range bucketSpec.Tags {
			tags[tag.Name] = pulumi.String(tag.Value)
		}
		_, err = s3.NewBucket(ctx, bucketSpec.BucketName, &s3.BucketArgs{
			Bucket: pulumi.String(bucketSpec.BucketName),
			Acl:    pulumi.String("private"),
			Tags:   tags,
		})
		if err != nil {
			return err
		}
	}
	return err
}
