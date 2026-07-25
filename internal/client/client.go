package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/richo542/sneak/internal/config"
)

type ProviderClient interface {
	TestConnection() error
}

func NewProviderClient(p *config.Provider) (ProviderClient, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	switch p.Type {
	case "jira":
		return &JiraProviderClient{
			cfg:    p,
			client: &client,
		}, nil
	case "azure":
		return &AzureProviderClient{
			cfg:    p,
			client: &client,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", p.Type)
	}
}
