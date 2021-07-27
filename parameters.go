package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const (
	DbServer            = "localhost"
	DefaultRoleName     = "Agent"
	MappingDevice       = "device"
	MappingLine         = "line"
	MappingBoth         = "both"
	DefaultMapping      = MappingBoth
	DefaultSetDirection = true
	DefaultCcxImporter  = false
)

type Intervals struct {
	Default int
	Min     int
	Max     int
}

type Config struct {
	Axl        ConfigAxl        `json:"axl" yaml:"axl"`               // AXl server
	Zqm        ConfigZqm        `json:"zqm" yaml:"zqm"`               // ZQM connection
	Log        ConfigLog        `json:"log" yaml:"log"`               // Log configuration
	Processing ConfigProcessing `json:"processing" yaml:"processing"` // processing
}

type ConfigAxl struct {
	Server            string `json:"server" yaml:"server"`                       // FQDN or IP address of AXL server
	User              string `json:"user" yaml:"user"`                           // AXL user
	Password          string `json:"password" yaml:"password"`                   // AXL user password
	AccessGroup       string `json:"accessGroup" yaml:"accessGroup"`             // Name of Access Control Group valid for allow login user to QM
	IgnoreCertificate bool   `json:"ignoreCertificate" yaml:"ignoreCertificate"` // Ignore AXL certificate
}

type ConfigZqm struct {
	JtapiUser  []string `json:"jtapiUser" yaml:"jtapiUser"`   // ZQM JTAPI user name
	DbServer   string   `json:"dbServer" yaml:"dbServer"`     // Database FQDN or IP. Default is localhost
	DbPort     int      `json:"dbPort" yaml:"dbPort"`         // Database TCP port. Default is 5432
	DbUser     string   `json:"dbUser" yaml:"dbUser"`         // Database user
	DbPassword string   `json:"dbPassword" yaml:"dbPassword"` // Database password
	JavaXTerm  string   `json:"javaXTerm" yaml:"javaXTerm"`   // Full path to JAVA-Xterm jar
	JavaFlush  string   `json:"javaFlush" yaml:"javaFlush"`   // Full path command line for terminal
}

type ConfigLog struct {
	Level          string `json:"level" yaml:"level"`                   // Log level FATAL, ERROR, WARNING, INFO, DEBUG, TRACE. Default is INFO
	FileName       string `json:"fileName" yaml:"fileName"`             // Log filename
	JSONFormat     bool   `json:"jsonFormat" yaml:"jsonFormat"`         // enable log in JSON format
	LogProgramInfo bool   `json:"logProgramInfo" yaml:"logProgramInfo"` // enable log program details (line, file name)
	MaxSize        int    `json:"maxSize" yaml:"maxSize"`               // Maximal log file size in MB
	MaxBackups     int    `json:"maxBackups" yaml:"maxBackups"`         // Maximal Number of backups
	MaxAge         int    `json:"maxAge" yaml:"maxAge"`                 // Maximal backup in days
	Quiet          bool   `json:"quiet" yaml:"quiet"`                   // Logging quiet - output only to file or only panic
}

type ConfigProcessing struct {
	HoursBack          int    `json:"hoursBack" yaml:"hoursBack"`                   // How many hours back analyze couples
	UserImportHour     []int  `json:"userImportHour" yaml:"userImportHour"`         // Import hours
	DefaultTeamName    string `json:"defaultTeamName" yaml:"defaultTeamName"`       // Name of SC team for new users
	DefaultRoleName    string `json:"defaultRoleName" yaml:"defaultRoleName"`       // Name of SC user role. Default 'Agent'
	UpdateInterval     int    `json:"updateInterval" yaml:"updateInterval"`         // Delay between couple update in minutes default is 5 minutes
	MappingType        string `json:"mappingType" yaml:"mappingType"`               // Use update couples based on lines, device names or both
	SetDirection       bool   `json:"setDirection" yaml:"setDirection"`             // Update direction in CR when update agents
	CoexistCcxImporter bool   `json:"coexistCcxImporter" yaml:"coexistCcxImporter"` // Is on same system enabled standard SC CCX Importer
}

type ConfigValid interface {
	Validate() (err error)
	Print() string
}

