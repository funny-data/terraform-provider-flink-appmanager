package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Savepoint struct {
	Model
	Metadata *SavepointMetadata `json:"metadata,omitempty"`
	Spec     *SavepointSpec     `json:"spec,omitempty"`
	Status   *SavepointStatus   `json:"status,omitempty"`
}

type SavepointMetadata struct {
	ID              string            `json:"id,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	CreatedAt       *time.Time        `json:"createdAt,omitempty"`
	ModifiedAt      *time.Time        `json:"modifiedAt,omitempty"`
	DeploymentID    string            `json:"deploymentID,omitempty"`
	JobID           string            `json:"jobID,omitempty"`
	Origin          string            `json:"origin,omitempty"`
	SavepointType   string            `json:"type,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	ResourceVersion int               `json:"resourceVersion,omitempty"`
}

type SavepointSpec struct {
	SavepointLocation string `json:"savepointLocation,omitempty"`
	FlinkSavepointID  string `json:"flinkSavepointID,omitempty"`
}

type SavepointStatus struct {
	State   string   `json:"state,omitempty"`
	Failure *Failure `json:"failure,omitempty"`
}

const (
	// SavepointStateStart The Savepoint was started, but is not completed yet.
	SavepointStateStart = "STARTED"

	// SavepointStateCompleted The Savepoint was completed successfully and can be restored from.
	SavepointStateCompleted = "COMPLETED"

	// SavepointStateFailed Creation of the Savepoint failed. Details on the cause of failure can be found in the status.failure field.
	SavepointStateFailed = "FAILED"

	// SavepointStatePendingDeletion
	// The Savepoint was marked for deletion.
	// It will automatically be deleted if it meets all prerequisites.
	// It can no longer be restored from.
	SavepointStatePendingDeletion = "PENDING_DELETION"

	// SavepointStateDeleting The Savepoint is currently being deleted. It can no longer be restored from.
	SavepointStateDeleting = "DELETING"

	// SavepointStateFailedDeletion
	// Deletion of the Savepoint failed.
	// Details on the cause of failure can be found in the status.failure field.
	// It can no longer be restored from. Deletion can be retried.
	SavepointStateFailedDeletion = "FAILED_DELETION"
)

const (
	// SavepointOriginUserRequest The Savepoint was requested manually by a user through Ververica Platform.
	SavepointOriginUserRequest = "USER_REQUEST"

	// SavepointOriginSuspend The Savepoint was requested when the corresponding Deployment was suspended.
	SavepointOriginSuspend = "SUSPEND"

	// SavepointOriginCopied
	// The Savepoint is either a copy of another Savepoint resource, or was created manually using an existing savepointLocation.
	// Both Savepoint resources point to the same physical Apache Flink® savepoint.
	SavepointOriginCopied = "COPIED"

	// SavepointOriginRetainedCheckpoint The Savepoint is a retained Apache Flink® checkpoint that was not discarded after the Apache Flink® job was cancelled.
	SavepointOriginRetainedCheckpoint = "RETAINED_CHECKPOINT"
)

const (
	// SavepointTypeUnknown The type of the underlying savepoint or checkpoint is not known.
	SavepointTypeUnknown = "UNKNOWN"

	// SavepointTypeFull The Savepoint resource references a savepoint or a full checkpoint.
	SavepointTypeFull = "FULL"

	// SavepointTypeIncremental The Savepoint resource references an incremental checkpoint.
	SavepointTypeIncremental = "INCREMENTAL"
)

const (
	AnnotationDeploymentSpecVersion = AnnotationPrefix + ".appmanager.controller.deployment.spec.version"
)

func (s Savepoint) String() string {
	marshal, _ := json.Marshal(s)
	return string(marshal)
}

// GetSavepoints List all savepoints. Can be filtered by Deployment ID or Job ID.
func (c *Client) GetSavepoints(deploymentID, jobID, restoreStrategy, namespace string) ([]Savepoint, int, error) {
	query := make(map[string]string)
	if len(deploymentID) > 0 {
		query["deploymentId"] = deploymentID
	}
	if len(jobID) > 0 {
		query["jobID"] = jobID
	}
	if len(restoreStrategy) > 0 {
		query["restoreStrategy"] = restoreStrategy
	}

	url, err := UrlWithQuery(savepointUrl(namespace), query)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	savepointResourceList := &SavepointResourceList{}
	code, err := c.get(url, savepointResourceList)
	if err != nil {
		return nil, code, err
	}
	return savepointResourceList.Items, code, nil
}

// GetSavepoint Get a savepoint by id
func (c *Client) GetSavepoint(savepointID, namespace string) (*Savepoint, int, error) {
	url := fmt.Sprintf("%s/%s", savepointUrl(namespace), savepointID)

	savepoint := &Savepoint{}
	code, err := c.get(url, savepoint)
	if err != nil {
		return nil, code, err
	}
	return savepoint, code, nil
}

// DeleteSavepoint
// Delete a Savepoint and its underlying data, if eligible.
// Alternatively, if the `force` flag is set in the query parameters, the Savepoint will always be deleted,
// though there will be no guarantees for the deletion of the underlying data.
func (c *Client) DeleteSavepoint(savepointID, namespace string, force bool) (int, error) {
	url := fmt.Sprintf("%s/%s?force=%t", savepointUrl(namespace), savepointID, force)

	code, err := c.delete(url, nil)
	if err != nil {
		return code, err
	}
	return code, nil
}

// CreateSavepoint Create a new savepoint
func (c *Client) CreateSavepoint(savepoint *Savepoint, namespace string) (*Savepoint, int, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(savepoint); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	code, err := c.post(savepointUrl(namespace), body, savepoint)
	if err != nil {
		return nil, code, err
	}
	return savepoint, code, nil
}
