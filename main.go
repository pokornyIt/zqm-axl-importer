package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"
)

const (
	applicationName = "ZQM AXL Importer 2.1.0"                               // application name
	letterBytes     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // map for random string
	letterIdxBits   = 6                                                      // 6 bits to represent a letter index
	letterIdxMask   = 1<<letterIdxBits - 1                                   // All 1-bits, as many as letterIdxBits
	letterIdxMax    = 63 / letterIdxBits                                     // # of letter indices fitting in 63 bits
	TimeFormat      = "15:04:05.0000"                                        // time format
	DateTimeFormat  = "2006-01-02 15:04:05.000"                              // Full date time format
)

var (
	src            = rand.NewSource(time.Now().UnixNano()) // randomize base string
	maxRandomSize  = 10                                    // required size of random string
	shortBodyChars = 120                                   // Max length print from string
)

func RandomString() string {
	sb := strings.Builder{}
	sb.Grow(maxRandomSize)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := maxRandomSize-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func ShortBody(body string) string {
	if strings.Index(body, "\n") > -1 {
		body = body[strings.Index(body, "\n")+1:]
	}
	if len(body) < shortBodyChars {
		return body
	}
	body = body[:shortBodyChars-2] + " ..."
	return body
}

func processDeviceOnSql(deviceIdList []UserDeviceLine) int {
	if len(deviceIdList) < 1 {
		log.WithField("error", "list data for processing is empty").Error("not valid list of users read from AXl server")
		return 1
	}
	conn, err := connectDb()
	if err != nil {
		log.WithField("error", err.Error()).Errorf("problem connect to DB. %s", err.Error())
		return 2
	} else {
		defer func() {
		_:
			conn.Close(context.Background())
		}()
		err = connectRunUserDeviceFunc(conn, deviceIdList)
		if err != nil {
			log.WithField("error", err.Error()).Error("can't update AXL source DB table")
			return 3
		}
		log.WithField("rows", len(deviceIdList)).Infof("now update prepare %d rows", len(deviceIdList))
		err = connectUpdateQm(conn)
	}
	return 0
}

func processLoginUserOnSql(users []LoginUser) int {
	if len(users) < 1 {
		log.WithField("error", "list data for processing is empty").Error("not valid list of users read from AXl server")
		return 1
	}
	conn, err := connectDb()
	if err != nil {
		log.WithField("error", err.Error()).Errorf("problem connect to DB. %s", err.Error())
		return 2
	} else {
		defer func() {
		_:
			conn.Close(context.Background())
		}()
		err = connectRunLoginUserFunc(conn, users)
		if err != nil {
			log.WithField("error", err.Error()).Error("can't update AXL source DB table for login users")
			return 3
		}
		log.WithField("rows", len(users)).Infof("now update prepare %d rows", len(users))
	}
	return 0
}

func IsTimeToAxlUpdate(now time.Time) bool {
	current := now.Hour()
	for _, hour := range config.Processing.UserImportHour {
		if current == hour {
			return true
		}
	}
	return false
}

func processAxlUpdate() {
	log.WithField("process", "AXL Update").Trace("start process AXL update")
	axlConnection := NewConnection(config.Axl.Server, config.Axl.User, config.Axl.Password)
	accessible, _ := axlConnection.IsLoginValid()
	if accessible {
		db, err := axlConnection.DbVersion()
		if err == nil && db != DbVersionError {
			needClearCache := false
			loginUser := axlConnection.GetLoginUserList()
			if loginUser != nil && len(loginUser.Rows) > 0 {
				log.WithFields(log.Fields{"validRows": len(loginUser.Rows)}).Infof("From source AXL table prepare %d valid login user rows", len(loginUser.Rows))
				i := processLoginUserOnSql(loginUser.Rows)
				needClearCache = i == 0
			}
			deviceIdList := axlConnection.GetUserDeviceLineList()
			if deviceIdList != nil {
				newList := deviceIdList.cleanDeviceLineList()
				log.WithFields(log.Fields{"validRows": len(newList)}).Infof("From source AXL table prepare %d valid user/device/line rows", len(newList))
				i := processDeviceOnSql(newList)
				needClearCache = needClearCache || i == 0
			}
			if needClearCache {
				refreshCache()
			}
		} else {
			log.Errorf("problem with AXL connection or DB version not supported")
		}
	}
	log.WithField("process", "AXL Update").Trace("end process AXL update")
}

func refreshCache() {
	if !config.Zqm.IsCleanCache() {
		log.WithField("process", "Clear cache").Trace("not clean cache configured")
		return
	}
	log.WithField("process", "Clear cache").Trace("start run clean tomcat cache")
	cmd := exec.Command("java", "-jar", config.Zqm.JavaXTerm, "java", "--url", "localhost:8765", "-i", config.Zqm.JavaFlush)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{"process": "Clear cache", "error": err.Error()}).Errorf("cache clean command ends with error %s", err)
		return
	}
	if out != nil {
		log.WithFields(log.Fields{"process": "Clear cache", "output": fmt.Sprintf(string(out))}).Debugf("return for command")
	}
	log.WithFields(log.Fields{"process": "Clear cache"}).Info("success clear cache")
}

