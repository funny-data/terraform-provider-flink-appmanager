package client

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type DeploymentDefaults struct {
	Model
	Metadata *DeploymentDefaultsMetadata `json:"metadata,omitempty"`
	Spec     *DeploymentSpec             `json:"spec,omitempty"`
}

type DeploymentDefaultsMetadata struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	CreatedAt       string            `json:"createdAt,omitempty"`
	ModifiedAt      string            `json:"modifiedAt,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	ResourceVersion int               `json:"resourceVersion,omitempty"`
}

func (dd DeploymentDefaults) String() string {
	marshal, _ := json.Marshal(dd)
	return string(marshal)
}

// GetDeploymentDefaults 获取默认部署配置
func (c *Client) GetDeploymentDefaults(namespace string) (*DeploymentDefaults, int, error) {
	dd := &DeploymentDefaults{}
	i, err := c.get(deploymentDefaultsUrl(namespace), dd)
	if err != nil {
		return nil, i, err
	}
	return dd, i, nil
}

// CoverDeploymentDefaults 覆盖当前默认部署配置
func (c *Client) CoverDeploymentDefaults(dd *DeploymentDefaults, namespace string) (*DeploymentDefaults, int, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(dd); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	i, err := c.put(deploymentDefaultsUrl(namespace), body, dd)
	if err != nil {
		return nil, i, err
	}
	return dd, i, nil
}

// UpdateDeploymentDefaults 更新默认部署配置，不更新空的属性
func (c *Client) UpdateDeploymentDefaults(dd *DeploymentDefaults, namespace string) (*DeploymentDefaults, int, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(dd); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	i, err := c.patch(deploymentDefaultsUrl(namespace), body, dd)
	if err != nil {
		return nil, i, err
	}
	return dd, i, nil
}
