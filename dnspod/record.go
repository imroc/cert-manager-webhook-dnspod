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
	s.log.Debug(
		"dnspod api request",
		"api", "DescribeRecordList",
		"request", req,
		"response", resp,
	)
	if err != nil {
		if isRecordNotFound(err) {
			s.log.Warn(
				"TXT record not found, skipping deletion",
				"recordName", recordName,
				"zone", zone,
				"request", req,
				"response", resp,
				"error", err,
			)
			return nil
		}
		s.log.Warn(
			"failed to list txt records",
			"recordName", recordName,
			"zone", zone,
			"request", req,
			"response", resp,
			"error", err,
		)
		return errors.WithStack(err)
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
			s.log.Error(
				"failed to delete TXT record",
				"recordValue", *record.Value,
				"recordId", *record.RecordId,
				"zone", zone,
				"request", req,
				"response", resp,
				"error", err,
			)
			return errors.WithStack(err)
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
	req := dnspod.NewCreateRecordRequest()
	req.Domain = common.StringPtr(zone)
	req.TTL = ttl
	req.Value = &key
	req.RecordType = &txtRecordType
	req.RecordLine = &recordLine
	req.SubDomain = common.StringPtr(extractRecordName(fqdn, zone))

	resp, err := client.CreateRecord(req)
	s.log.Debug("dnspod api request", "api", "CreateTXTRecord", "request", req, "response", resp)
	if err != nil {
		s.Error(
			err, "dnspod api request failed",
			"request", req,
			"response", resp,
		)
		return errors.WithStack(err)
	}
	return nil
}
