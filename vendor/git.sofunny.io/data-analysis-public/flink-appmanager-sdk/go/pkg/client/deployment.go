package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Deployment struct {
	Model
	Metadata *DeploymentMetadata `json:"metadata,omitempty"`
	Spec     *DeploymentSpec     `json:"spec,omitempty"`
	Status   *DeploymentStatus   `json:"status,omitempty"`
}

type DeploymentMetadata struct {
	Id              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	DisplayName     string            `json:"displayName,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	CreateAt        *time.Time        `json:"createAt,omitempty"`
	ModifiedAt      *time.Time        `json:"modifiedAt,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	ResourceVersion int               `json:"resourceVersion,omitempty"`
}

type DeploymentSpec struct {
	State                        string              `json:"state,omitempty"`
	UpgradeStrategy              *UpgradeStrategy    `json:"upgradeStrategy,omitempty"`
	RestoreStrategy              *RestoreStrategy    `json:"restoreStrategy,omitempty"`
	DeploymentTargetId           string              `json:"deploymentTargetId,omitempty"`
	DeploymentTargetIName        string              `json:"deploymentTargetIName,omitempty"`
	SessionClusterName           string              `json:"sessionClusterName,omitempty"`
	MaxSavepointCreationAttempts int                 `json:"maxSavepointCreationAttempts,omitempty"`
	MaxJobCreationAttempts       int                 `json:"maxJobCreationAttempts,omitempty"`
	Template                     *DeploymentTemplate `json:"template,omitempty"`
}

type UpgradeStrategy struct {
	Kind string `json:"kind,omitempty"`
}

type RestoreStrategy struct {
	Kind                  string `json:"kind,omitempty"`
	AllowNonRestoredState bool   `json:"allowNonRestoredState,omitempty"`
}

type DeploymentTemplate struct {
	Metadata *DeploymentTemplateMetadata `json:"metadata,omitempty"`
	Spec     *DeploymentTemplateSpec     `json:"spec,omitempty"`
}

type DeploymentTemplateMetadata struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}

type DeploymentTemplateSpec struct {
	Artifact             *JarArtifact             `json:"artifact,omitempty"`
	Parallelism          int                      `json:"parallelism,omitempty"`
	NumberOfTaskManagers int                      `json:"numberOfTaskManagers,omitempty"`
	Resources            map[string]*ResourceSpec `json:"resources,omitempty"`
	FlinkConfiguration   map[string]string        `json:"flinkConfiguration,omitempty"`
}

type DeploymentStatus struct {
	State   string                   `json:"state,omitempty"`
	Running *DeploymentStatusRunning `json:"running,omitempty"`
}

type DeploymentStatusRunning struct {
	JobId          string                 `json:"jobId,omitempty"`
	TransitionTime *time.Time             `json:"transitionTime,omitempty"`
	Conditions     []*DeploymentCondition `json:"conditions,omitempty"`
}

