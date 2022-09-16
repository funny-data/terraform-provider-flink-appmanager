package client

import "time"

type Model struct {
	Kind       string `json:"kind,omitempty"`
	ApiVersion string `json:"apiVersion,omitempty"`
}

type Failure struct {
	Message  string     `json:"message"`
	Reason   string     `json:"reason"`
	FailedAt *time.Time `json:"failedAt"`
}

type JarArtifact struct {
	Kind                   string   `json:"kind,omitempty"`
	JarUri                 string   `json:"jarUri,omitempty"`
	MainArgs               string   `json:"mainArgs,omitempty"`
	EntryClass             string   `json:"entryClass,omitempty"`
	AdditionalDependencies []string `json:"additionalDependencies,omitempty"`
	FlinkVersion           string   `json:"flinkVersion,omitempty"`
	FlinkImageRegistry     string   `json:"flinkImageRegistry,omitempty"`
	FlinkImageRepository   string   `json:"flinkImageRepository,omitempty"`
	FlinkImageTag          string   `json:"flinkImageTag,omitempty"`
}

type Logging struct {
	LoggingProfile              string            `json:"loggingProfile,omitempty"`
	Log4j2ConfigurationTemplate string            `json:"log4j2ConfigurationTemplate,omitempty"`
	Log4jLoggers                map[string]string `json:"log4jLoggers,omitempty"`
}

const (
	AnnotationPrefix = "com.xmfunny.flink"
)
