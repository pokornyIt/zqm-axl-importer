package main

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const DbVersionError = "Error"

var dbVersionSupport = [...]string{"10.0", "12.0", "14.0", "16.0"}

type Connection struct {
	id          string
	server      string
	user        string
	pwd         string
	dbVersion   string
	timeOut     int
	sequence    int
	isAuthValid bool
	client      *http.Client
}

type AxlRowsData struct {
	XMLName xml.Name      `xml:"return"`
	Rows    []interface{} `xml:"row"`
}

type IApi interface {
	DbVersion() (string, error)
	IsLoginValid() bool
	ToParamString() string
	SetClient(client *http.Client)
}

type IRow interface {
	ToString() string
	ToStringList() []string
}

type ITable interface {
	IRow
	IsValid() bool
}

type IApiData interface {
	ITable
	Status() string
}

func NewConnection(server string, user string, pwd string, time ...int) *Connection {
	a := RandomString()
	log.WithFields(log.Fields{"id": a, "server": server, "user": user, "pwd": pwd}).Tracef("create connection")
	timeOut := 30
	if len(time) > 0 {
		timeOut = time[0]
	}
	con := Connection{server: server, user: user, pwd: pwd, dbVersion: "", timeOut: timeOut, sequence: 10, isAuthValid: false, client: nil, id: a}
	return &con
}

func (s *Connection) SetClient(client *http.Client) {
	s.client = client
}

func (s *Connection) IsLoginValid() (bool, error) {
	if s.dbVersion == "" {
		_, err := s.DbVersion()
		if err != nil {
			log.WithFields(log.Fields{"id": s.id, "error": err, "server": s.server}).Errorf("problem validate connection to AXL server.")
			return false, err
		}
	}
	log.WithFields(log.Fields{"id": s.id, "AxlUser": s.user, "server": s.server}).Infof("login to server is valid %t", s.isAuthValid)
	return s.isAuthValid, nil
}

func (s *Connection) DbVersion() (string, error) {
	log.WithFields(log.Fields{"id": s.id, "server": s.server}).Trace("start identify AXL DB version")
	var err error
	if s.dbVersion == "" {
		for i := 0; i < len(dbVersionSupport); i++ {
			s.dbVersion = dbVersionSupport[i]
			log.WithFields(log.Fields{"id": s.id, "server": s.server}).Debugf("test version [%s]", s.dbVersion)

			request := NewRequest(s.client, s)
			resp := request.DbVersionRequest()
			if resp.err != nil {
				s.dbVersion = DbVersionError
				resp.Close()
				return resp.lastMessage, resp.err
			}
			if resp.statusCode == 599 {
				s.dbVersion = DbVersionError
				resp.Close()
				continue
			}
			if resp.statusCode == 401 {
				s.isAuthValid = false
				resp.Close()
				log.WithFields(log.Fields{"id": s.id, "error": resp.statusMessage, "server": s.server}).Errorf("problem with AXL authorization")
				return "Problem with AXL authorization", fmt.Errorf(resp.statusMessage)
			}
			if resp.statusCode == 200 {
				v, err := VersionData(resp.GetResponseBody())
				if err == nil {
					log.WithFields(log.Fields{"id": s.id, "AXLVersion": v.Version, "AXL-DB": v.GetDbVersion(), "server": s.server}).Infof("actual AXL version [%s], DbVersion [%s]", v.Version, v.GetDbVersion())
					s.dbVersion = v.GetDbVersion()
					resp.Close()
					break
				}
				log.WithFields(log.Fields{"id": s.id, "server": s.server}).Warningf("problem convert XML data to version structure")
			}
			s.dbVersion = DbVersionError
			resp.Close()
		}
	}
	s.isAuthValid = true
	if s.dbVersion == DbVersionError {
		log.WithFields(log.Fields{"id": s.id, "error": err, "server": s.server}).Errorf("not support DB version.")
	}
	return s.dbVersion, err
}

func (s *Connection) ToParamString() string {
	return fmt.Sprintf("%s - %s", s.server, s.user)
}
