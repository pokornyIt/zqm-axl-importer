package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const xmlHeaderFormat = "<soapenv:Envelope xmlns:soapenv=\"http://schemas.xmlsoap.org/soap/envelope/\" xmlns:ns=\"http://www.cisco.com/AXL/API/%s\">"
const cmVersion = "<soapenv:Header/><soapenv:Body><ns:getCCMVersion sequence=\"%d\">\n</ns:getCCMVersion></soapenv:Body></soapenv:Envelope>"
const sqlRequest = "<soapenv:Header/><soapenv:Body><ns:executeSQLQuery sequence=\"%d\">\n<sql>%s</sql></ns:executeSQLQuery></soapenv:Body></soapenv:Envelope>"

type Request struct {
	id         string
	client     *http.Client
	connection *Connection
	request    *http.Request
}

func NewRequest(client *http.Client, connection *Connection) *Request {
	r := Request{
		id:         RandomString(),
		client:     client,
		connection: connection,
	}
	log.WithFields(log.Fields{"id": r.id, "server": r.connection.server}).Tracef("Prepare new request")
	return &r
}

func (s *Request) getCmVersionBody() string {
	s.connection.sequence++
	return fmt.Sprintf(xmlHeaderFormat+cmVersion, s.connection.dbVersion, s.connection.sequence)
}

func (s *Request) getSqlRequestBody(sql string) string {
	s.connection.sequence++
	return fmt.Sprintf(xmlHeaderFormat+sqlRequest, s.connection.dbVersion, s.connection.sequence, sql)
}

func (s *Request) DbVersionRequest() *Response {
	sql := s.getCmVersionBody()
	return s.doAxlRequest(sql)
}

func (s *Request) SqlPagingGenerate(sql string, size int, total int) []string {
	var data []string
	if size >= total {
		data = append(data, sql)
	} else {
		loop := total / size
		for loop*size < total {
			loop++
		}
		selectNext := sql[len("select"):]
		for i := 0; i < loop; i++ {
			data = append(data, fmt.Sprintf("SELECT SKIP %d LIMIT %d %s", loop*size, size, selectNext))
		}
	}

	return data
}

func (s *Request) SqlRequest(sql string) *Response {
	d := s.getSqlRequestBody(sql)
	return s.doAxlRequest(d)
}

func (s *Connection) authorization() string {
	auth := s.user + ":" + s.pwd
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (s *Request) setHeader() {
	s.request.Header.Set("Content-Type", "text/xml")
	s.request.Header.Set("Rows-Agent", "Recording Info 1.0")
	s.request.Header.Set("Accept", "*/*")
	s.request.Header.Set("Cache-Control", "no-cache")
	s.request.Header.Set("Pragma", "no-cache")
	//s.request.Header.Set("Authorization", "Basic "+s.connection.authorization())
	s.request.Host = s.connection.server
	s.request.SetBasicAuth(s.connection.user, s.connection.pwd)
}

func (s *Connection) urlString() string {
	urlName := fmt.Sprintf("https://%s:8443/axl/", s.server)
	log.WithFields(log.Fields{"id": s.id, "server": s.server}).Tracef("Request URI: %s", urlName)
	return urlName
}

func (s *Request) Client() {
	if s.client == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		s.client = &http.Client{Timeout: time.Duration(s.connection.timeOut) * time.Second, Transport: tr}
	}
}

func (s *Request) finishRequest() *Response {
	s.setHeader()
	s.Client()
	resp, err := s.client.Do(s.request)
	if err != nil {
		log.WithFields(log.Fields{"id": s.id, "error": err, "server": s.connection.server}).Errorf("Problem %s response.", s.request.Method)
		return s.NewAxlResponse(nil, err, "Problem "+s.request.Method+" response")
	}
	return s.NewAxlResponse(resp, nil, "")
}

func (s *Request) doAxlRequest(body string) *Response {
	log.WithFields(log.Fields{"id": s.id, "server": s.connection.server, "body": ShortBody(body)}).Trace("Process AXL request")

	req, err := http.NewRequest("POST", s.connection.urlString(), bytes.NewBuffer([]byte(body)))
	if err != nil {
		log.WithFields(log.Fields{"id": s.id, "error": err, "server": s.connection.server}).Errorf("Problem create new POST request.")
		return s.NewAxlResponse(nil, err, "Problem create new POST request.")
	}
	s.request = req
	log.WithFields(log.Fields{"id": s.id, "server": s.connection.server}).Trace("Success create request")
	return s.finishRequest()
}

func (s *Request) doPageAxlRequest() {

}
