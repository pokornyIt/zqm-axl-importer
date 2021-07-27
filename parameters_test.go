package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestConfig_ProcessLoadFile(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t       string
		success bool
		name    string
	}{
		{"", false, "empty config"},
		{"<xml></xml>", false, "empty no yaml or JSON"},
		{YamlSuccessFile, true, "yaml success config"},
		{JsonSuccessFile, true, "JSON success config"},
	}
	for _, table := range tables {
		c := NewConfig()
		content := table.t
		err := c.ProcessLoadFile([]byte(content))
		if (table.success && err == nil) || (!table.success && err != nil) {
			continue
		}
		if err != nil {
			t.Errorf("config [%s] fail. Error: [%s]", table.name, err)
		} else {
			t.Errorf("config [%s] not failed", table.name)
		}

	}
}

func TestConfigAxl_Validate(t *testing.T) {
	t.Parallel()

	tables := []struct {
		t          ConfigAxl
		success    bool
		errMessage string
		name       string
	}{
		{ConfigAxl{Server: "", User: "", Password: "", IgnoreCertificate: false, AccessGroup: ""},
			false, "AXL server not defined", "Empty all"},
		{ConfigAxl{Server: "_localhost@invalid", User: "", Password: "", IgnoreCertificate: false},
			false, "FQDN or IP", "Invalid server name"},
		{ConfigAxl{Server: "localhost", User: "", Password: "", IgnoreCertificate: false},
			false, "user", "Empty user"},
		{ConfigAxl{Server: "localhost", User: "user", Password: "", IgnoreCertificate: false},
			false, "password", "Empty password"},
		{ConfigAxl{Server: "localhost", User: "user", Password: "pwd", IgnoreCertificate: true, AccessGroup: "access"},
			false, "password", "All ok"},
	}

	for _, table := range tables {
		err := table.t.Validate()
		if err == nil && table.success {
			continue
		}
		if err != nil && !strings.Contains(err.Error(), table.errMessage) {
			t.Errorf("Config log validation for %s fails", table.name)
		}
	}
}

func TestConfigZqm_Validate(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t           ConfigZqm
		dbServer    string
		errContains string
		name        string
	}{
		{ConfigZqm{JtapiUser: []string{""}, DbServer: "localhost", DbUser: "", DbPassword: "", DbPort: DbPort.Default}, "localhost", "JTAPI", "invalid JTAPI"},
		{ConfigZqm{JtapiUser: []string{"aa"}, DbServer: "", DbUser: "", DbPassword: "", DbPort: DbPort.Default}, "localhost", "DB user", "invalid DB user"},
		{ConfigZqm{JtapiUser: []string{"aa"}, DbServer: "", DbUser: "aa", DbPassword: ""}, "localhost", "DB password", "invalid PWD"},
		{ConfigZqm{JtapiUser: []string{"aa"}, DbServer: "zqm", DbUser: "aa", DbPassword: "aa", DbPort: DbPort.Default}, "zqm", "", "all ok"},
		{ConfigZqm{JtapiUser: []string{"aa"}, DbServer: "_p", DbUser: "aa", DbPassword: "aa", DbPort: DbPort.Default}, "_p", "FQDN", "invalid FQDN"},
		{ConfigZqm{JtapiUser: []string{"aa"}, DbServer: "zqm", DbUser: "aa", DbPassword: "aa", DbPort: DbPort.Min - 10}, "zqm", "DB port", fmt.Sprintf("invalid Port %d", DbPort.Min-10)},
		{ConfigZqm{JtapiUser: []string{"aa"}, DbServer: "zqm", DbUser: "aa", DbPassword: "aa", DbPort: DbPort.Max + 10}, "zqm", "DB port", fmt.Sprintf("invalid Port %d", DbPort.Max+10)},
	}
	for _, table := range tables {
		err := table.t.Validate()
		if len(table.errContains) > 0 {
			if err != nil && !strings.Contains(err.Error(), table.errContains) {
				t.Errorf("ZQM config not expect error for [%s] - [%s / %s]", table.name, table.t.DbServer, table.dbServer)
			}
		} else {
			if table.t.DbServer != table.dbServer {
				t.Errorf("DB server not expect for [%s] - [%s / %s]", table.name, table.t.DbServer, table.dbServer)
			}
		}
	}
}

