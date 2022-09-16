package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const NamespaceURI string = "/namespaces"

type Namespace struct {
	Model
	Metadata *NamespaceMetadata `json:"metadata,omitempty"`
	Status   *NamespaceStatus   `json:"status,omitempty"`
}

type NamespaceMetadata struct {
	Id              string     `json:"id,omitempty"`
	Name            string     `json:"name,omitempty"`
	CreateAt        *time.Time `json:"createdAt,omitempty"`
	ModifiedAt      *time.Time `json:"modifiedAt,omitempty"`
	ResourceVersion int        `json:"resourceVersion,omitempty"`
}

type NamespaceStatus struct {
	State string `json:"state,omitempty"`
}

const (
	// 命名空间初始化，用于准备相关资源信息 用户无法使用
	NamespaceInit = "INIT"
	// 存活，用户可以正常使用
	NamespaceActive = "ACTIVE"
	// 标记为删除，正在删除空间下资源
	NamespaceMarkedForDELETION = "MARKED_FOR_DELETION"
)

func (n Namespace) String() string {
	marshal, _ := json.Marshal(n)
	return string(marshal)
}

// GetNamespaces 获取所有的 namespace
func (c *Client) GetNamespaces() ([]Namespace, int, error) {
	url := fmt.Sprintf("%s/namespaces", baseUrl)
	namespaces := &NamespaceResourceList{}
	i, err := c.get(url, namespaces)
	if err != nil {
		return nil, i, err
	}
	return namespaces.Items, i, nil
}

// GetNamespace 获取名称为 {name} 的 namespace
func (c *Client) GetNamespace(name string) (*Namespace, int, error) {
	namespace := &Namespace{}
	i, err := c.get(namespaceUrl(name), namespace)
	if err != nil {
		return nil, i, err
	}
	return namespace, i, nil
}

// DeleteNamespace 删除名称为 {name} 的 namespace
func (c *Client) DeleteNamespace(name string) (*Namespace, int, error) {
	namespace := &Namespace{}
	i, err := c.delete(namespaceUrl(name), namespace)
	if err != nil {
		return nil, i, err
	}
	return namespace, i, nil
}

// CreateNamespace 创建名称为 {name} 的 namespace
func (c *Client) CreateNamespace(name string) (*Namespace, int, error) {
	namespace := &Namespace{}
	i, err := c.post(namespaceUrl(name), nil, namespace)
	if err != nil {
		return nil, i, err
	}
	return namespace, i, nil
}

// WaitNamespaceStateChange 等待 namespace 状态扭转完成
func (c *Client) WaitNamespaceStateChange(name string, state string) (*Namespace, int, error) {
	validateState := func() bool {
		switch state {
		case NamespaceInit, NamespaceActive, NamespaceMarkedForDELETION:
			return true
		default:
			return false
		}
	}

	getState := func() (interface{}, string, int, error) {
		n, i, err := c.GetNamespace(name)
		if err != nil {
			return nil, "", i, err
		}
		return n, n.Status.State, i, nil
	}

	n, i, err := c.waitStateChange(state, validateState, getState)
	if err != nil {
		return nil, i, err
	}
	return n.(*Namespace), i, nil
}

// WaitNamespaceDeleteCompleted 等待 namespace 删除完成
func (c *Client) WaitNamespaceDeleteCompleted(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Cfg.Timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errors.New("operation timeout")
		case <-time.After(c.Cfg.Interval):
			_, i, _ := c.GetNamespace(name)
			if i == http.StatusNotFound {
				return nil
			}
		}
	}
}

// DeleteNamespaceCompleted 删除 namespace，并等待删除完成
func (c *Client) DeleteNamespaceCompleted(name string) error {
	_, i, err := c.DeleteNamespace(name)
	if i == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	return c.WaitNamespaceDeleteCompleted(name)
}
