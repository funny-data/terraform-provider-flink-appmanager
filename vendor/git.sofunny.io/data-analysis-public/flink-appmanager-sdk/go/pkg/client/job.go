package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Job struct {
	Model
	Metadata *JobMetadata
	Spec     *JobSpec
	Status   *JobStatus
}

type JobMetadata struct {
	ID                 string            `json:"id,omitempty"`
	Namespace          string            `json:"namespace,omitempty"`
	CreatedAt          *time.Time        `json:"createdAt,omitempty"`
	ModifiedAt         *time.Time        `json:"modifiedAt,omitempty"`
	DeploymentID       string            `json:"deploymentId,omitempty"`
	DeploymentName     string            `json:"deploymentName,omitempty"`
	SessionClusterName string            `json:"sessionClusterName,omitempty"`
	Annotations        map[string]string `json:"annotations,omitempty"`
	ResourceVersion    int
}

type JobSpec struct {
	SavepointLocation      string            `json:"savepointLocation,omitempty"`
	AllowNonRestoredState  bool              `json:"allowNonRestoredState,omitempty"`
	Parallelism            int               `json:"parallelism,omitempty"`
	NumberOfTaskManagers   int               `json:"numberOfTaskManagers,omitempty"`
	Artifact               *JarArtifact      `json:"artifact,omitempty"`
	Logging                *Logging          `json:"logging,omitempty"`
	FlinkConfiguration     map[string]string `json:"flinkConfiguration,omitempty"`
	UserFlinkConfiguration map[string]string `json:"userFlinkConfiguration,omitempty"`
	Resources              map[string]string `json:"resources,omitempty"`
}

type JobStatus struct {
	State   string            `json:"state,omitempty"`
	Failure *Failure          `json:"failure,omitempty"`
	Started *JobStatusStarted `json:"started,omitempty"`
}

type JobStatusStarted struct {
	StartedAt                *time.Time `json:"startedAt,omitempty"`
	FlinkJobID               string     `json:"flinkJobId,omitempty"`
	LastUpdateTime           *time.Time `json:"lastUpdateTime,omitempty"`
	ObservedFlinkJobRestarts int        `json:"observedFlinkJobRestarts,omitempty"`
	ObservedFlinkJobStatus   string     `json:"observedFlinkJobStatus,omitempty"`
}

// Flink AppManager 上 Job 的状态
const (
	// The Job is in the process of starting. This includes requesting resources from the Deployment Target and submitting the specified job to it.
	JobStarting = "STARTING"
	// The Job is standby
	JobStandby = "STANDBY"
	// The Job is in the process of terminating.
	JobTerminating = "TERMINATING"
	// The Deployment Target or Flink have reported an unrecoverable failure. The Platform may report additional failure details.
	JobFailed = "FAILED"
	// The Job has terminated.
	JobTerminated = "TERMINATED"
	// The Job has terminated and finished successfully, e.g. a finite streaming or batch job.
	JobFinished = "FINISHED"
	// The Job was successfully started.
	JobStarted = "STARTED"
)

// Flink RestAPI Job 状态
const (
	FlinkJobInitializing = "INITIALIZING"
	FlinkJobCreated      = "CREATED"
	FlinkJobRunning      = "RUNNING"
	FlinkJobFailing      = "FAILING"
	FlinkJobFailed       = "FAILED"
	FlinkJobCancelling   = "CANCELLING"
	FlinkJobCancelled    = "CANCELED"
	FlinkJobFinished     = "FINISHED"
	FlinkJobRestarting   = "RESTARTING"
	FlinkJobSuspended    = "SUSPENDED"
	FlinkJobReconciling  = "RECONCILING"
)

func (j Job) String() string {
	marshal, _ := json.Marshal(j)
	return string(marshal)
}

// GetJobByID Get a job by id
func (c *Client) GetJobByID(jobID, namespace string) (*Job, int, error) {
	url := fmt.Sprintf("%s/%s", jobUrl(namespace), jobID)

	job := &Job{}
	code, err := c.get(url, job)

	if err != nil {
		return nil, code, err
	}
	return job, code, nil
}

// GetJobs List all jobs. Can be filtered by DeploymentId
func (c *Client) GetJobs(deploymentID, namespace string) ([]Job, int, error) {
	query := make(map[string]string)
	if len(deploymentID) > 0 {
		query["deploymentId"] = deploymentID
	}

	url, err := UrlWithQuery(jobUrl(namespace), query)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	jobs := &JobResourceList{}
	code, err := c.get(url, jobs)
	if err != nil {
		return nil, code, err
	}
	return jobs.Items, code, nil
}