var (
	showConfig     = kingpin.Flag("show", "Show actual configuration and ends").Default("false").Bool()
	configFile     = kingpin.Flag("config", "Configuration file default is \"server.yml\".").PlaceHolder("cfg.yml").Default("server.yml").String()
	runOnce        = kingpin.Flag("cli", "Run only once and ends").Default("false").Bool()
	config         = NewConfig()
	LogMaxSize     = Intervals{Default: 50, Min: 1, Max: 5000}        // Limits and defaults for Log MaxSize
	LogMaxBackups  = Intervals{Default: 5, Min: 0, Max: 100}          // Limits and defaults for Log MaxBackups
	LogMaxAge      = Intervals{Default: 30, Min: 1, Max: 365}         // Limits and defaults for Log MaxAge
	DbPort         = Intervals{Default: 5432, Min: 1025, Max: 65535}  // Limits and defaults for Db port
	UpdateInterval = Intervals{Default: 5, Min: 1, Max: 30 * 24 * 60} // Limits and defaults for Update Agent interval
	HoursBack      = Intervals{Default: 48, Min: 1, Max: 30 * 24}     // Limits and defaults for Update call attach data
	UserImportHour = Intervals{Default: 4, Min: 0, Max: 23}           // Limits for Processing AXL update
)

func NewConfig() *Config {
	return &Config{
		Axl: ConfigAxl{Server: "",
			User:              "",
			Password:          "",
			AccessGroup:       "",
			IgnoreCertificate: false,
		},
		Zqm: ConfigZqm{
			JtapiUser:  []string{},
			DbServer:   DbServer,
			DbPort:     DbPort.Default,
			DbUser:     "",
			DbPassword: "",
			JavaXTerm:  "",
			JavaFlush:  "",
		},
		Log: ConfigLog{
			Level:          "INFO",
			FileName:       "",
			JSONFormat:     false,
			LogProgramInfo: false,
			MaxSize:        LogMaxSize.Default,
			MaxAge:         LogMaxAge.Default,
			MaxBackups:     LogMaxBackups.Default,
			Quiet:          false,
		},
		Processing: ConfigProcessing{
			HoursBack:          HoursBack.Default,
			UserImportHour:     []int{UserImportHour.Default},
			DefaultTeamName:    "_CUCM_imported",
			DefaultRoleName:    DefaultRoleName,
			UpdateInterval:     UpdateInterval.Default,
			MappingType:        DefaultMapping,
			SetDirection:       DefaultSetDirection,
			CoexistCcxImporter: DefaultCcxImporter,
		},
	}
}

func (c *Config) LoadFile(filename string) (err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return c.ProcessLoadFile(content)
}

/*
Process config content
*/
func (c *Config) ProcessLoadFile(content []byte) (err error) {
	err = yaml.UnmarshalStrict(content, c)
	if err != nil {
		err1 := json.Unmarshal(content, c)
		if err1 != nil {
			return err
		}
	}
	return c.Validate()
}