func TestConfigLog_Validate(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t    ConfigLog
		e    ConfigLog
		name string
	}{
		{ConfigLog{Level: "", FileName: "", JSONFormat: false, LogProgramInfo: false, MaxSize: LogMaxSize.Default, MaxBackups: LogMaxBackups.Default, MaxAge: LogMaxAge.Default, Quiet: false},
			ConfigLog{Level: "INFO", FileName: "", JSONFormat: false, LogProgramInfo: false, MaxSize: LogMaxSize.Default, MaxBackups: LogMaxBackups.Default, MaxAge: LogMaxAge.Default, Quiet: false},
			"default values"},
		{ConfigLog{Level: "DEB", FileName: ".\\a\\b.c", JSONFormat: false, LogProgramInfo: false, MaxSize: 10, MaxBackups: 5, MaxAge: 8, Quiet: true},
			ConfigLog{Level: "DEBUG", FileName: "a/b.c", JSONFormat: false, LogProgramInfo: false, MaxSize: 10, MaxBackups: 5, MaxAge: 8, Quiet: true},
			"Debug with log and quiet"},
		{ConfigLog{Level: "ERR", FileName: "/./a/b/", JSONFormat: false, LogProgramInfo: false, MaxSize: 10, MaxBackups: 5, MaxAge: 8, Quiet: true},
			ConfigLog{Level: "ERROR", FileName: "", JSONFormat: false, LogProgramInfo: false, MaxSize: 10, MaxBackups: 5, MaxAge: 8, Quiet: true},
			"logfile name wrong end"},
	}
	for _, table := range tables {
		_ = table.t.Validate()
		if table.t.Level != table.e.Level {
			t.Errorf("Not expect Log level for [%s] - [%s / %s]", table.name, table.t.Level, table.e.Level)
		}
		cmd := strings.ReplaceAll(table.e.FileName, "/", string(os.PathSeparator))
		if table.t.FileName != cmd {
			t.Errorf("Not expect File Name for [%s] - [%s / %s]", table.name, table.t.FileName, cmd)
		}
		if table.t.JSONFormat != table.e.JSONFormat {
			t.Errorf("Not expect JSON format for [%s]", table.name)
		}
		if table.t.LogProgramInfo != table.e.LogProgramInfo {
			t.Errorf("Not expect Log Program info for [%s]", table.name)
		}
		if table.t.MaxBackups != table.e.MaxBackups {
			t.Errorf("Not expect Max backups for [%s]", table.name)
		}
		if table.t.MaxAge != table.e.MaxAge {
			t.Errorf("Not expect Max Age for [%s]", table.name)
		}
		if table.t.MaxSize != table.e.MaxSize {
			t.Errorf("Not expect Max Level for [%s]", table.name)
		}
		if table.t.Quiet != table.e.Quiet {
			t.Errorf("Not expect Quiet for [%s]", table.name)
		}
	}
}

func TestConfigLog_LogToFile(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t       string
		success bool
	}{
		{t: "", success: false},
		{t: "file.log", success: true},
	}
	for _, table := range tables {
		cfg := NewConfig()
		cfg.Log.FileName = FixFileName(table.t)
		if cfg.Log.LogToFile() != table.success {
			t.Errorf("Not expect response for log [%s]=>[%s]", table.t, cfg.Log.FileName)
		}
	}
}

