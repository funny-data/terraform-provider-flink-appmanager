package provider

import (
	"context"
	"git.sofunny.io/data-analysis/flink-app/anti-cheat-panel/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.ResourceType = namespaceResourceType{}
var _ resource.Resource = namespaceResource{}
var _ resource.ResourceWithImportState = namespaceResource{}

type namespaceResourceType struct{}

func (r namespaceResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (r namespaceResourceType) NewResource(_ context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return namespaceResource{
		provider: p,
	}, diags
}

type namespaceResource struct {
	provider appManagerProvider
}

// Create 创建部署空间
func (r namespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// 获取配置参数
	var plan Namespace
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 创建部署空间
	namespaceName := plan.Name.Value
	_, _, err := r.provider.client.CreateNamespace(namespaceName)
	if err != nil {
		resp.Diagnostics.AddError("Error creating namespace", "Could not create namespace, unexpected error: "+err.Error())
		return
	}

	// 等待部署空间创建
	namespaceState, _, err := r.provider.client.WaitNamespaceStateChange(namespaceName, client.NamespaceActive)
	if err != nil {
		resp.Diagnostics.AddError("Error namespace state change", "Could not namespace state change, unexpected error: "+err.Error())
		return
	}

	var result = Namespace{
		ID:    types.String{Value: namespaceState.Metadata.Id},
		Name:  types.String{Value: namespaceState.Metadata.Name},
		State: types.String{Value: namespaceState.Status.State},
	}

	// 保存状态
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read 读取部署空间
func (r namespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 获取状态参数
	var state Namespace
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 获取部署空间详情
	namespace, _, err := r.provider.client.GetNamespace(state.Name.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error reading namespace", "Could not read namespace: "+err.Error())
		return
	}

	var result = Namespace{
		ID:    types.String{Value: namespace.Metadata.Id},
		Name:  types.String{Value: namespace.Metadata.Name},
		State: types.String{Value: namespace.Status.State},
	}

	// 部署空间写入状态
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r namespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

// Delete 删除部署空间
func (r namespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// 获取状态参数
	var state Namespace
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 删除部署空间
	err := r.provider.client.DeleteNamespaceCompleted(state.Name.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error Delete namespace", "Could not delete namespace, unexpected error: "+err.Error())
		return
	}
}

// ImportState 导入状态
func (r namespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
