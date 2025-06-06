package dnspod

import (
	"log/slog"
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
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
		log:      slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})),
		logLevel: logLevel,
	}
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
func (s *Solver) Name() string {
	return "dnspod"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (s *Solver) Present(ch *v1alpha1.ChallengeRequest) error {
	s.log.Info("present ACME DNS01 challenge", "challenge", ch)
	cfg, dnspodClient, err := s.getConfigAndClient(ch)
	if err != nil {
		s.Error(err, "failed to get config dnspod client when present", "challenge", ch)
		return errors.WithStack(err)
	}

	err = s.createTxtRecord(
		dnspodClient,
		util.UnFqdn(ch.ResolvedZone),
		ch.ResolvedFQDN,
		ch.Key,
		cfg.RecordLine,
		cfg.TTL,
	)
	if err != nil {
		s.Error(err, "failed to create txt record", "challenge", ch)
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
func (s *Solver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	s.log.Info("clean up relevant TXT record of ACME DNS01 challenge", "challenge", ch)
	_, dnspodClient, err := s.getConfigAndClient(ch)
	if err != nil {
		s.Error(err, "failed to get dnspod client when cleanp up", "challenge", ch)
		return errors.WithStack(err)
	}

	if err := s.ensureTxtRecordsDeleted(dnspodClient, util.UnFqdn(ch.ResolvedZone), ch.ResolvedFQDN, ch.Key); err != nil {
		s.Error(err, "failed to ensure txt records deleted", "challenge", ch)
		return errors.WithStack(err)
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
func (s *Solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	client, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		s.Error(err, "Failed to create kubernetes client")
		return errors.WithStack(err)
	}
	s.client = client
	return nil
}
