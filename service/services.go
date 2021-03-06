package service

// ServiceCreateRequest is a request to create a service.
type ServiceCreateRequest struct {
	Name   string `json:"service_name"`
	Port   int    `json:"port"`
	Domain string `json:"domain"`
	Regex  string `json:"url_regex"`
}

// ServiceCreateResponse is a response to create a service.
type ServiceCreateResponse struct {
	Name   string `json:"service_name"`
	Port   int    `json:"port"`
	Domain string `json:"domain"`
	Regex  string `json:"url_regex"`
}

// ServicesResponse is a services response sent to a client.
type ServicesResponse struct {
	Services []ServiceResponse `json:"services"`
}

// ServiceResponse is a service response sent to a client.
type ServiceResponse struct {
	Name      string                 `json:"name"`
	Port      int                    `json:"port"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Upstreams []UpstreamResponse     `json:"upstreams"`
}

// UpstreamResponse is an upstream response sent to a client.
type UpstreamResponse struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// UserInfoResponse is a user info response.
type UserInfoResponse struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
}
