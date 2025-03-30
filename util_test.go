package main

import (
	"testing"

	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
)

func TestUtil(t *testing.T) {
	resolvedZone := "prod.api.cloud-lotus.com."
	authZone, err := util.FindZoneByFqdn(resolvedZone, util.RecursiveNameservers)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(authZone)
}
