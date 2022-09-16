package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	NamespaceUri          = "namespaces"
	DeploymentTargetUri   = "deployment-targets"
	SessionClusterUri     = "sessionclusters"
	ArtifactUri           = "artifacts"
	DeploymentUri         = "deployments"
	JobUri                = "jobs"
	SavepointUri          = "savepoints"
	DeploymentDefaultsUri = "deployment-defaults"
)

const DefaultAPIVersion = "v1"

type Client struct {
	HttpClient *http.Client
	Cfg        Config
}

type Config struct {
	Endpoint string `yaml:"endpoint,omitempty"`
	Version  string `yaml:"version,omitempty"`

	// wait config
	Interval time.Duration `yaml:"interval,omitempty"`
	Timeout  time.Duration `yaml:"timeout,omitempty"`
}

var (
	baseUrl               string
	uiConfigUrl           string
	namespaceUrl          func(string) string
	deploymentTargetUrl   func(string) string
	sessionClusterUrl     func(string) string
	artifactUrl           func(string) string
	deploymentUrl         func(string) string
	jobUrl                func(string) string
	savepointUrl          func(string) string
	deploymentDefaultsUrl func(string) string
)

func SetUp(config Config) *Client {
	if config.Version == "" {
		config.Version = DefaultAPIVersion
	}

	baseUrl = fmt.Sprintf("%s/api/%s", config.Endpoint, config.Version)
	uiConfigUrl = fmt.Sprintf("%s/ui/config.json", config.Endpoint)
	namespaceUrl = func(namespace string) string {
		return fmt.Sprintf("%s/%s/%s", baseUrl, NamespaceUri, namespace)
	}
	deploymentTargetUrl = func(namespace string) string {
		return fmt.Sprintf("%s/%s", namespaceUrl(namespace), DeploymentTargetUri)
	}
	sessionClusterUrl = func(namespace string) string {
		return fmt.Sprintf("%s/%s", namespaceUrl(namespace), SessionClusterUri)
	}
	artifactUrl = func(namespace string) string {
		return fmt.Sprintf("%s/%s", namespaceUrl(namespace), ArtifactUri)
	}
	deploymentUrl = func(namespace string) string {
		return fmt.Sprintf("%s/%s", namespaceUrl(namespace), DeploymentUri)
	}
	jobUrl = func(namespace string) string {
		return fmt.Sprintf("%s/%s", namespaceUrl(namespace), JobUri)
	}
	savepointUrl = func(namespace string) string {
		return fmt.Sprintf("%s/%s", namespaceUrl(namespace), SavepointUri)
	}
	deploymentDefaultsUrl = func(namespace string) string {
		return fmt.Sprintf("%s/%s", namespaceUrl(namespace), DeploymentDefaultsUri)
	}

	return &Client{
		HttpClient: &http.Client{},
		Cfg:        config,
	}
}

func (c *Client) get(url string, v interface{}) (int, error) {
	return c.do(http.MethodGet, url, nil, v)
}

func (c *Client) postFile(url, contentType string, body io.Reader, v interface{}) (int, error) {
	res, err := c.HttpClient.Post(url, contentType, body)
	if err != nil {
		return res.StatusCode, err
	}

	return unmarshalResponse(res, v)
}

func (c *Client) post(url string, body io.Reader, v interface{}) (int, error) {
	return c.do(http.MethodPost, url, body, v)
}

func (c *Client) delete(url string, v interface{}) (int, error) {
	return c.do(http.MethodDelete, url, nil, v)
}

func (c *Client) put(url string, body io.Reader, v interface{}) (int, error) {
	return c.do(http.MethodPut, url, body, v)
}

func (c *Client) patch(url string, body io.Reader, v interface{}) (int, error) {
	return c.do(http.MethodPatch, url, body, v)
}

func (c *Client) do(method, url string, body io.Reader, v interface{}) (int, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return res.StatusCode, err
	}

	if v == nil {
		return res.StatusCode, nil
	}
	return unmarshalResponse(res, v)
}

// unmarshalResponse 解析 response
func unmarshalResponse(res *http.Response, v interface{}) (int, error) {
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode > http.StatusMultipleChoices {
		apiException := &ApiException{}
		if err = json.Unmarshal(resBody, apiException); err != nil {
			return res.StatusCode, err
		}
		return res.StatusCode, errors.New(apiException.ExceptionFormat())
	}

	switch v := v.(type) {
	case *[]byte:
		*v = append(*v, resBody...)
	default:
		if err = json.Unmarshal(resBody, v); err != nil {
			return http.StatusInternalServerError, err
		}
	}
	return res.StatusCode, nil
}

// UrlWithQuery url 添加 query 参数
func UrlWithQuery(u string, query map[string]string) (string, error) {
	url, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	values := url.Query()
	for key, value := range query {
		values.Add(key, value)
	}

	url.RawQuery = values.Encode()
	return url.String(), nil
}

func (c *Client) waitStateChange(state string, validateState func() bool, getState func() (interface{}, string, int, error)) (interface{}, int, error) {
	if !validateState() {
		return nil, http.StatusBadRequest, fmt.Errorf("use a wrong state: %s", state)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Cfg.Timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, http.StatusInternalServerError, ctx.Err()
		case <-time.After(c.Cfg.Interval):
			o, s, i, err := getState()
			if err != nil {
				return nil, i, err
			}

			if s == state {
				return o, http.StatusOK, nil
			}
		}
	}
}
