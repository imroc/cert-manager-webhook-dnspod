package dnspod

import (
	"context"
	"fmt"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/pkg/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Solver) loadSecretData(namespace string, ref cmmeta.SecretKeySelector) (string, error) {
	secret, err := c.client.CoreV1().Secrets(namespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "failed to get secret '%s/%s'", namespace, ref.Name)
	}
	data, ok := secret.Data[ref.Key]
	if !ok {
		return "", fmt.Errorf("no data found for %q in secret '%s/%s'", ref.Key, namespace, ref.Name)
	}
	return string(data), nil
}

func (c *Solver) getConfigAndClient(ch *v1alpha1.ChallengeRequest) (*Config, *dnspod.Client, error) {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load config from challenge request")
	}

	secretId, err := c.loadSecretData(ch.ResourceNamespace, cfg.SecretIdRef)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load secret id from secret")
	}

	secretKey, err := c.loadSecretData(ch.ResourceNamespace, cfg.SecretKeyRef)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load secret key from secret")
	}

	credential := common.NewCredential(secretId, secretKey)
	dnspodClient, err := dnspod.NewClient(credential, "", profile.NewClientProfile())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create dnspod client")
	}
	return cfg, dnspodClient, nil
}