func (c *Config) Validate() (err error) {
	err = c.Axl.Validate()
	if err != nil {
		return err
	}
	err = c.Zqm.Validate()
	if err != nil {
		return err
	}
	err = c.Log.Validate()
	if err != nil {
		return err
	}
	err = c.Processing.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (a *ConfigAxl) Validate() (err error) {
	if len(a.Server) < 1 {
		return errors.New("AXL server not defined")
	}
	if !validServer(a.Server) {
		return errors.New("AXL server not valid FQDN or IP")
	}
	if len(a.User) < 1 {
		return errors.New("AXL user not defined")
	}
	if len(a.Password) < 1 {
		return errors.New("AXL password not defined")
	}
	if len(a.AccessGroup) < 1 {
		return errors.New("AXL AccessControl Group not defined")
	}

	return nil
}

func (a *ConfigZqm) Validate() (err error) {
	if len(a.JtapiUser) < 1 {
		return errors.New("ZQM JTAPI user not defined")
	}
	for _, user := range a.JtapiUser {
		if len(user) < 1 {
			return errors.New("ZQM JTAPI username is empty")
		}
	}
	if len(a.DbServer) < 1 {
		a.DbServer = "localhost"
	} else if !validServer(a.DbServer) {
		return errors.New("ZQM DB server not valid FQDN or IP")
	}
	if len(a.DbUser) < 1 {
		return errors.New("ZQM DB user not defined")
	}
	if len(a.DbPassword) < 1 {
		return errors.New("ZQM DB password not defined")
	}
	if !DbPort.Validate(a.DbPort) {
		return errors.New(fmt.Sprintf("ZQM DB port is out of range (%d-%d)", DbPort.Min, DbPort.Max))
	}
	if len(a.JavaXTerm) > 0 && !FileExists(a.JavaXTerm) {
		return errors.New(fmt.Sprintf("JAVA-XTERM jar file %s not found", a.JavaXTerm))
	}
	if len(a.JavaXTerm) == 0 {
		a.JavaFlush = ""
	}
	if len(a.JavaFlush) > 0 && !FileExists(a.JavaFlush) {
		return errors.New(fmt.Sprintf("Java flush comman file %s not exists", a.JavaFlush))
	}

	return nil
}

func (a *ConfigZqm) IsCleanCache() bool {
	if len(a.JavaXTerm) < 0 {
		return false
	}
	if !FileExists(a.JavaXTerm) {
		return false
	}
	if len(a.JavaFlush) < 1 {
		return false
	}
	if !FileExists(a.JavaFlush) {
		return false
	}
	return true
}

func (a *ConfigLog) Validate() (err error) {
	lvl := validLogLevel(a.Level)
	a.Level = strings.ToUpper(lvl.String())
	a.FileName = FixFileName(a.FileName)
	a.MaxSize = LogMaxSize.ValidOrDefault(a.MaxSize)
	a.MaxAge = LogMaxSize.ValidOrDefault(a.MaxAge)
	a.MaxBackups = LogMaxSize.ValidOrDefault(a.MaxBackups)

	return nil
}

func (a *ConfigProcessing) Validate() (err error) {
	if !HoursBack.Validate(a.HoursBack) {
		return errors.New(fmt.Sprintf("hours back must be between %d and %d hours", HoursBack.Min, HoursBack.Max))
	}
	proc := unique(a.UserImportHour)
	sort.Ints(proc)
	if len(proc) < 1 || len(proc) > 24 {
		return errors.New("import hours must be define minimal one per day or maximal 24 per day")
	}
	for i, hour := range a.UserImportHour {
		if !UserImportHour.Validate(hour) {
			return errors.New(fmt.Sprintf("import hour on position %d not between %d and %d (actual: %d)", i, UserImportHour.Min, UserImportHour.Max, hour))
		}
	}
	if len(a.DefaultTeamName) < 1 {
		return errors.New("default team name not defined")
	}
	if len(a.DefaultRoleName) < 1 {
		a.DefaultRoleName = DefaultRoleName
	}
	UpdateInterval.Max = a.HoursBack * 30
	if !UpdateInterval.Validate(a.UpdateInterval) {
		return errors.New(fmt.Sprintf("update interval not between %d and  %d", UpdateInterval.Max, UpdateInterval.Max))
	}
	if len(a.MappingType) > 0 {
		a.MappingType = strings.ToLower(a.MappingType)
		if !(a.MappingType == MappingBoth || a.MappingType == MappingDevice || a.MappingType == MappingLine) {
			a.MappingType = DefaultMapping
		}
	} else {
		a.MappingType = DefaultMapping
	}

	return nil
}

func (a *ConfigLog) LogToFile() bool {
	return len(a.FileName) > 0
}

func (i *Intervals) Validate(actual int) bool {
	return actual >= i.Min && i.Max >= actual
}

func (i *Intervals) ValidOrDefault(actual int) int {
	if i.Validate(actual) {
		return actual
	}
	return i.Default
}

func (c *Config) Print() string {
	a := fmt.Sprintf("Application %s\r\n", applicationName)
	a = fmt.Sprintf("%s\t- Runtime version         %s\r\n", a, runtime.Version())
	a = fmt.Sprintf("%s\t- CPUs                    %d\r\n", a, runtime.NumCPU())
	a = fmt.Sprintf("%s\t- Architecture            %s\r\n", a, runtime.GOARCH)
	a = fmt.Sprintf("%s\t- Config file             %s\r\n", a, *configFile)
	a = fmt.Sprintf("%s\t- Run once                %t\r\n", a, *runOnce)
	a = fmt.Sprintf("%s%s", a, c.Axl.Print())
	a = fmt.Sprintf("%s%s", a, c.Zqm.Print())
	a = fmt.Sprintf("%s%s", a, c.Processing.Print())
	a = fmt.Sprintf("%s%s", a, c.Log.Print())

	return a
}

func (a *ConfigProcessing) Print() string {
	o := fmt.Sprintf("Processing\r\n")
	o = fmt.Sprintf("%s\t- Hours back              %d\r\n", o, a.HoursBack)
	o = fmt.Sprintf("%s\t- Default team name       %s\r\n", o, a.DefaultTeamName)
	o = fmt.Sprintf("%s\t- User import hours       %s\r\n", o, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a.UserImportHour)), ", "), "[]"))
	o = fmt.Sprintf("%s\t- Couple mapping by       %s\r\n", o, a.MappingType)
	o = fmt.Sprintf("%s\t- Update call direction   %t\r\n", o, a.SetDirection)
	o = fmt.Sprintf("%s\t- Coexist CCX Importer    %t\r\n", o, a.CoexistCcxImporter)
	return o
}

