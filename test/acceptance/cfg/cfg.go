package cfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// Example config:
//  {
//    "kubeconfig_path": "/Users/user/.kube/config",
//    "api": "api.example.com",
//    "admin_user": "admin",
//    "admin_password": "PASSWORD"
//    "apps_domain": "apps.example.com"
//  }
type Config struct {
	KubeConfigPath     string `json:"kubeconfig_path"`
	KeepClusterChanges bool   `json:"keep_cluster_changes"`
	KeepCFChanges      bool   `json:"keep_cf_changes"`

	API           string `json:"api"`
	AdminUser     string `json:"admin_user"`
	AdminPassword string `json:"admin_password"`
	AppsDomain    string `json:"apps_domain"`

	ExistingUser         string `json:"existing_user"`
	ExistingUserPassword string `json:"existing_user_password"`
	ShouldKeepUser       bool   `json:"keep_user_at_suite_end"`
	UseExistingUser      bool   `json:"use_existing_user"`

	UseExistingOrganization bool   `json:"use_existing_organization"`
	ExistingOrganization    string `json:"existing_organization"`
}

func (c *Config) GetAdminUser() string {
	return c.AdminUser
}

func (c *Config) GetAdminPassword() string {
	return c.AdminPassword
}

func (c *Config) GetUseExistingOrganization() bool {
	return c.UseExistingOrganization
}

func (c *Config) GetUseExistingSpace() bool {
	return false
}

func (c *Config) GetExistingOrganization() string {
	return c.ExistingOrganization
}

func (c *Config) GetExistingSpace() string {
	panic("implement me")
}

func (c *Config) GetUseExistingUser() bool {
	return c.UseExistingUser
}

func (c *Config) GetExistingUser() string {
	return c.ExistingUser
}

func (c *Config) GetExistingUserPassword() string {
	return c.ExistingUserPassword
}

func (c *Config) GetShouldKeepUser() bool {
	return c.ShouldKeepUser
}

func (c *Config) GetConfigurableTestPassword() string {
	return ""
}

func (c *Config) GetAdminClient() string {
	return ""
}

func (c *Config) GetAdminClientSecret() string {
	return ""
}

func (c *Config) GetExistingClient() string {
	return ""
}

func (c *Config) GetExistingClientSecret() string {
	panic("implement me")
}

func (c *Config) GetApiEndpoint() string {
	return c.API
}

func (c *Config) GetSkipSSLValidation() bool {
	return true
}

func (c *Config) GetNamePrefix() string {
	return "ACCEPTANCE"
}

func (c *Config) GetScaledTimeout(timeout time.Duration) time.Duration {
	return time.Duration(float64(timeout) * 2)
}

func (c *Config) CfPushTimeoutDuration() time.Duration {
	return 240 * time.Second
}

func (c *Config) CfPushDockerTimeoutDuration() time.Duration {
	return 120 * time.Second
}

func (c *Config) DefaultTimeoutDuration() time.Duration {
	return 60 * time.Second
}

func NewConfig(configPath string, kubeConfigPath string, keepClusterChanges bool, keepCFChanges bool) (*Config, error) {
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config %v", err)
	}

	config := &Config{}
	err = json.Unmarshal([]byte(configFile), config)

	if err != nil {
		return nil, fmt.Errorf("error parsing json %v", err)
	}

	config.KubeConfigPath = kubeConfigPath
	config.KeepClusterChanges = keepClusterChanges
	config.KeepCFChanges = keepCFChanges

	return config, nil
}