func TestConfigProcessing_Validate(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t          ConfigProcessing
		errContain string
		name       string
	}{
		{t: ConfigProcessing{HoursBack: HoursBack.Default, UserImportHour: []int{UserImportHour.Default}, DefaultTeamName: "team", DefaultRoleName: DefaultRoleName, UpdateInterval: UpdateInterval.Default},
			errContain: "", name: "default values"},
		{t: ConfigProcessing{HoursBack: HoursBack.Max + 10, UserImportHour: []int{UserImportHour.Default}, DefaultTeamName: "team", DefaultRoleName: DefaultRoleName, UpdateInterval: UpdateInterval.Default},
			errContain: "hours back", name: "hours back over"},
		{t: ConfigProcessing{HoursBack: HoursBack.Max, UserImportHour: []int{}, DefaultTeamName: "team", DefaultRoleName: DefaultRoleName, UpdateInterval: UpdateInterval.Default},
			errContain: "one per day", name: "empty User Import"},
		{t: ConfigProcessing{HoursBack: HoursBack.Max, UserImportHour: []int{UserImportHour.Default, UserImportHour.Max + 10}, DefaultTeamName: "team", DefaultRoleName: DefaultRoleName, UpdateInterval: UpdateInterval.Default},
			errContain: "import hour", name: "User Import out of hours"},
		{t: ConfigProcessing{HoursBack: HoursBack.Max, UserImportHour: []int{UserImportHour.Default, UserImportHour.Max, UserImportHour.Max}, DefaultTeamName: "team", DefaultRoleName: DefaultRoleName, UpdateInterval: UpdateInterval.Default},
			errContain: "", name: "User Import duplicate hours"},
		{t: ConfigProcessing{HoursBack: HoursBack.Default, UserImportHour: []int{UserImportHour.Default}, DefaultTeamName: "", DefaultRoleName: DefaultRoleName, UpdateInterval: UpdateInterval.Default},
			errContain: "team name", name: "team not defined"},
		{t: ConfigProcessing{HoursBack: HoursBack.Default, UserImportHour: []int{UserImportHour.Default}, DefaultTeamName: "Team", DefaultRoleName: DefaultRoleName, UpdateInterval: UpdateInterval.Max},
			errContain: "update interval", name: "update interval out of scope"},
	}
	for _, table := range tables {
		err := table.t.Validate()
		if len(table.errContain) > 0 {
			if err != nil && !strings.Contains(err.Error(), table.errContain) {
				t.Errorf("not expected response for [%s]. Error: %s", table.name, err)
			}
		} else {
			if err != nil {
				t.Errorf("not expected response for success test [%s]. Error: %s", table.name, err)
			}
		}
	}
}

func TestConfigProcessing_Validate1(t *testing.T) {
	t.Parallel()
	cfg := ConfigProcessing{HoursBack: HoursBack.Default, UserImportHour: []int{UserImportHour.Default}, DefaultTeamName: "team", DefaultRoleName: DefaultRoleName, UpdateInterval: UpdateInterval.Default, MappingType: MappingBoth}
	tables := []struct {
		set    string
		expect string
	}{
		{"both", MappingBoth},
		{"DEvICE", MappingDevice},
		{"LINE", MappingLine},
		{"LINE+", MappingBoth},
		{"5654645", MappingBoth},
	}
	for _, table := range tables {
		cfg.MappingType = table.set
		err := cfg.Validate()
		if err != nil {
			t.Errorf("validation return not expect error. Error: %s", err)
		}
		if table.expect != cfg.MappingType {
			t.Errorf("not valid mapping for [%s], expect [%s] got [%s]", table.set, table.expect, cfg.MappingType)
		}
	}
}

func TestConfig_Print(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	Suite(t, cfg)
}

func TestConfigLog_Print(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	Suite(t, &cfg.Log)
}

func TestConfigAxl_Print(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	Suite(t, &cfg.Axl)
}

func TestConfigZqm_Print(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	Suite(t, &cfg.Zqm)
}

func TestConfigProcessing_Print(t *testing.T) {
	t.Parallel()
	cfg := NewConfig()
	Suite(t, &cfg.Processing)
}

func TestIntervals_Print(t *testing.T) {
	t.Parallel()
	if len(LogMaxAge.Print()) < 1 {
		t.Errorf("problem pring interval")
	}
}

func TestIntervals_Validate(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t int
		i Intervals
		e bool
	}{
		{1, Intervals{5, 0, 10}, true},
		{0, Intervals{5, 0, 10}, true},
		{10, Intervals{5, 0, 10}, true},
		{-1, Intervals{5, 0, 10}, false},
		{11, Intervals{5, 0, 10}, false},
	}
	for _, table := range tables {
		if table.i.Validate(table.t) != table.e {
			t.Errorf("not validate %d in interval [%d/%d]", table.t, table.i.Min, table.i.Max)
		}
	}
}

func TestIntervals_ValidOrDefault(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t int
		i Intervals
		e int
	}{
		{1, Intervals{5, 0, 10}, 1},
		{0, Intervals{5, 0, 10}, 0},
		{10, Intervals{5, 0, 10}, 10},
		{-1, Intervals{5, 0, 10}, 5},
		{11, Intervals{5, 0, 10}, 5},
	}
	for _, table := range tables {
		if table.i.ValidOrDefault(table.t) != table.e {
			t.Errorf("not expect %d for %d in interval [%d/%d]", table.e, table.t, table.i.Min, table.i.Max)
		}
	}

}

