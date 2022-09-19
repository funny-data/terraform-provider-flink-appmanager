package client

import "encoding/json"

type SystemInformation struct {
	Model
	Status *SystemInformationStatus `json:"status,omitempty"`
}

type SystemInformationStatus struct {
	JvmVersion     string `json:"jvmVersion,omitempty"`
	CommitShaLong  string `json:"commitShaLong,omitempty"`
	CommitShaShort string `json:"commitShaShort,omitempty"`
	BuildVersion   string `json:"buildVersion,omitempty"`
	BuildTime      string `json:"buildTime,omitempty"`
}

func (si SystemInformation) String() string {
	marshal, _ := json.Marshal(si)
	return string(marshal)
}

const SystemInfoUri = "ui/appmanager/status/system-info"

func (c *Client) GetSystemInfo() (*SystemInformation, int, error) {
	si := &SystemInformation{}
	i, err := c.get(systemInfoUrl, si)
	if err != nil {
		return nil, i, err
	}
	return si, i, nil
}
