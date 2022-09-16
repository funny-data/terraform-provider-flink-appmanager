package client

type SessionClusterResourceList struct {
	Model
	Items []SessionCluster `json:"items"`
}

type DeploymentResourceList struct {
	Model
	Items []Deployment `json:"items"`
}

type SavepointResourceList struct {
	Model
	Items []Savepoint `json:"items"`
}

type JobResourceList struct {
	Model
	Items []Job `json:"items"`
}

type NamespaceResourceList struct {
	Model
	Items []Namespace `json:"items"`
}

type DeploymentTargetResourceList struct {
	Model
	Items []DeploymentTarget `json:"items"`
}

type ArtifactList struct {
	Model
	Items []Artifact `json:"items"`
}