func (a *ConfigAxl) Print() string {
	o := fmt.Sprintf("AXL\r\n")
	o = fmt.Sprintf("%s\t- Server                  %s\r\n", o, a.Server)
	o = fmt.Sprintf("%s\t- User                    %s\r\n", o, a.User)
	o = fmt.Sprintf("%s\t- Access Control Group    %s\r\n", o, a.AccessGroup)
	return o
}

func (a *ConfigZqm) Print() string {
	o := fmt.Sprintf("ZQM\r\n")
	o = fmt.Sprintf("%s\t- JTAPI User              [%s]\r\n", o, strings.Join(a.JtapiUser, ", "))
	o = fmt.Sprintf("%s\t- DB Server               %s:%d\r\n", o, a.DbServer, a.DbPort)
	o = fmt.Sprintf("%s\t- DB User                 %s\r\n", o, a.DbUser)
	o = fmt.Sprintf("%s\t- JAVAX-Xterm             %s\r\n", o, a.JavaXTerm)
	o = fmt.Sprintf("%s\t- Java Flush command      %s\r\n", o, a.JavaFlush)
	return o
}

func (a *ConfigLog) Print() string {
	o := fmt.Sprintf("Logging\r\n")
	o = fmt.Sprintf("%s\t- Level                   %s\r\n", o, a.Level)
	o = fmt.Sprintf("%s\t- Use JSON format         %t\r\n", o, a.JSONFormat)
	if len(a.FileName) > 0 {
		o = fmt.Sprintf("%s\t- Logging file            %s\r\n", o, a.FileName)
		o = fmt.Sprintf("%s\t- Logging program details %t\r\n", o, a.LogProgramInfo)
		o = fmt.Sprintf("%s\t- Maximal file size in MB %d\r\n", o, a.MaxSize)
		o = fmt.Sprintf("%s\t- Number of backups       %d\r\n", o, a.MaxBackups)
		o = fmt.Sprintf("%s\t- Maximal age in days     %d\r\n", o, a.MaxAge)
		o = fmt.Sprintf("%s\t- Backup compress         %t\r\n", o, true)
	} else {
		o = fmt.Sprintf("%s\t- Don't logging to file\r\n", o)
	}
	return o
}

func (i *Intervals) Print() string {
	return fmt.Sprintf("%d (min: %d / max:%d)", i.Default, i.Min, i.Max)
}

func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	var list []int
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func validServer(srv string) bool {
	ipAddress := regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	invalidAddress := regexp.MustCompile(`^((\d+)\.){3}(\d+)$`)
	dnsName := regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	if ipAddress.MatchString(srv) {
		return true
	}
	if invalidAddress.MatchString(srv) {
		return false
	}
	return dnsName.MatchString(srv)
}

func IsValidFileName(file string) bool {
	file = strings.Trim(file, " ")
	if len(file) < 1 {
		return false
	}
	file = strings.ReplaceAll(file, "\\", "/")
	_, f := path.Split(file)
	if f == "" || f == "." || f == ".." {
		return false
	}
	return true
}

func FixFileName(file string) string {
	if !IsValidFileName(file) {
		return ""
	}
	file = strings.ReplaceAll(file, "\\", "/")
	dir, file1 := path.Split(file)
	dir = path.Clean(dir)
	file1 = path.Join(dir, file1)
	return strings.ReplaceAll(file1, "/", string(os.PathSeparator))
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil || info == nil {
		return false
	}
	return !info.IsDir()
}
