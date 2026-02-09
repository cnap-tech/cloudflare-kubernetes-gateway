package controller

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudflare/cloudflare-go/v2"
	"github.com/cloudflare/cloudflare-go/v2/option"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gw "sigs.k8s.io/gateway-api/apis/v1"
)

// ConfigMode determines how tunnel ingress configuration is managed.
type ConfigMode string

const (
	// ConfigModeRemote pushes ingress config to the Cloudflare API.
	// Routes appear in the Cloudflare dashboard. Used for customer-owned tunnels.
	ConfigModeRemote ConfigMode = "remote"

	// ConfigModeLocal writes ingress config to a ConfigMap mounted in cloudflared.
	// Faster updates, no polling delay. Used for CNAP-managed preview domains.
	ConfigModeLocal ConfigMode = "local"
)

// CloudflareConfig holds the API client and configuration extracted from the GatewayClass Secret.
type CloudflareConfig struct {
	AccountID  string
	Client     *cloudflare.Client
	ConfigMode ConfigMode
}

func InitCloudflareApi(ctx context.Context, c client.Client, gatewayClassName string) (*CloudflareConfig, error) {
	log := log.FromContext(ctx)

	gatewayClass := &gw.GatewayClass{}
	if err := c.Get(ctx, types.NamespacedName{Name: gatewayClassName}, gatewayClass); err != nil {
		log.Error(err, "Failed to get gatewayclass")
		return nil, err
	}
	if gatewayClass.Spec.ControllerName != controllerName {
		return nil, nil
	}

	if gatewayClass.Spec.ParametersRef == nil {
		return nil, errors.New("GatewayClass is missing a Secret ParameterRef")
	}

	secretRef := types.NamespacedName{
		Namespace: string(*gatewayClass.Spec.ParametersRef.Namespace),
		Name:      gatewayClass.Spec.ParametersRef.Name,
	}
	secret := &core.Secret{}
	if err := c.Get(ctx, secretRef, secret); err != nil {
		log.Error(err, "unable to fetch Secret from GatewayClass ParameterRef")
		return nil, err
	}

	account := strings.TrimSpace(string(secret.Data["ACCOUNT_ID"]))
	token := strings.TrimSpace(string(secret.Data["TOKEN"]))

	// Build client options
	opts := []option.RequestOption{
		option.WithAPIToken(token),
	}

	// Optional: redirect all API calls through a proxy (e.g. CNAP's API).
	// When set, the controller calls the proxy instead of api.cloudflare.com directly.
	if baseURL := strings.TrimSpace(string(secret.Data["API_BASE_URL"])); baseURL != "" {
		log.Info("Using custom API base URL", "baseURL", baseURL)
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	api := cloudflare.NewClient(opts...)

	// Parse config mode (default: remote)
	configMode := ConfigModeRemote
	if mode := strings.TrimSpace(string(secret.Data["CONFIG_MODE"])); mode != "" {
		switch ConfigMode(strings.ToLower(mode)) {
		case ConfigModeLocal:
			configMode = ConfigModeLocal
		case ConfigModeRemote:
			configMode = ConfigModeRemote
		default:
			log.Info("Unknown CONFIG_MODE value, defaulting to remote", "value", mode)
		}
	}

	return &CloudflareConfig{
		AccountID:  account,
		Client:     api,
		ConfigMode: configMode,
	}, nil
}
