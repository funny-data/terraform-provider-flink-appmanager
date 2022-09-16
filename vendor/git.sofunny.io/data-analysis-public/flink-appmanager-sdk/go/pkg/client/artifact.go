package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type Artifact struct {
	Model
	Metadata *ArtifactMetadata `json:"metadata"`
}

type ArtifactMetadata struct {
	Filename   string `json:"filename,omitempty"`
	Content    string `json:"content,omitempty"`
	Uri        string `json:"uri,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
}

func (a Artifact) String() string {
	marshal, _ := json.Marshal(a)
	return string(marshal)
}

const (
	ArtifactKindJar = "jar"
)

// UploadJar only upload jar file
func (c *Client) UploadJar(filename, namespace string, file io.Reader) (string, int, error) {
	// validate suffix
	hasSuffix := strings.HasSuffix(filename, ArtifactKindJar)
	if !hasSuffix {
		return "", http.StatusInternalServerError, errors.New("upload jar check: not a jar file")
	}

	url := fmt.Sprintf("%s/upload", artifactUrl(namespace))

	return c.upload(filename, file, url)
}

// UploadPropertyFile upload other property file
func (c *Client) UploadPropertyFile(filename, namespace string, file io.Reader) (string, int, error) {
	// NOTE: 目前与 jar 上传到同一目录下，Flink AppManager V2 后会支持上传到其他目录
	url := fmt.Sprintf("%s/upload", artifactUrl(namespace))

	return c.upload(filename, file, url)
}

// upload file via http request
func (c *Client) upload(filename string, file io.Reader, url string) (string, int, error) {
	body := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(body)

	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	if _, err = io.Copy(fileWriter, file); err != nil {
		return "", http.StatusInternalServerError, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	artifact := &Artifact{}
	code, err := c.postFile(url, contentType, body, artifact)
	if err != nil {
		return "", code, err
	}

	return artifact.Metadata.Uri, code, nil
}

// Delete the specific file
func (c *Client) DeleteArtifact(filename, namespace string) (bool, int, error) {

	if filename == "" {
		return false, http.StatusBadRequest, fmt.Errorf("filename cannot be empty")
	}

	url := fmt.Sprintf("%s/delete", artifactUrl(namespace))
	url, err := UrlWithQuery(url, map[string]string{"filename": filename})
	if err != nil {
		return false, http.StatusInternalServerError, err
	}

	var isSuccess bool
	code, err := c.delete(url, &isSuccess)
	if err != nil {
		return false, code, err
	}
	return isSuccess, code, nil
}

// GetArtifacts 获取资源列表
func (c *Client) GetArtifacts(namespace string) ([]Artifact, int, error) {
	url := fmt.Sprintf("%s/list", artifactUrl(namespace))
	artifacts := &ArtifactList{}
	i, err := c.get(url, artifacts)
	if err != nil {
		return nil, i, err
	}
	return artifacts.Items, i, nil
}

// CreateDirectory 创建文件夹
func (c *Client) CreateDirectory(dir, namespace string) (int, error) {
	url := fmt.Sprintf("%s/mkdir", artifactUrl(namespace))
	url, err := UrlWithQuery(url, map[string]string{"path": dir})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	i, err := c.post(url, nil, nil)
	if err != nil {
		return i, err
	}
	return i, nil
}

// GetArtifactMetadata 获取文件的元数据信息
func (c *Client) GetArtifactMetadata(filename, namespace string) (*Artifact, int, error) {
	url := fmt.Sprintf("%s/getMetadata", artifactUrl(namespace))
	url, err := UrlWithQuery(url, map[string]string{"filename": filename})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	artifact := &Artifact{}
	i, err := c.get(url, artifact)
	if err != nil {
		return nil, i, err
	}
	return artifact, i, nil
}

// DownloadFile 下载资源
func (c *Client) DownloadFile(filename, namespace string) ([]byte, int, error) {
	url := fmt.Sprintf("%s/download", artifactUrl(namespace))
	url, err := UrlWithQuery(url, map[string]string{"filename": filename})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var data []byte
	i, err := c.get(url, &data)
	if err != nil {
		return nil, i, err
	}
	return data, i, nil
}
