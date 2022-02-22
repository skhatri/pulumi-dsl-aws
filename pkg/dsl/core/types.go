package core

type Tag struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type PipelineItem struct {
	CatalogName string                 `json:"catalog" yaml:"catalog"`
	Name        string                 `json:"name" yaml:"name"`
	Spec        map[string]interface{} `json:"spec" yaml:"spec"`
}

type Requirement struct {
	Pipeline []PipelineItem `json:"pipeline" yaml:"pipeline"`
}

