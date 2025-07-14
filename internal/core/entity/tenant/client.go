package entity

import "errors"

// Client
// This struct contains information about the client making the request, such as IP address and user agent.
type Client struct {
	IpAddress string `json:"ipAddress"` // IP address of the client
	UserAgent string `json:"userAgent"` // User agent string of the client
}

func (c *Client) IsValid() error {

	if c == nil {
		return errors.New("invalid client: client event cannot be nil")
	}

	if c.IpAddress == "" {
		return errors.New("invalid client: ipAddress is required")
	}

	if c.UserAgent == "" {
		return errors.New("invalid client: userAgent is required")
	}
	return nil
}

func (c *Client) GetClient() Client {
	if c == nil {
		return Client{}
	}
	return *c
}

func (c *Client) SetUserAgent(userAgent string) {
	if c == nil {
		return
	}

	c.UserAgent = userAgent
}

func (c *Client) SetIpAddress(ipAddress string) {
	if c == nil {
		return
	}

	c.IpAddress = ipAddress
}

func (c *Client) GetIpAddress() string {
	if c == nil {
		return ""
	}
	return c.IpAddress
}

func (c *Client) GetUserAgent() string {
	if c == nil {
		return ""
	}
	return c.UserAgent
}
