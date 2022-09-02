package provider

import (
	"context"
	"git.sofunny.io/data-analysis/flink-app/anti-cheat-panel/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math/big"
)

var _ provider.ResourceType = sessionClusterResourceType{}
var _ resource.Resource = sessionClusterResource{}
var _ resource.ResourceWithImportState = sessionClusterResource{}

type sessionClusterResourceType struct {
}

func (r sessionClusterResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"state": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"deployment_target_name": {
				Type:     types.StringType,
				Optional: true,
			},
			"flink_version": {
				Type:     types.StringType,
				Required: true,
			},
			"flink_image_tag": {
				Type:     types.StringType,
				Required: true,
			},
			"number_of_task_managers": {
				Type:     types.Int64Type,
				Required: true,
			},
			"resources": {
				Type: types.MapType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"cpu":    types.NumberType,
					"memory": types.StringType,
				}}},
				Required: true,
			},
			"flink_configuration": {
				Type:     types.MapType{ElemType: types.StringType},
				Required: true,
			},
		},
	}, nil
}

func (r sessionClusterResourceType) NewResource(_ context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	return sessionClusterResource{
		provider: p,
	}, diags
}

type sessionClusterResource struct {
	provider appManagerProvider
}

// Create 创建运行集群
func (r sessionClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// 读取配置
	var plan SessionCluster
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 创建SessionCluster集群
	sc, err := r.RunSessionCluster(buildSessionClusterDTO(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Error create sessionCluster", "could not create sessionCluster, unexpected error: "+err.Error())
	}

	// 根据SessionCluster集群信息构建tf值
	var result = buildSessionClusterTfValue(sc)
	// 写出集群状态
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read 读取集群信息
func (r sessionClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 获取状态参数
	var state SessionCluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 查询SessionCluster集群信息
	sessionCluster, _, err := r.provider.client.GetSessionCluster(state.Name.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error reading sessionCluster", "Could not read sessionCluster, unexpected error:: "+err.Error())
		return
	}

	// 根据SessionCluster集群信息构建tf值
	var result = buildSessionClusterTfValue(sessionCluster)

	// 写出集群状态
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update 更新集群
func (r sessionClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// 获取状态参数
	var state SessionCluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.Value
	_, err := r.StopSessionCluster(name)
	if err != nil {
		resp.Diagnostics.AddError("Error stop sessionCluster", "Could stop sessionCluster, unexpected error: "+err.Error())
		return
	}

	// 读取配置
	var plan SessionCluster
	planDiags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 创建SessionCluster集群
	sc, err := r.RunSessionCluster(buildSessionClusterDTO(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Error create sessionCluster", "could not create sessionCluster, unexpected error: "+err.Error())
	}

	// 根据SessionCluster集群信息构建tf值
	var result = buildSessionClusterTfValue(sc)
	// 写出集群状态
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete 删除集群
func (r sessionClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SessionCluster
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sessionClusterName := state.Name.Value
	// 停止SessionCluster
	_, err := r.StopSessionCluster(sessionClusterName)
	if err != nil {
		resp.Diagnostics.AddError("Error stop sessionCluster", "Could not stop sessionCluster, unexpected error: "+err.Error())
		return
	}

	// 删除集群名称
	_, _, err = r.provider.client.DeleteSessionCluster(sessionClusterName)
	if err != nil {
		resp.Diagnostics.AddError("Error delete sessionCluster", "Could not delete sessionCluster, unexpected error: "+err.Error())
		return
	}
}

// ImportState 导入状态
func (r sessionClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// 将sessionCluster值转换成tf值
func buildSessionClusterTfValue(sc *client.SessionCluster) *SessionCluster {
	resources := make(map[string]*ResourceSpec)
	for k, v := range sc.Spec.Resources {
		resources[k] = &ResourceSpec{
			Cpu:    types.Number{Value: big.NewFloat(v.Cpu)},
			Memory: types.String{Value: v.Memory},
		}
	}

	return &SessionCluster{
		ID:                   types.String{Value: sc.Metadata.Id},
		Name:                 types.String{Value: sc.Metadata.Name},
		State:                types.String{Value: sc.Status.State},
		DeploymentTargetName: types.String{Value: sc.Spec.DeploymentTargetName},
		FlinkVersion:         types.String{Value: sc.Spec.FlinkVersion},
		FlinkImageTag:        types.String{Value: sc.Spec.FlinkImageTag},
		NumberOfTaskManagers: types.Int64{Value: int64(sc.Spec.NumberOfTaskManagers)},
		Resources:            resources,
		FlinkConfiguration:   sc.Spec.FlinkConfiguration,
	}
}

// 将tf值转换成sessionCluster请求参数
func buildSessionClusterDTO(sc *SessionCluster) *client.SessionCluster {
	resources := make(map[string]*client.ResourceSpec)
	for k, v := range sc.Resources {
		cpu, _ := v.Cpu.Value.Float64()
		resources[k] = &client.ResourceSpec{
			Cpu:    cpu,
			Memory: v.Memory.Value,
		}
	}

	return &client.SessionCluster{
		Metadata: &client.SessionClusterMetadata{Name: sc.Name.Value},
		Spec: &client.SessionClusterSpec{
			State:                sc.State.Value,
			DeploymentTargetName: sc.DeploymentTargetName.Value,
			FlinkVersion:         sc.FlinkVersion.Value,
			FlinkImageTag:        sc.FlinkImageTag.Value,
			NumberOfTaskManagers: int(sc.NumberOfTaskManagers.Value),
			FlinkConfiguration:   sc.FlinkConfiguration,
			Resources:            resources,
		},
	}
}

// StopSessionCluster 停止SessionCluster
func (r sessionClusterResource) StopSessionCluster(sessionClusterName string) (*client.SessionCluster, error) {
	// 停止SessionCluster
	sc := &client.SessionCluster{
		Metadata: &client.SessionClusterMetadata{Name: sessionClusterName},
		Spec:     &client.SessionClusterSpec{State: client.ClusterStopped},
	}
	_, _, err := r.provider.client.UpdateSessionCluster(sc)
	if err != nil {
		return nil, err
	}

	// 等待SessionCluster停止
	sc, _, err = r.provider.client.WaitSessionClusterStateChange(sessionClusterName, client.ClusterStopped)
	if err != nil {
		return nil, err
	}

	return sc, nil
}

// RunSessionCluster 创建出运行的SessionCluster
func (r sessionClusterResource) RunSessionCluster(scCfg *client.SessionCluster) (*client.SessionCluster, error) {
	// 写死使用默认日志配置
	scCfg.Spec.Logging = &client.Logging{Log4jLoggers: map[string]string{"": "INFO"}, LoggingProfile: "default"}
	// 强制启动
	scCfg.Spec.State = client.ClusterRunning

	_, _, err := r.provider.client.CreateOrReplaceSessionCluster(scCfg)
	if err != nil {
		return nil, err
	}

	// 等待集群创建
	state, _, err := r.provider.client.WaitSessionClusterStateChange(scCfg.Metadata.Name, client.ClusterRunning)
	if err != nil {
		return nil, err
	}

	return state, nil
}
