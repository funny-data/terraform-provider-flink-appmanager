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

var _ resource.Resource = &NamespaceResource{}
var _ resource.ResourceWithImportState = &NamespaceResource{}

func NewNamespaceResource() resource.Resource {
	return &NamespaceResource{}
}

// NamespaceResource defines the resource implementation.
type NamespaceResource struct {
	client *client.Client
}

func (r *NamespaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *NamespaceResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"state": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (r *NamespaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create 创建部署空间
func (r *NamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// 获取配置参数
	var plan NamespaceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 创建部署空间
	namespaceName := plan.Name.Value
	_, _, err := r.client.CreateNamespace(namespaceName)
	if err != nil {
		resp.Diagnostics.AddError("Error creating namespace", "Could not create namespace, unexpected error: "+err.Error())
		return
	}

	// 等待部署空间创建
	namespaceState, _, err := r.client.WaitNamespaceStateChange(namespaceName, client.NamespaceActive)
	if err != nil {
		resp.Diagnostics.AddError("Error namespace state change", "Could not namespace state change, unexpected error: "+err.Error())
		return
	}

	var result = NamespaceResourceModel{
		ID:    types.String{Value: namespaceState.Metadata.Id},
		Name:  types.String{Value: namespaceState.Metadata.Name},
		State: types.String{Value: namespaceState.Status.State},
	}

	// 保存状态
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Read 读取部署空间
func (r *NamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 获取状态参数
	var state NamespaceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 获取部署空间详情
	namespace, _, err := r.client.GetNamespace(state.Name.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error reading namespace", "Could not read namespace: "+err.Error())
		return
	}

	var result = NamespaceResourceModel{
		ID:    types.String{Value: namespace.Metadata.Id},
		Name:  types.String{Value: namespace.Metadata.Name},
		State: types.String{Value: namespace.Status.State},
	}

	// 部署空间写入状态
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *NamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

// Delete 删除部署空间
func (r *NamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// 获取状态参数
	var state NamespaceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 删除部署空间
	err := r.client.DeleteNamespaceCompleted(state.Name.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error Delete namespace", "Could not delete namespace, unexpected error: "+err.Error())
		return
	}
}

// ImportState 导入状态
func (r *NamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
