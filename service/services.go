package service

type EndpointRequest struct {
	ServiceName string `json:"service_name"`
	Domain      string `json:"domain"`
	Regex       string `json:"url_regex"`
}

type ServicesResponse struct {
	Services []ServiceResponse `json:"services"`
}

type ServiceResponse struct {
	Name      string                 `json:"service_name"`
	Type      string                 `json:"service_type"`
	Config    map[string]interface{} `json:"config"`
	Upstreams []UpstreamResponse     `json:"upstreams"`
}

type UpstreamResponse struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}
