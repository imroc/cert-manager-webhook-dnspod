package dnspod

import (
	"log/slog"
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
)

// Solver implements the logic needed to 'present'an ACME challenge TXT record
// for dnspod.
type Solver struct {
	client   *kubernetes.Clientset
	log      *slog.Logger
	logLevel *slog.LevelVar
}

func NewSolver() *Solver {
	logLevel := &slog.LevelVar{}
	logLevel.Set(slog.LevelInfo)
	return &Solver{
		log:      slog.New(slog.NewJSONHandler(os.Stdin, &slog.HandlerOptions{Level: logLevel})),
		logLevel: logLevel,
	}
}

func (c *Solver) SetLogLevel(level string) error {
	if err := c.logLevel.UnmarshalText([]byte(level)); err != nil {
		return errors.Wrap(err, "failed to parse log level, valid values are: debug, info, warn, error")
	}
	return nil
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
func (c *Solver) Name() string {
	return "dnspod"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *Solver) Present(ch *v1alpha1.ChallengeRequest) error {
	c.log.Debug("Present", "cr", ch)
	cfg, dnspodClient, err := c.getConfigAndClient(ch)
	if err != nil {
		c.Error(err, "failed to get config dnspod client when present", "cr", ch)
		return errors.WithStack(err)
	}

	err = c.createTxtRecord(
		dnspodClient,
		ch.ResolvedZone,
		ch.ResolvedFQDN,
		ch.Key,
		cfg.RecordLine,
		cfg.TTL,
	)
	if err != nil {
		c.Error(err, "failed to create txt record", "cr", ch)
		return errors.WithStack(err)
	}
	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *Solver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	c.log.Debug("CleanUp", "cr", ch)
	_, dnspodClient, err := c.getConfigAndClient(ch)
	if err != nil {
		c.Error(err, "failed to get dnspod client when cleanp up", "cr", ch)
		return errors.WithStack(err)
	}

	if err := c.ensureTxtRecordsDeleted(dnspodClient, ch.ResolvedZone, ch.ResolvedFQDN, ch.Key); err != nil {
		c.Error(err, "failed to ensure txt records deleted", "cr", ch)
		return errors.WithStack(err)
	}
	return nil
}

func (c *Solver) Error(err error, msg string, args ...any) {
	args = append(args, "error", err)
	c.log.Error(msg, args...)
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
func (c *Solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		c.Error(err, "Failed to new kubernetes client")
		return err
	}
	c.client = cl
	return nil
}
