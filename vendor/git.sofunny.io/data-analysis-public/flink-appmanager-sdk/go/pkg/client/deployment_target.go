package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DeploymentTarget struct {
	Model
	Metadata *DeploymentTargetMetadata `json:"metadata,omitempty"`
	Spec     *DeploymentTargetSpec     `json:"spec,omitempty"`
}

type DeploymentTargetMetadata struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	CreatedAt       string            `json:"createdAt,omitempty"`
	ModifiedAt      string            `json:"modifiedAt,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	ResourceVersion int               `json:"resourceVersion,omitempty"`
}

type DeploymentTargetSpec struct {
	Kubernetes *KubernetesTarget `json:"kubernetes,omitempty"`
}

type KubernetesTarget struct {
	Namespace string `json:"namespace,omitempty"`
}

func (dt DeploymentTarget) String() string {
	marshal, _ := json.Marshal(dt)
	return string(marshal)
}

// GetDeploymentTargets List all deployment targets
func (c *Client) GetDeploymentTargets(namespace string) ([]DeploymentTarget, int, error) {
	deploymentTargets := &DeploymentTargetResourceList{}
	i, err := c.get(deploymentTargetUrl(namespace), deploymentTargets)
	if err != nil {
		return nil, i, err
	}
	return deploymentTargets.Items, i, nil
}

// GetDeploymentTarget Get a deployment target by name
func (c *Client) GetDeploymentTarget(name, namespace string) (*DeploymentTarget, int, error) {
	if len(name) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("name cannot be empty")
	}

	url := fmt.Sprintf("%s/%s", deploymentTargetUrl(namespace), name)

	dt := &DeploymentTarget{}
	i, err := c.get(url, dt)
	if err != nil {
		return nil, i, err
	}
	return dt, i, nil
}

// CreateDeploymentTarget Create a deployment target
// e.g. {"metadata":{"name":"default"},"spec":{"kubernetes":{"namespace":"default"}}}
func (c *Client) CreateDeploymentTarget(dt *DeploymentTarget, namespace string) (*DeploymentTarget, int, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(dt); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	i, err := c.post(deploymentTargetUrl(namespace), body, dt)
	if err != nil {
		return nil, i, err
	}
	return dt, i, nil
}

// DeleteDeploymentTarget Delete a deployment target
func (c *Client) DeleteDeploymentTarget(name, namespace string) (*DeploymentTarget, int, error) {
	if len(name) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("name cannot be empty")
	}

	url := fmt.Sprintf("%s/%s", deploymentTargetUrl(namespace), name)
	dt := &DeploymentTarget{}
	i, err := c.delete(url, dt)
	if err != nil {
		return nil, i, err
	}
	return dt, i, nil
}
