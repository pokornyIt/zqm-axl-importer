package main

import (
	"strings"
	"testing"
)

func TestVersionData(t *testing.T) {
	t.Parallel()
	for i, table := range versionTables {
		_, err := VersionData(table.t)
		if len(table.version) > 0 {
			if err != nil {
				t.Errorf("get data error %s", err)
			}
		} else {
			if err == nil {
				t.Errorf("not get error as expect on test line %d", i)
			}
		}
	}
}

func TestVersion_ToString(t *testing.T) {
	t.Parallel()
	for _, table := range versionTables {
		if len(table.version) < 1 {
			continue
		}
		ver, _ := VersionData(table.t)
		if ver == nil {
			t.Errorf("version struct empty vithout error %s", table.version)
		} else if ver.Version != table.version {
			t.Errorf("invalid version [%s/%s]", ver.Version, table.version)
		} else if !strings.HasPrefix(ver.ToString(), "V: [") {
			t.Errorf("invalid to string <%s>", ver.ToString())
		}
	}
}

func TestVersion_ToStringList(t *testing.T) {
	t.Parallel()
	for _, table := range versionTables {
		if len(table.version) < 1 {
			continue
		}
		ver, _ := VersionData(table.t)
		if len(ver.ToStringList()) != 1 {
			t.Errorf("invalid number in array <%d>", len(ver.ToStringList()))
		} else if !strings.HasPrefix(ver.ToStringList()[0], "V: [") {
			t.Errorf("invalid to string <%s>", ver.ToString())
		}
	}
}

func TestVersion_IsValid(t *testing.T) {
	t.Parallel()
	for _, table := range versionTables {
		if len(table.version) < 1 {
			continue
		}
		ver, _ := VersionData(table.t)
		if !ver.IsValid() {
			t.Errorf("invalid identify valid version")
		}
	}
	ver, _ := VersionData(versionSuccess11)
	ver.Version = ""
	if ver.IsValid() {
		t.Errorf("invalid identify invalid version")
	}

}

func TestVersion_GetDbVersion(t *testing.T) {
	t.Parallel()
	for _, table := range versionTables {
		if len(table.version) < 1 {
			continue
		}
		ver, _ := VersionData(table.t)
		if ver.GetDbVersion() != table.dbVersion {
			t.Errorf("invalid DB version [%s/%s]", ver.GetDbVersion(), table.dbVersion)
		}
	}
}

const (
	versionSuccess11 = `<return><componentVersion><version>11.5.1.14900(11)</version></componentVersion></return>`
	versionSuccess10 = `<return><componentVersion><version>10.5.0</version></componentVersion></return>`
	failResponse     = `<faultcode>axis2ns1:Client</faultcode><faultstring>The endpoint reference (EPR) for the Operation not found is https://192.168.111.131:8443/axl/services/AXLAPIService and the WSA Action = CUCM:DB ver=10.0 getCCMVersiona</faultstring><detail />`
)

var versionTables = []struct {
	t         string
	version   string
	dbVersion string
}{
	{versionSuccess11, "11.5.1.14900(11)", "11.0"},
	{versionSuccess10, "10.5.0", "10.0"},
	{failResponse, "", ""},
}
