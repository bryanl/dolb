package service

// ServiceCreateRequest is a request to create a service.
type ServiceCreateRequest struct {
	Name   string `json:"service_name"`
	Domain string `json:"domain"`
	Regex  string `json:"url_regex"`
}

// ServicesResponse is a services response sent to a client.
type ServicesResponse struct {
	Services []ServiceResponse `json:"services"`
}

// ServiceResponse is a service response sent to a client.
type ServiceResponse struct {
	Name      string                 `json:"service_name"`
	Type      string                 `json:"service_type"`
	Config    map[string]interface{} `json:"config"`
	Upstreams []UpstreamResponse     `json:"upstreams"`
}

// UpstreamResponse is an upstream response sent to a client.
type UpstreamResponse struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}