func TestUnique(t *testing.T) {
	t.Parallel()

	tables := []struct {
		t    []int
		e    []int
		name string
	}{
		{[]int{1, 2, 2, 3, 4, 5, 6, 6}, []int{1, 2, 3, 4, 5, 6}, "standard array"},
		{[]int{1, 1, 1, 1, 1, 1, 1, 1}, []int{1}, "only one value"},
		{[]int{8, 1, 8, 1, 1, 8, 1, 8}, []int{1, 8}, "two values"},
		{[]int{}, []int{}, "empty array"},
	}
	for _, tab := range tables {
		uni := unique(tab.t)
		sort.Ints(uni)
		if len(diff(uni, tab.e)) > 0 {
			t.Errorf("Get unique data in array fails for [%s]", tab.name)
		}
	}
}

func TestValidServer(t *testing.T) {
	t.Parallel()

	tables := []struct {
		t       string
		success bool
	}{
		{"localhost", true},
		{"192.168.11.111", true},
		{"127.0.0.1", true},
		{"c09-cucm-a.devlab.zoomint.local", true},
		{"c09-a.internal.global", true},
		{"256.1.1.1", false},
		{"", false},
		{"zd@pd.local", false},
		{"192.168.111.111:8443", false},
	}
	for _, table := range tables {
		if validServer(table.t) != table.success {
			t.Errorf("Problem validate server [%s] - expect is %t", table.t, table.success)
		}
	}
}

func TestIsValidFileName(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t       string
		success bool
	}{
		{"a/b", true},
		{"a//b", true},
		{"", false},
		{"..", false},
		{"/", false},
		{"a/..", false},
		{"a/.", false},
		{"./a/./b/", false},
		{"./a/b/b.", true},
	}

	for _, table := range tables {
		test := IsValidFileName(table.t)
		if test != table.success {
			t.Errorf("For test value [%s] return worng validity", table.t)
		}
	}
}

func TestFixFileName(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t string
		e string
	}{
		{"a\\b\\c", "a/b/c"},
		{"/a\\b\\c", "/a/b/c"},
		{"/a\\b\\c/", ""},
		{"/a\\b\\c/..", ""},
		{"/a\\b\\c/../ab", "/a/b/ab"},
	}
	for _, table := range tables {
		ret := FixFileName(table.t)
		cmp := strings.ReplaceAll(table.e, "/", string(os.PathSeparator))
		if ret != cmp {
			t.Errorf("Invalid fixed path for test %s - [%s <> %s]", table.t, ret, cmp)
		}
	}
}

func Suite(t *testing.T, prn ConfigValid) {
	if len(prn.Print()) < 1 {
		t.Errorf("problem with print function")
	}
}

func diff(X, Y []int) []int {

	var diff []int
	values := map[int]struct{}{}

	for _, x := range X {
		values[x] = struct{}{}
	}

	for _, x := range Y {
		if _, ok := values[x]; !ok {
			diff = append(diff, x)
		}
	}

	return diff
}

const YamlSuccessFile = string(`axl:
  server: c09-cucm-a.devlab.zoomint.com
  user: ccmadmin
  password: zoomadmin
  accessGroup: ZOOM QM Access Group
zqm:
  jtapiUser: 
    - callrec
  dbServer: pm028.pm.zoomint.com
  dbUser: axluser
  dbPassword: a4lUs3r.

log:
  level: TRACE
  fileName:
  jsonFormat: false
  logProgramInfo: true

processing:
  hoursBack: 24
  userImportHour:
    - 4
    - 16
  defaultTeamName: _CUCM_imported
  mappingType: both
  setDirection: true
  coexistCcxImporter: false
`)

const JsonSuccessFile = string(`{
  "axl": {
    "server": "c09-cucm-a.devlab.zoomint.com",
    "user": "ccmadmin",
    "password": "zoomadmin",
	"accessGroup": "ZOOM QM Access Group",    
	"ignoreCertificate": true
  },
  "zqm": {
    "jtapiUser": [
      "callrec",
      "zpokorny"
    ],
    "dbServer": "pm028.pm.zoomint.com",
    "dbPort": 5432,
    "dbUser": "axluser",
    "dbPassword": "a4lUs3r."
  },
  "log": {
    "level": "TRACE",
    "fileName": "",
    "jsonFormat": false,
    "logProgramInfo": true
  },
  "processing": {
    "hoursBack": 24,
    "userImportHour": [
      4,
      16
    ],
    "defaultTeamName": "_CUCM_imported",
    "defaultRoleName": "Agent",
    "updateInterval": 5,
    "mappingType": "both",
    "setDirection": true,
    "coexistCcxImporter ": true
  }
}`)
