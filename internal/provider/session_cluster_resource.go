package provider

import (
	"context"
	"fmt"
	"git.sofunny.io/data-analysis-public/flink-appmanager-sdk/go/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"math/big"
	"strings"
)

var _ resource.Resource = &SessionClusterResource{}
var _ resource.ResourceWithImportState = &SessionClusterResource{}

func NewSessionClusterResource() resource.Resource {
	return &SessionClusterResource{}
}

// SessionClusterResource defines the resource implementation.
type SessionClusterResource struct {
	client *client.Client
}

func (r *SessionClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_session_cluster"
}

func (r *SessionClusterResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"namespace": {
				Type:     types.StringType,
				Required: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"deployment_target_name": {
				Type:     types.StringType,
				Optional: true,
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

func (r *SessionClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create 创建运行集群
func (r *SessionClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// 读取配置
	var plan SessionClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 创建SessionCluster集群
	sc, err := r.RunSessionCluster(plan.Namespace.Value, buildSessionClusterDTO(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Error create sessionCluster", "could not create sessionCluster, unexpected error: "+err.Error())
	}

	// 根据SessionCluster集群信息构建tf值
	var result = buildSessionClusterTfValue(sc)
	// 写出集群状态
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Read 读取集群信息
func (r *SessionClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 获取状态参数
	var state SessionClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 查询SessionCluster集群信息
	sessionCluster, _, err := r.client.GetSessionCluster(state.Name.Value, state.Namespace.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error reading sessionCluster", "Could not read sessionCluster, unexpected error:: "+err.Error())
		return
	}

	// 根据SessionCluster集群信息构建tf值
	var result = buildSessionClusterTfValue(sessionCluster)

	// 写出集群状态
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Update 更新集群
func (r *SessionClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// 获取状态参数
	var state SessionClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.Value
	_, err := r.StopSessionCluster(name, state.Namespace.Value)
	if err != nil {
		resp.Diagnostics.AddError("Error stop sessionCluster", "Could stop sessionCluster, unexpected error: "+err.Error())
		return
	}

	// 读取配置
	var plan SessionClusterResourceModel
	planDiags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 创建SessionCluster集群
	sc, err := r.RunSessionCluster(plan.Namespace.Value, buildSessionClusterDTO(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Error create sessionCluster", "could not create sessionCluster, unexpected error: "+err.Error())
	}

	// 根据SessionCluster集群信息构建tf值
	var result = buildSessionClusterTfValue(sc)
	// 写出集群状态
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Delete 删除集群
func (r *SessionClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SessionClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sessionClusterName := state.Name.Value
	// 停止SessionCluster
	_, err := r.StopSessionCluster(state.Namespace.Value, sessionClusterName)
	if err != nil {
		resp.Diagnostics.AddError("Error stop sessionCluster", "Could not stop sessionCluster, unexpected error: "+err.Error())
		return
	}

	// 删除集群名称
	_, _, err = r.client.DeleteSessionCluster(sessionClusterName, state.Namespace.Value)

	if err != nil {
		resp.Diagnostics.AddError("Error delete sessionCluster", "Could not delete sessionCluster, unexpected error: "+err.Error())
		return
	}
}

// ImportState 导入状态
func (r *SessionClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: namespace,sessionCluseteName. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[1])...)
}

// 将sessionCluster值转换成tf值
func buildSessionClusterTfValue(sc *client.SessionCluster) *SessionClusterResourceModel {
	resources := make(map[string]*ResourceSpec)
	for k, v := range sc.Spec.Resources {
		resources[k] = &ResourceSpec{
			Cpu:    types.Number{Value: big.NewFloat(v.Cpu)},
			Memory: types.String{Value: v.Memory},
		}
	}

	return &SessionClusterResourceModel{
		ID:                   types.String{Value: sc.Metadata.Id},
		Namespace:            types.String{Value: sc.Metadata.Namespace},
		Name:                 types.String{Value: sc.Metadata.Name},
		State:                types.String{Value: sc.Status.State},
		DeploymentTargetName: types.String{Value: sc.Spec.DeploymentTargetName},
		FlinkImageTag:        types.String{Value: sc.Spec.FlinkImageTag},
		NumberOfTaskManagers: types.Int64{Value: int64(sc.Spec.NumberOfTaskManagers)},
		Resources:            resources,
		FlinkConfiguration:   sc.Spec.FlinkConfiguration,
	}
}

// 将tf值转换成sessionCluster请求参数
func buildSessionClusterDTO(sc *SessionClusterResourceModel) *client.SessionCluster {
	resources := make(map[string]*client.ResourceSpec)
	for k, v := range sc.Resources {
		cpu, _ := v.Cpu.Value.Float64()
		resources[k] = &client.ResourceSpec{
			Cpu:    cpu,
			Memory: v.Memory.Value,
		}
	}

	return &client.SessionCluster{
		Metadata: &client.SessionClusterMetadata{Name: sc.Name.Value, Namespace: sc.Namespace.Value},
		Spec: &client.SessionClusterSpec{
			State:                sc.State.Value,
			DeploymentTargetName: sc.DeploymentTargetName.Value,
			FlinkImageTag:        sc.FlinkImageTag.Value,
			NumberOfTaskManagers: int(sc.NumberOfTaskManagers.Value),
			FlinkConfiguration:   sc.FlinkConfiguration,
			Resources:            resources,
		},
	}
}

// StopSessionCluster 停止SessionCluster
func (r *SessionClusterResource) StopSessionCluster(namespace string, sessionClusterName string) (*client.SessionCluster, error) {
	// 停止SessionCluster
	sc := &client.SessionCluster{
		Metadata: &client.SessionClusterMetadata{Name: sessionClusterName, Namespace: namespace},
		Spec:     &client.SessionClusterSpec{State: client.ClusterStopped},
	}
	_, _, err := r.client.UpdateSessionCluster(sc, namespace)
	if err != nil {
		return nil, err
	}

	// 等待SessionCluster停止
	sc, _, err = r.client.WaitSessionClusterStateChange(sessionClusterName, client.ClusterStopped, namespace)
	if err != nil {
		return nil, err
	}

	return sc, nil
}

// RunSessionCluster 创建出运行的SessionCluster
func (r *SessionClusterResource) RunSessionCluster(namespace string, scCfg *client.SessionCluster) (*client.SessionCluster, error) {
	// 写死使用默认日志配置
	scCfg.Spec.Logging = &client.Logging{Log4jLoggers: map[string]string{"": "INFO"}, LoggingProfile: "default"}
	// 强制启动
	scCfg.Spec.State = client.ClusterRunning

	_, _, err := r.client.CreateOrReplaceSessionCluster(scCfg, namespace)
	if err != nil {
		return nil, err
	}

	// 等待集群创建
	state, _, err := r.client.WaitSessionClusterStateChange(scCfg.Metadata.Name, client.ClusterRunning, namespace)
	if err != nil {
		return nil, err
	}

	return state, nil
}