type DeploymentCondition struct {
	ConditionType      string `json:"type,omitempty"`
	Status             string `json:"status,omitempty"`
	Message            string `json:"message,omitempty"`
	Reason             string `json:"reason,omitempty"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     string `json:"lastUpdateTime,omitempty"`
}

func (d Deployment) String() string {
	marshal, _ := json.Marshal(d)
	return string(marshal)
}

const (
	LabelSelector = "labelSelector"
)

// deployment state
const (
	// A Flink job was successfully started as specified in the deployment.spec and the Flink job has not been terminated.
	DeploymentRunning = "RUNNING"
	// The Deployment has successfully created a Savepoint and terminated the job.
	DeploymentSuspended = "SUSPENDED"
	// Any previously active Flink job was terminated via cancellation.
	DeploymentCancelled = "CANCELLED"

	// The Deployment is currently in transition to achieve the desired state.
	// 仅实际状态
	DeploymentTransitioning = "TRANSITIONING"
	// A Deployment transition has failed.
	// 仅实际状态
	DeploymentFailed = "FAILED"
	// The Flink job was finite and finished execution successfully, e.g. a finite streaming or a batch job.
	// 仅实际状态
	DeploymentFinished = "FINISHED"
)

// Deployment Conditions
const (
	// The Flink job corresponding to the running Deployment was last observed in an unknown status.
	DeploymentConditionClusterUnreachable = "ClusterUnreachable"
	// The Flink job corresponding to the running Deployment was last observed to have restarted 3 or more times within the last 10 minutes.
	DeploymentConditionJobFailing = "JobFailing"
	// The Flink job corresponding to the running Deployment was last observed to have restarted 3 or more times within the last 60 minutes. Note that JobUnstable currently implies JobFailing.
	DeploymentConditionJobUnstable = "JobUnstable"
)

// upgrade strategy
const (
	DeploymentUpgradeStrategyNone      = "NONE"
	DeploymentUpgradeStrategyStateless = "STATELESS"
	DeploymentUpgradeStrategyStateful  = "STATEFUL"
)

// restore strategy
const (
	DeploymentRestoreStrategyNone            = "NONE"
	DeploymentRestoreStrategyLatestState     = "LATEST_STATE"
	DeploymentRestoreStrategyLatestSavepoint = "LATEST_SAVEPOINT"
)

// GetDeployments list all deployments
func (c *Client) GetDeployments(labels map[string]string, namespace string) ([]Deployment, int, error) {
	u, err := urlWithLabels(deploymentUrl(namespace), labels)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	deploymentResourceList := &DeploymentResourceList{}
	code, err := c.get(u, deploymentResourceList)
	if err != nil {
		return nil, code, err
	}

	return deploymentResourceList.Items, code, nil
}

// urlWithLabels put labelSelector into url
func urlWithLabels(u string, labels map[string]string) (string, error) {
	query := map[string]string{
		LabelSelector: mapToString(labels),
	}
	return UrlWithQuery(u, query)
}

// mapToString transform the map to string
// e.g. {"a":"b","c":"d"}  =>  a=b,c=d
func mapToString(m map[string]string) string {
	items := make([]string, len(m))
	index := 0
	for key, value := range m {
		items[index] = fmt.Sprintf("%s=%s", key, value)
		index++
	}
	return strings.Join(items, ",")
}

// GetDeployment get the deployment by name, not id
func (c *Client) GetDeployment(name, namespace string) (*Deployment, int, error) {
	u := fmt.Sprintf("%s/%s", deploymentUrl(namespace), name)

	deployment := &Deployment{}
	code, err := c.get(u, deployment)
	if err != nil {
		return nil, code, err
	}

	return deployment, code, nil
}

// DeleteDeployment delete the deployment by name, not id
func (c *Client) DeleteDeployment(name, namespace string) (*Deployment, int, error) {
	u := fmt.Sprintf("%s/%s", deploymentUrl(namespace), name)

	deployment := &Deployment{}
	code, err := c.delete(u, deployment)
	if err != nil {
		return nil, code, err
	}

	return deployment, code, nil
}

// CreateDeployment create a deployment
func (c *Client) CreateDeployment(deployment *Deployment, namespace string) (*Deployment, int, error) {
	deployment.Metadata.Namespace = namespace

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(deployment); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	code, err := c.post(deploymentUrl(namespace), body, deployment)
	if err != nil {
		return nil, code, err
	}
	return deployment, code, nil
}

// CreateOrReplaceDeployment create or replace the deployment
func (c *Client) CreateOrReplaceDeployment(deployment *Deployment, namespace string) (*Deployment, int, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(deployment); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	u := fmt.Sprintf("%s/%s", deploymentUrl(namespace), deployment.Metadata.Name)
	code, err := c.put(u, body, deployment)
	if err != nil {
		return nil, code, err
	}
	return deployment, code, nil
}

// UpdateDeployment Update a deployment
func (c *Client) UpdateDeployment(deployment *Deployment, namespace string) (*Deployment, int, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(deployment); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	u := fmt.Sprintf("%s/%s", deploymentUrl(namespace), deployment.Metadata.Name)
	code, err := c.patch(u, body, deployment)
	if err != nil {
		return nil, code, err
	}
	return deployment, code, nil
}

// WaitDeploymentStateChange 等待 Deployment 状态扭转完成
func (c *Client) WaitDeploymentStateChange(name, state, namespace string) (*Deployment, int, error) {
	validateState := func() bool {
		switch state {
		case DeploymentRunning, DeploymentSuspended, DeploymentCancelled,
			DeploymentTransitioning, DeploymentFailed, DeploymentFinished:
			return true
		default:
			return false
		}
	}

	getState := func() (interface{}, string, int, error) {
		d, i, err := c.GetDeployment(name, namespace)
		if err != nil {
			return nil, "", i, err
		}
		return d, d.Status.State, i, nil
	}

	d, i, err := c.waitStateChange(state, validateState, getState)
	if err != nil {
		return nil, i, err
	}
	return d.(*Deployment), i, nil
}
