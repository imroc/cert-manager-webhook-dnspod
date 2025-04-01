package dnspod

import (
	"encoding/json"

	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	defaultTTL = 600
)

// Config represents the configuration of dnspod resolver
type Config struct {
	TTL          *uint64                  `json:"ttl"`
	RecordLine   string                   `json:"recordLine"`
	SecretIdRef  cmmeta.SecretKeySelector `json:"secretIdRef"`
	SecretKeyRef cmmeta.SecretKeySelector `json:"secretKeyRef"`
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (*Config, error) {
	ttl := uint64(defaultTTL)
	cfg := &Config{TTL: &ttl}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return nil, errors.Wrap(err, "error decoding solver config")
	}
	if err := validateSecretKeySelector(&cfg.SecretIdRef); err != nil {
		return nil, errors.Wrap(err, "invalid secretIdRef")
	}
	if err := validateSecretKeySelector(&cfg.SecretKeyRef); err != nil {
		return nil, errors.Wrap(err, "invalid secretKeyRef")
	}
	return cfg, nil
}

func validateSecretKeySelector(s *cmmeta.SecretKeySelector) error {
	if s.Name == "" {
		return ErrNeedSecretName
	}
	if s.Key == "" {
		return ErrNeedSecretKey
	}
	return nil
}