func processCallsUpdate() int {
	conn, err := connectDb()
	if err != nil {
		log.Errorf("problem connect to DB. %s", err.Error())
		return 2
	} else {
		log.WithField("update at", time.Now().Format(TimeFormat)).Infof("now update call data")
		defer func() {
		_:
			conn.Close(context.Background())
		}()
		if config.Processing.MappingType == MappingBoth || config.Processing.MappingType == MappingDevice {
			err = connectUpdateCalls(conn, processCallUpdateByDevice)
			if err != nil {
				return 1
			}
		}
		if config.Processing.MappingType == MappingBoth || config.Processing.MappingType == MappingLine {
			err = connectUpdateCalls(conn, processCallUpdateByLine)
			if err != nil {
				return 2
			}
		}
	}
	return 0
}

func scheduleCallsUpdate(done chan bool, wg *sync.WaitGroup) {
	tick := time.NewTicker(time.Minute * time.Duration(config.Processing.UpdateInterval))
	defer wg.Done()
	defer tick.Stop()
	processCallsUpdate()
	for {
		select {
		case <-tick.C:
			log.Trace("process call updates")
			processCallsUpdate()
		case <-done:
			log.Debug("call update routine shutdown")
			return
		}
	}
}

func scheduleAxlUpdate(done chan bool, wg *sync.WaitGroup) {
	tick := time.NewTicker(time.Hour)
	defer wg.Done()
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			log.Trace("tick for AXL update")
			if IsTimeToAxlUpdate(time.Now()) {
				log.Trace("update AXL")
				processAxlUpdate()
			}
		case <-done:
			log.Debug("AXL update routine shutdown")
			return
		}
	}
}

func serviceLoop() {
	doneUpdate := make(chan bool, 1)
	doneAxl := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	var wg sync.WaitGroup
	signal.Notify(quit, os.Interrupt)

	log.Infof("start scheduled routines")
	go scheduleCallsUpdate(doneUpdate, &wg)
	go scheduleAxlUpdate(doneAxl, &wg)
	wg.Add(2)

	s := <-quit
	doneAxl <- true
	doneUpdate <- true
	log.Infof("stop request signal is [%s]", s)
	wg.Wait()
}

func main() {
	timeStart := time.Now()
	exitCode := 0

	kingpin.Version(VersionDetail())
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	err := config.LoadFile(*configFile)
	if err != nil && !*showConfig {
		fmt.Printf("Problem read config file [%s]. Error: %s\r\n", *configFile, err)
		os.Exit(1)
	}
	initLog()
	if *showConfig {
		fmt.Println(config.Print())
		log.WithFields(log.Fields{"ApplicationName": applicationName}).Info("show only configuration and exit")
		if err != nil {
			fmt.Printf("Problem in config file [%s]. Error: %s\r\n", *configFile, err)
		}
		os.Exit(0)
	}
	if *runOnce {
		processAxlUpdate()
		processCallsUpdate()
	} else {
		serviceLoop()
	}
	timeEnd := time.Now()
	log.WithFields(log.Fields{"duration": timeEnd.Sub(timeStart).String()}).Infof("Program end at %s", time.Now().Format(TimeFormat))
	time.Sleep(time.Second)
	os.Exit(exitCode)
}
