package provider

import (
	"context"
	"fmt"
	"git.sofunny.io/data-analysis-public/flink-appmanager-sdk/go/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DeploymentTargetResource{}
var _ resource.ResourceWithImportState = &DeploymentTargetResource{}

const (
	DefaultK8SNamespace = "default"
)

func NewDeploymentTargetResource() resource.Resource {
	return &DeploymentTargetResource{}
}

// DeploymentTargetResource defines the resource implementation.
type DeploymentTargetResource struct {
	client *client.Client
}

func (r *DeploymentTargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_target"
}

func (r *DeploymentTargetResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (r *DeploymentTargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = c
}

// Create 创建部署目标
func (r *DeploymentTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// 获取配置参数
	var plan DeploymentTargetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
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
	deploymentTarget, _, err := r.client.CreateDeploymentTarget(dt, plan.Namespace.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error create deploymentTarget", "Could not create deploymentTarget, unexpected error: "+err.Error())
		return
	}

	var result = DeploymentTargetResourceModel{
		ID:           types.String{Value: deploymentTarget.Metadata.ID},
		Namespace:    types.String{Value: deploymentTarget.Metadata.Namespace},
		Name:         types.String{Value: deploymentTarget.Metadata.Name},
		K8SNamespace: types.String{Value: deploymentTarget.Spec.Kubernetes.Namespace},
	}

	// 保存状态
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Read 读取部署目标
func (r *DeploymentTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 获取状态参数
	var state DeploymentTargetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deploymentTarget, _, err := r.client.GetDeploymentTarget(state.Name.Value, state.Namespace.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error reading deploymentTarget", "Could not read deploymentTarget: "+err.Error())
		return
	}

	var result = DeploymentTargetResourceModel{
		ID:           types.String{Value: deploymentTarget.Metadata.ID},
		Name:         types.String{Value: deploymentTarget.Metadata.Name},
		Namespace:    types.String{Value: deploymentTarget.Metadata.Namespace},
		K8SNamespace: types.String{Value: deploymentTarget.Spec.Kubernetes.Namespace},
	}

	// 部署目标写入状态
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *DeploymentTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

// Delete 删除部署目标
func (r *DeploymentTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// 获取状态参数
	var state DeploymentTargetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 删除部署目标
	_, _, err := r.client.DeleteDeploymentTarget(state.Name.Value, state.Namespace.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error delete deploymentTarget", "Could not deleted deploymentTarget, unexpected error: "+err.Error())
		return
	}
}

// ImportState 导入状态
func (r *DeploymentTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
