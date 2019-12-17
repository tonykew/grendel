package tor

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	DELLOS10_RESTCONF_MACTABLE = "/restconf/data/dell-l2-mac:oper-params"
)

type DellOS10 struct {
	endpoint string
	user     string
	password string
	client   *http.Client
}

type dellMacTable struct {
	DynamicCount int                  `json:"dynamic-mac-count`
	StaticCount  int                  `json:"static-mac-count`
	Entries      []*dellMacTableEntry `json:"fwd-table"`
}

type dellMacTableEntry struct {
	PortIndex int    `json:"dot1d-port-index"`
	Type      string `json:"entry-type"`
	Ifname    string `json:"if-name"`
	MAC       string `json:"mac-addr"`
	Status    string `json:"status"`
	VLAN      string `json:"vlan"`
}

type dellRestconfError struct {
	AppTag  string `json:"error-app-tag"`
	Message string `json:"error-message"`
	Tag     string `json:"error-tag"`
	Type    string `json:"type"`
}

func NewDellOS10(endpoint, user, password, cacert string, insecure bool) (*DellOS10, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}}

	pem, err := ioutil.ReadFile(cacert)
	if err == nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("Failed to read cacert: %s", cacert)
		}

		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: certPool, InsecureSkipVerify: false}}
	}

	d := &DellOS10{
		user:     user,
		password: password,
		endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   &http.Client{Timeout: time.Second * 20, Transport: tr},
	}

	return d, nil
}

func (d *DellOS10) URL(resource string) string {
	return fmt.Sprintf("%s%s", d.endpoint, resource)
}

func (d *DellOS10) getRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if d.user != "" && d.password != "" {
		req.SetBasicAuth(d.user, d.password)
	}

	return req, nil
}

func (d *DellOS10) GetMACTable() (MACTable, error) {
	req, err := d.getRequest(d.URL(DELLOS10_RESTCONF_MACTABLE))
	if err != nil {
		return nil, err
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 500 {
		return nil, fmt.Errorf("Failed to fetch mac table with HTTP status code: %d", res.StatusCode)
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	//log.Debugf("DELLOS10 json response: %s", rawJson)

	var dmacTable map[string]*dellMacTable
	err = json.Unmarshal(rawJson, &dmacTable)
	if err != nil {
		return nil, err
	}

	if rec, ok := dmacTable["dell-l2-mac:oper-params"]; ok {
		macTable := make(MACTable, 0)

		for _, entry := range rec.Entries {
			macTable[entry.MAC] = &MACTableEntry{
				Ifname: entry.Ifname,
				Port:   entry.PortIndex,
				VLAN:   entry.VLAN,
				Type:   entry.Type,
			}
		}

		return macTable, nil
	}

	var derr map[string]map[string][]*dellRestconfError
	err = json.Unmarshal(rawJson, &derr)
	if err != nil {
		return nil, err
	}

	if erec, ok := derr["ietf-restconf:errors"]; ok {
		if rec, ok := erec["error"]; ok {
			if len(rec) > 0 {
				return nil, fmt.Errorf("Failed to fetch mac table: %s - %s", rec[0].Tag, rec[0].Message)
			}
		}
	}

	return nil, errors.New("Failed to fetch mac table, unknown error")
}
