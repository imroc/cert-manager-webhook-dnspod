package dnspod

import (
	"strings"

	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/pkg/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

var txtRecordType = "TXT"

func (s *Solver) ensureTxtRecordsDeleted(client *dnspod.Client, zone, fqdn, key string) error {
	recordName := extractRecordName(fqdn, zone)
	req := dnspod.NewDescribeRecordListRequest()
	req.Domain = &zone
	req.RecordType = &txtRecordType
	resp, err := client.DescribeRecordList(req)
	s.log.Debug("dnspod api request", "api", "DescribeRecordList", "request", req, "response", resp)
	if err != nil {
		if isRecordNotFound(err) {
			s.log.Info("TXT record not found, skipping delete", "recordName", recordName, "zone", zone)
			return nil
		}
		return errors.Wrapf(err, "failed to list txt records for zone %s", zone)
	}
	for _, record := range resp.Response.RecordList {
		if *record.Value != key {
			continue
		}
		req := dnspod.NewDeleteRecordRequest()
		req.Domain = &zone
		req.RecordId = record.RecordId
		resp, err := client.DeleteRecord(req)
		s.log.Debug("dnspod api request", "api", "DeleteRecord", "request", req, "response", resp)
		if err != nil {
			return errors.Wrapf(err, "failed to delete txt record (value:%s id:%d) in zone %s", *record.Value, *record.RecordId, zone)
		}
	}
	return nil
}

func extractRecordName(fqdn, domain string) string {
	name := util.UnFqdn(fqdn)
	if idx := strings.LastIndex(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}

func (s *Solver) createTxtRecord(client *dnspod.Client, zone, fqdn, key, recordLine string, ttl *uint64) error {
	if recordLine == "" {
		recordLine = "默认"
	}
	req := dnspod.NewCreateTXTRecordRequest()
	req.Domain = common.StringPtr(zone)
	req.TTL = ttl
	req.Value = &key
	req.RecordLine = &recordLine
	req.SubDomain = common.StringPtr(extractRecordName(fqdn, zone))

	resp, err := client.CreateTXTRecord(req)
	s.log.Debug("dnspod api request", "api", "CreateTXTRecord", "request", req, "response", resp)
	if err != nil {
		s.Error(err, "dnspod api request failed", "api", "CreateTXTRecord", "request", req, "response", resp)
		return errors.WithStack(err)
	}
	return nil
}
