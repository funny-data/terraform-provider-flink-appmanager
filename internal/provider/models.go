package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// SessionClusterResourceModel SessionClusterResourceModel Model
type SessionClusterResourceModel struct {
	ID                   types.String             `tfsdk:"id"`
	Namespace            types.String             `tfsdk:"namespace"`
	Name                 types.String             `tfsdk:"name"`
	State                types.String             `tfsdk:"state"`
	DeploymentTargetName types.String             `tfsdk:"deployment_target_name"`
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

// DeploymentTargetResourceModel 部署目标Model
type DeploymentTargetResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Namespace    types.String `tfsdk:"namespace"`
	Name         types.String `tfsdk:"name"`
	K8SNamespace types.String `tfsdk:"k8s_namespace"`
}

// NamespaceResourceModel 部署空间Model
type NamespaceResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	State types.String `tfsdk:"state"`
}
