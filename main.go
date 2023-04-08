package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jetstack/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"k8s.io/client-go/kubernetes"
	"os"
	"strings"

	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

const (
	defaultTTL = 600
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName,
		&customDNSProviderSolver{},
	)
}

// customDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type customDNSProviderSolver struct {
	client *kubernetes.Clientset
	dnspod map[string]*dnspod.Client
	// If a Kubernetes 'clientset' is needed, you must:
	// 1. uncomment the additional `client` field in this structure below
	// 2. uncomment the "k8s.io/client-go/kubernetes" import at the top of the file
	// 3. uncomment the relevant code in the Initialize method below
	// 4. ensure your webhook's service account has the required RBAC role
	//    assigned to it for interacting with the Kubernetes APIs you need.
	// client kubernetes.Clientset
}

// customDNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.
type customDNSProviderConfig struct {
	// Change the two fields below according to the format of the configuration
	// to be decoded.
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.

	// Email           string `json:"email"`
	TTL          *uint64                  `json:"ttl"`
	SecretId     string                   `json:"secretId"`
	SecretKeyRef cmmeta.SecretKeySelector `json:"secretKeyRef"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
// For example, `cloudflare` may be used as the name of a solver.
func (c *customDNSProviderSolver) Name() string {
	return "dnspod"
}

func (c *customDNSProviderSolver) getClient(ch *v1alpha1.ChallengeRequest, cfg customDNSProviderConfig) (*dnspod.Client, error) {
	dnspodClient, ok := c.dnspod[cfg.SecretId]
	if !ok {
		if cfg.SecretId == "" {
			return nil, errors.New("no secret id found in config")
		}

		ref := cfg.SecretKeyRef

		secret, err := c.client.CoreV1().Secrets(ch.ResourceNamespace).Get(context.TODO(), ref.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		secretKey, ok := secret.Data[ref.Key]
		if !ok {
			return nil, fmt.Errorf("no secret key for %q in secret '%s/%s'", ref.Name, ref.Key, ch.ResourceNamespace)
		}

		credential := common.NewCredential(cfg.SecretId, string(secretKey))
		dnspodClient, err = dnspod.NewClient(credential, "", profile.NewClientProfile())
		if err != nil {
			return nil, fmt.Errorf("failed to create dnspod client: %s", err.Error())
		}
		klog.Infof("create dnspod client successfully")
		c.dnspod[cfg.SecretId] = dnspodClient
	}
	return dnspodClient, nil
}

func getDomainAndID(client *dnspod.Client, zone string) (*string, *uint64, error) {
	resp, err := client.DescribeDomainList(nil)
	if err != nil {
		return nil, nil, err
	}

	authZone, err := util.FindZoneByFqdn(zone, util.RecursiveNameservers)
	if err != nil {
		return nil, nil, err
	}

	var hostedDomain *dnspod.DomainListItem
	for _, domain := range resp.Response.DomainList {
		if *domain.Name == util.UnFqdn(authZone) {
			hostedDomain = domain
			break
		}
	}
	if hostedDomain == nil {
		return nil, nil, fmt.Errorf("no domain found in zone %s", zone)
	}
	hostedDomainID := *hostedDomain.DomainId

	if hostedDomainID == 0 {
		return nil, nil, fmt.Errorf("Zone %s not found in dnspod for zone %s", authZone, zone)
	}

	return hostedDomain.Name, hostedDomain.DomainId, nil
}

func findTxtRecords(client *dnspod.Client, domainID *uint64, zone, fqdn string) ([]*dnspod.RecordListItem, error) {
	recordName := extractRecordName(fqdn, zone)
	req := dnspod.NewDescribeRecordListRequest()
	req.DomainId = domainID
	req.Domain = &zone
	req.Subdomain = &recordName
	recordType := "TXT"
	req.RecordType = &recordType
	resp, err := client.DescribeRecordList(req)
	if err != nil {
		if e, ok := err.(*terrors.TencentCloudSDKError); ok {
			if e.Code == "ResourceNotFound.NoDataOfRecord" {
				klog.Infof("Ignore TXT record not found %s.%s", recordName, zone)
				return nil, nil
			}
		}
		klog.Errorf("Failed to list records (%d, %s): %v", *domainID, recordName, err)
		return nil, fmt.Errorf("dnspod API call has failed: %v", err)
	}
	return resp.Response.RecordList, nil
}

func extractRecordName(fqdn, zone string) string {
	if idx := strings.Index(fqdn, "."+zone); idx != -1 {
		return fqdn[:idx]
	}

	return util.UnFqdn(fqdn)
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *customDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		klog.Errorf("Failed to log config %v: %v", ch.Config, err)
		return err
	}

	dnspodClient, err := c.getClient(ch, cfg)
	if err != nil {
		klog.Errorf("Failed to get dnspod client %v: %v", cfg, err)
		return err
	}

	domain, domainID, err := getDomainAndID(dnspodClient, ch.ResolvedZone)
	if err != nil {
		klog.Errorf("Failed to get domain id %s: %v", ch.ResolvedZone, err)
		return err
	}

	req := dnspod.NewCreateRecordRequest()
	req.Domain = domain
	req.DomainId = domainID
	req.Domain = &ch.ResolvedZone
	recordType := "TXT"
	name := extractRecordName(ch.ResolvedFQDN, ch.ResolvedZone)
	req.TTL = cfg.TTL
	req.RecordType = &recordType
	req.Value = &ch.Key
	line := "默认"
	req.RecordLine = &line
	req.SubDomain = &name

	klog.Infof("create record: %+#v", *req)
	_, err = dnspodClient.CreateRecord(req)
	if err != nil {
		klog.Errorf("Failed to create record: %v", err)
		return fmt.Errorf("dnspod API call failed: %v", err)
	}

	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *customDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		klog.Errorf("Failed to log config %v: %v", ch.Config, err)
		return err
	}

	dnspodClient, err := c.getClient(ch, cfg)
	if err != nil {
		klog.Errorf("Failed to get dnspod client %v: %v", cfg, err)
		return err
	}

	domain, domainID, err := getDomainAndID(dnspodClient, ch.ResolvedZone)
	if err != nil {
		klog.Errorf("Failed to get domain id %s: %v", ch.ResolvedZone, err)
		return err
	}

	records, err := findTxtRecords(dnspodClient, domainID, ch.ResolvedZone, ch.ResolvedFQDN)
	if err != nil {
		klog.Errorf("Failed to find txt records (%s, %s, %s): %v", domainID, ch.ResolvedZone, ch.ResolvedFQDN, err)
		return err
	}

	for _, record := range records {
		if *record.Value != ch.Key {
			continue
		}
		req := dnspod.NewDeleteRecordRequest()
		req.Domain = domain
		req.DomainId = domainID
		req.RecordId = record.RecordId
		_, err = dnspodClient.DeleteRecord(req)
		if err != nil {
			klog.Errorf("Failed to delete record (%d, %d): %v", *domainID, *record.RecordId, err)
			return err
		}
	}

	return nil
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *customDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		klog.Errorf("Failed to new kubernetes client: %v", err)
		return err
	}
	c.client = cl

	c.dnspod = make(map[string]*dnspod.Client)

	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (customDNSProviderConfig, error) {
	ttl := uint64(defaultTTL)
	cfg := customDNSProviderConfig{TTL: &ttl}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
