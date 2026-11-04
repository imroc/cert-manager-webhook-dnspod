package dnspod

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	api "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// dnsPodRecords exercises the actual SDK requests without DNSPod credentials or
// external DNS changes. Each CreateRecord appends a distinct TXT value.
type dnsPodRecords struct {
	mu      sync.Mutex
	records []map[string]any
	nextID  uint64
}

func (d *dnsPodRecords) RoundTrip(req *http.Request) (*http.Response, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var body struct {
		Domain    string
		SubDomain string
		Value     string
		RecordId  uint64
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.Domain != "validation.example.org" {
		return nil, fmt.Errorf("unexpected DNS zone %q", body.Domain)
	}
	result := map[string]any{"RequestId": "test-request"}
	switch req.Header.Get("X-TC-Action") {
	case "CreateRecord":
		d.nextID++
		d.records = append(d.records, map[string]any{"Name": body.SubDomain, "Value": body.Value, "RecordId": d.nextID, "Type": "TXT"})
		result["RecordId"] = d.nextID
	case "DescribeRecordList":
		result["RecordList"] = d.records
	case "DeleteRecord":
		found := false
		for i, record := range d.records {
			if record["RecordId"].(uint64) == body.RecordId {
				d.records = append(d.records[:i], d.records[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("deleted nonexistent record %d", body.RecordId)
		}
	default:
		return nil, fmt.Errorf("unexpected API action %q", req.Header.Get("X-TC-Action"))
	}
	raw, err := json.Marshal(map[string]any{"Response": result})
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
}

func TestSharedCNAMETargetCleanupPreservesOtherChallenges(t *testing.T) {
	records := &dnsPodRecords{records: []map[string]any{}}
	client, err := api.NewClient(common.NewCredential("test-id", "test-key"), "", profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}
	client.WithHttpTransport(records)
	solver := NewSolver()
	const zone = "validation.example.org"
	const target = "_acme-challenge.validation.example.org."
	const otherTarget = "_acme-challenge.other.validation.example.org."
	keys := []string{"public-challenge", "office-challenge", "internal-challenge", "dashboard-challenge"}
	ttl := uint64(600)

	// Challenges for different original domains resolve to the same CNAME target.
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(value string) {
			defer wg.Done()
			if err := solver.createTxtRecord(client, zone, target, value, "", &ttl); err != nil {
				t.Error(err)
			}
		}(key)
	}
	wg.Wait()
	// A matching value at another record name must also survive cleanup.
	if err := solver.createTxtRecord(client, zone, otherTarget, keys[0], "", &ttl); err != nil {
		t.Fatal(err)
	}

	if err := solver.ensureTxtRecordsDeleted(client, zone, target, keys[0]); err != nil {
		t.Fatal(err)
	}
	if err := solver.ensureTxtRecordsDeleted(client, zone, target, keys[0]); err != nil {
		t.Fatalf("repeated cleanup: %v", err)
	}
	got := map[string]bool{}
	for _, record := range records.records {
		got[record["Name"].(string)+"/"+record["Value"].(string)] = true
	}
	want := map[string]bool{
		"_acme-challenge/office-challenge":       true,
		"_acme-challenge/internal-challenge":     true,
		"_acme-challenge/dashboard-challenge":    true,
		"_acme-challenge.other/public-challenge": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cleanup removed another challenge: got %v, want %v", got, want)
	}

	for _, key := range keys[1:] {
		if err := solver.ensureTxtRecordsDeleted(client, zone, target, key); err != nil {
			t.Fatal(err)
		}
	}
	if len(records.records) != 1 || records.records[0]["Name"] != "_acme-challenge.other" {
		t.Fatalf("unrelated record was removed: %v", records.records)
	}
}
