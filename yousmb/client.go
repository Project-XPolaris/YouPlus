package yousmb

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/projectxpolaris/youplus/config"
	"strings"
)

var DefaultClient *Client

func init() {
	DefaultClient = &Client{}
}

type Client struct {
}
type CreateShareOption struct {
	Name       string
	Path       string
	Public     bool
	ValidUsers []string
	WriteList  []string
}

func (c *Client) CreateNewShare(option *CreateShareOption) error {
	properties := map[string]interface{}{
		"path":           option.Path,
		"browseable":     "yes",
		"available":      "yes",
		"directory mask": "0775",
		"create mask":    "0775",
		"writable":       "yes",
		"public":         "yes",
	}
	if !option.Public {
		properties["public"] = "no"
		properties["valid users"] = strings.Join(option.ValidUsers, ",")
		properties["write list"] = strings.Join(option.WriteList, ",")
	}
	requestBody := map[string]interface{}{
		"name":       option.Name,
		"properties": properties,
	}
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/folders/add"))
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) CreateNewShareWithRaw(properties map[string]interface{}) error {
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(properties).
		Post(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/folders/add"))
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) AddUser(username string, password string) error {
	requestBody := map[string]interface{}{
		"username": username,
		"password": password,
	}
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/users"))
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveUser(username string) error {
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("username", username).
		Delete(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/users"))
	if err != nil {
		return err
	}
	return nil
}

type SMBSection struct {
	Name   string            `json:"name"`
	Fields map[string]string `json:"fields"`
}
type SMBConfigResponse struct {
	Sections []SMBSection `json:"sections"`
}

func (c *Client) GetConfig() (*SMBConfigResponse, error) {
	var body SMBConfigResponse
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&body).
		Get(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/config"))
	if err != nil {
		return nil, err
	}
	return &body, err
}

type Info struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *Client) GetInfo() (*Info, error) {
	var body Info
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&body).
		Get(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/info"))
	if err != nil {
		return nil, err
	}
	return &body, err
}

func (c *Client) RemoveFolder(name string) error {
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("name", name).
		Get(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/folders/remove"))
	if err != nil {
		return err
	}
	return err
}

type FolderRequestBody struct {
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties"`
}

func (c *Client) UpdateFolder(body *FolderRequestBody) error {
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(fmt.Sprintf("%s%s", config.Config.YouSMBAddr, "/folders/update"))
	if err != nil {
		return err
	}
	return err
}
