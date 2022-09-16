package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bhmj/jsonslice"
)

type SessionCluster struct {
	Model
	Metadata *SessionClusterMetadata `json:"metadata,omitempty"`
	Spec     *SessionClusterSpec     `json:"spec,omitempty"`
	Status   *SessionClusterStatus   `json:"status,omitempty"`
}

type SessionClusterMetadata struct {
	Id              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	CreatedAt       string            `json:"createdAt,omitempty"`
	ModifiedAt      string            `json:"modifiedAt,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	ResourceVersion int               `json:"resourceVersion,omitempty"`
}

type SessionClusterSpec struct {
	State                string                   `json:"state,omitempty"`
	DeploymentTargetName string                   `json:"deploymentTargetName,omitempty"`
	FlinkVersion         string                   `json:"flinkVersion,omitempty"`
	FlinkImageRegistry   string                   `json:"flinkImageRegistry,omitempty"`
	FlinkImageTag        string                   `json:"flinkImageTag,omitempty"`
	FlinkImageRepository string                   `json:"flinkImageRepository,omitempty"`
	FlinkImagePullPolicy string                   `json:"flinkImagePullPolicy,omitempty"`
	NumberOfTaskManagers int                      `json:"numberOfTaskManagers,omitempty"`
	Resources            map[string]*ResourceSpec `json:"resources,omitempty"`
	FlinkConfiguration   map[string]string        `json:"flinkConfiguration,omitempty"`
	Logging              *Logging                 `json:"logging,omitempty"`
}

type ResourceSpec struct {
	Cpu    float64 `json:"cpu,omitempty"`
	Memory string  `json:"memory,omitempty"`
}

type SessionClusterStatus struct {
	State   string                       `json:"state,omitempty"`
	Failure *Failure                     `json:"failure,omitempty"`
	Running *SessionClusterStatusRunning `json:"running,omitempty"`
}

type SessionClusterStatusRunning struct {
	StartedAt          string `json:"startedAt,omitempty"`
	LastUpdateTime     string `json:"lastUpdateTime,omitempty"`
	TaskManagerNumbers int    `json:"taskManagerNumbers,omitempty"`
}

type FlinkImageInfo struct {
	FlinkVersion      string
	RepositoryAddress string
	RepositoryName    string
	PullPolicy        string
	Tag               string
}

var (
	flinkVersionPath = func(tag string) string {
		return fmt.Sprintf("$.flinkImageTagsAndRepository['%s'].flinkVersion", tag)
	}
	repositoryPath = func(tag string) string {
		return fmt.Sprintf("$.flinkImageTagsAndRepository['%s'].image.repository", tag)
	}
	pullPolicyPath = func(tag string) string {
		return fmt.Sprintf("$.flinkImageTagsAndRepository['%s'].image.pullPolicy", tag)
	}
)

func (sc SessionCluster) String() string {
	marshal, _ := json.Marshal(sc)
	return string(marshal)
}

const (
	// not running
	ClusterStopped = "STOPPED"
	// available to user use the session cluster
	ClusterRunning = "RUNNING"
	// create resource,waiting k8s resource finish
	ClusterStarting = "STARTING"
	// try resource and waite update , if suc, it will turn updating. if we have no encourage
	// resource,it will be blocking this status if user change tm nums equals old tm nums,it will be
	// turn running and not update
	ClusterPendingUpdate = "PENDING_UPDATE"
	// update the resource
	ClusterUpdating = "UPDATING"
	// stop the running job、delete the resource、 drop cluster jm and tm pods ...
	ClusterStopping = "STOPPING"
	// STARTING OR UPDATING have exception,status to change FAILED
	ClusterFailed = "FAILED"
)

// GetSessionClusters get session cluster list from Flink AppManager
func (c *Client) GetSessionClusters(namespace string) ([]SessionCluster, int, error) {
	sessionClusterList := &SessionClusterResourceList{}
	code, err := c.get(sessionClusterUrl(namespace), sessionClusterList)
	if err != nil {
		return nil, code, err
	}

	return sessionClusterList.Items, code, nil
}

// GetSessionCluster get session cluster by name
func (c *Client) GetSessionCluster(name, namespace string) (*SessionCluster, int, error) {
	url := fmt.Sprintf("%s/%s", sessionClusterUrl(namespace), name)

	sessionCluster := &SessionCluster{}
	code, err := c.get(url, sessionCluster)
	if err != nil {
		return nil, code, err
	}

	return sessionCluster, code, nil
}

// CreateSessionCluster Create a session cluster
func (c *Client) CreateSessionCluster(sc *SessionCluster, namespace string) (*SessionCluster, int, error) {
	i, err := c.mergeImageInfo(sc)
	if err != nil {
		return nil, i, err
	}

	sc.Metadata.Namespace = namespace

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(sc); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	code, err := c.post(sessionClusterUrl(namespace), body, sc)
	if err != nil {
		return nil, code, err
	}
	return sc, code, nil
}

// CreateOrReplaceSessionCluster Create or replace the session cluster
func (c *Client) CreateOrReplaceSessionCluster(sc *SessionCluster, namespace string) (*SessionCluster, int, error) {
	i, err := c.mergeImageInfo(sc)
	if err != nil {
		return nil, i, err
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(sc); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	u := fmt.Sprintf("%s/%s", sessionClusterUrl(namespace), sc.Metadata.Name)
	code, err := c.put(u, body, sc)
	if err != nil {
		return nil, code, err
	}
	return sc, code, nil
}

// UpdateSessionCluster Update a session cluster
func (c *Client) UpdateSessionCluster(sc *SessionCluster, namespace string) (*SessionCluster, int, error) {
	// 仅当 Flink Image Tag 修改时才进行镜像信息的补充
	if len(sc.Spec.FlinkImageTag) > 0 {
		i, err := c.mergeImageInfo(sc)
		if err != nil {
			return nil, i, err
		}
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(sc); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	u := fmt.Sprintf("%s/%s", sessionClusterUrl(namespace), sc.Metadata.Name)
	code, err := c.patch(u, body, sc)
	if err != nil {
		return nil, code, err
	}
	return sc, code, nil
}

// DeleteSessionCluster Delete a session cluster
func (c *Client) DeleteSessionCluster(name, namespace string) (*SessionCluster, int, error) {
	u := fmt.Sprintf("%s/%s", sessionClusterUrl(namespace), name)

	sc := &SessionCluster{}
	code, err := c.delete(u, sc)
	if err != nil {
		return nil, code, err
	}

	return sc, code, nil
}

// WaitSessionClusterStateChange 等待 Session Cluster 状态扭转完成
func (c *Client) WaitSessionClusterStateChange(name, state, namespace string) (*SessionCluster, int, error) {
	validateState := func() bool {
		switch state {
		case ClusterStopped, ClusterRunning, ClusterStarting, ClusterPendingUpdate,
			ClusterUpdating, ClusterStopping, ClusterFailed:
			return true
		default:
			return false
		}
	}

	getState := func() (interface{}, string, int, error) {
		sc, i, err := c.GetSessionCluster(name, namespace)
		if err != nil {
			return nil, "", i, err
		}
		return sc, sc.Status.State, i, nil
	}

	sc, i, err := c.waitStateChange(state, validateState, getState)
	if err != nil {
		return nil, i, err
	}
	return sc.(*SessionCluster), i, nil
}

// mergeImageInfo session cluster 对象扩充 flink 镜像信息
func (c *Client) mergeImageInfo(sc *SessionCluster) (int, error) {
	fii, i, err := c.getFlinkImageInfo(sc.Spec.FlinkImageTag)
	if err != nil {
		return i, err
	}
	sc.Spec.FlinkImageRegistry = fii.RepositoryAddress
	sc.Spec.FlinkImageRepository = fii.RepositoryName
	sc.Spec.FlinkImagePullPolicy = fii.PullPolicy
	sc.Spec.FlinkVersion = fii.FlinkVersion
	return http.StatusOK, nil
}

// getFlinkImageInfo 从 config 获取 flink 镜像信息
func (c *Client) getFlinkImageInfo(imageTag string) (*FlinkImageInfo, int, error) {
	var config []byte
	i, err := c.get(uiConfigUrl, &config)
	if err != nil {
		return nil, i, err
	}

	flinkVersion, err := jsonslice.Get(config, flinkVersionPath(imageTag))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("get flink version failed: %v", err)
	}
	if len(flinkVersion) == 0 {
		return nil, http.StatusInternalServerError, errors.New("get nil flink version")
	}

	repository, err := jsonslice.Get(config, repositoryPath(imageTag))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("get repository failed: %v", err)
	}

	pullPolicy, err := jsonslice.Get(config, pullPolicyPath(imageTag))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("get pull policy failed: %v", err)
	}
	if len(pullPolicy) == 0 {
		return nil, http.StatusInternalServerError, errors.New("get nil pull policy")
	}

	address, name, found := strings.Cut(string(repository[1:len(repository)-1]), "/")
	if !found {
		return nil, http.StatusInternalServerError, errors.New("can not found / in repository")
	}

	return &FlinkImageInfo{
		FlinkVersion:      string(flinkVersion[1 : len(flinkVersion)-1]),
		RepositoryAddress: address,
		RepositoryName:    name,
		PullPolicy:        string(pullPolicy[1 : len(pullPolicy)-1]),
	}, i, nil
}
