package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// SessionCluster SessionCluster Model
type SessionCluster struct {
	ID                   types.String             `tfsdk:"id"`
	Namespace            types.String             `tfsdk:"namespace"`
	Name                 types.String             `tfsdk:"name"`
	State                types.String             `tfsdk:"state"`
	DeploymentTargetName types.String             `tfsdk:"deployment_target_name"`
	FlinkVersion         types.String             `tfsdk:"flink_version"`
	FlinkImageTag        types.String             `tfsdk:"flink_image_tag"`
	NumberOfTaskManagers types.Int64              `tfsdk:"number_of_task_managers"`
	Resources            map[string]*ResourceSpec `tfsdk:"resources"`
	FlinkConfiguration   map[string]string        `tfsdk:"flink_configuration"`
}

// ResourceSpec 资源自定Model
type ResourceSpec struct {
	Cpu    types.Number `tfsdk:"cpu"`
	Memory types.String `tfsdk:"memory"`
}

// DeploymentTarget 部署目标Model
type DeploymentTarget struct {
	ID           types.String `tfsdk:"id"`
	Namespace    types.String `tfsdk:"namespace"`
	Name         types.String `tfsdk:"name"`
	K8SNamespace types.String `tfsdk:"k8s_namespace"`
}

// Namespace 部署空间Model
type Namespace struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	State types.String `tfsdk:"state"`
}
