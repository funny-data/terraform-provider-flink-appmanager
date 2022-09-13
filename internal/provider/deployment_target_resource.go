package provider

import (
	"context"
	"git.sofunny.io/data-analysis-public/flink-appmanager-sdk/go/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.ResourceType = deploymentTargetResourceType{}
var _ resource.Resource = deploymentTargetResource{}
var _ resource.ResourceWithImportState = deploymentTargetResource{}

const (
	DefaultK8SNamespace = "default"
)

type deploymentTargetResourceType struct {
}

func (r deploymentTargetResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"namespace": {
				Type:     types.StringType,
				Required: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"k8s_namespace": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (r deploymentTargetResourceType) NewResource(_ context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return deploymentTargetResource{
		provider: p,
	}, diags
}

type deploymentTargetResource struct {
	provider appManagerProvider
}

// Create 创建部署目标
func (r deploymentTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// 获取配置参数
	var plan DeploymentTarget
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	k8sNamespace := plan.K8SNamespace.Value
	if k8sNamespace == "" {
		k8sNamespace = DefaultK8SNamespace
	}

	// 创建部署目标
	dt := &client.DeploymentTarget{
		Metadata: &client.DeploymentTargetMetadata{Name: plan.Name.Value, Namespace: plan.Namespace.Value},
		Spec: &client.DeploymentTargetSpec{Kubernetes: &client.KubernetesTarget{
			Namespace: k8sNamespace,
		}},
	}
	deploymentTarget, _, err := r.provider.client.CreateDeploymentTarget(dt, plan.Namespace.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error create deploymentTarget", "Could not create deploymentTarget, unexpected error: "+err.Error())
		return
	}

	var result = DeploymentTarget{
		ID:           types.String{Value: deploymentTarget.Metadata.ID},
		Namespace:    types.String{Value: deploymentTarget.Metadata.Namespace},
		Name:         types.String{Value: deploymentTarget.Metadata.Name},
		K8SNamespace: types.String{Value: deploymentTarget.Spec.Kubernetes.Namespace},
	}

	// 保存状态
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read 读取部署目标
func (r deploymentTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 获取状态参数
	var state DeploymentTarget
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deploymentTarget, _, err := r.provider.client.GetDeploymentTarget(state.Name.Value, state.Namespace.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error reading deploymentTarget", "Could not read deploymentTarget: "+err.Error())
		return
	}

	var result = DeploymentTarget{
		ID:           types.String{Value: deploymentTarget.Metadata.ID},
		Name:         types.String{Value: deploymentTarget.Metadata.Name},
		Namespace:    types.String{Value: deploymentTarget.Metadata.Namespace},
		K8SNamespace: types.String{Value: deploymentTarget.Spec.Kubernetes.Namespace},
	}

	// 部署目标写入状态
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r deploymentTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

// Delete 删除部署目标
func (r deploymentTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// 获取状态参数
	var state DeploymentTarget
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 删除部署目标
	_, _, err := r.provider.client.DeleteDeploymentTarget(state.Name.Value, state.Namespace.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error delete deploymentTarget", "Could not deleted deploymentTarget, unexpected error: "+err.Error())
		return
	}
}

// ImportState 导入状态
func (r deploymentTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
