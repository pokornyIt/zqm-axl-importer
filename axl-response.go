package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

type Response struct {
	id            string
	response      *http.Response
	err           error
	lastMessage   string
	body          string
	statusCode    int
	statusMessage string
}

func (s *Request) NewAxlResponse(r *http.Response, e error, message string) *Response {
	c := new(Response)
	c.id = s.id
	c.response = r
	c.err = e
	c.lastMessage = message
	if r != nil {
		c.statusCode = r.StatusCode
		c.statusMessage = r.Status
	} else {
		c.statusCode = 500
		c.statusMessage = "500 Problem Connect to server"
	}
	log.WithField("id", c.id).Debugf("Create new response")
	return c
}

func (r *Response) Close() {
	if r.response != nil && r.response.Body != nil {
		_ = r.response.Body.Close()
	}
	r.response = nil
}

func (r *Response) responseReturnData() error {
	log.WithFields(log.Fields{"id": r.id, "HTTPStatus": r.response.Status}).Debugf("Response status is %s", r.response.Status)
	bodies, err := ioutil.ReadAll(r.response.Body)
	_ = r.response.Body.Close()
	r.body = ""

	if err != nil {
		log.WithFields(log.Fields{"id": r.id, "error": err}).Errorf("Problem get body from response.")
		return err
	}
	if r.statusCode > 299 {
		r.getFailurePart(string(bodies))
	} else {
		r.getReturnPart(string(bodies))
	}
	log.WithField("id", r.id).Trace("Body read success")
	return nil
}

func (r *Response) getReturnPart(body string) {
	r.getBetween(body, "<return>", "</return>", "<return/>")
}

func (r *Response) getFailurePart(body string) {
	r.getBetween(body, "<soapenv:Fault>", "</soapenv:Fault>", "<soapenv:Fault/>")
}

func (r *Response) getBetween(body string, start string, end string, short string) {
	if strings.Index(body, start) > -1 {
		body = body[strings.Index(body, start):]
	} else if strings.Index(body, short) > -1 {
		r.body = short
		return
	}
	r.body = ""
	if strings.Index(body, end) < 0 {
		return
	}
	r.body = body[:strings.Index(body, end)] + end
}

func (r *Response) ResponseError() (string, error) {
	if r.statusCode == 401 {
		if r.err == nil {
			r.err = errors.New("authorization AXL problem")
		}
		return "Problem with AXL authorization", r.err
	}
	if r.statusCode == 599 {
		if r.err == nil {
			r.err = errors.New("database version problem of request format problem")
		}
		return "Database version or request format problem ", r.err
	}
	if r.statusCode == 200 {
		return "", nil
	}
	if r.statusCode == 500 {
		body := r.GetResponseBody()
		fault, err := NewFaultMessage(body)
		if err != nil {
			return "problem analyze fail response message", err
		}
		if fault.IsQueryTooLarge() {
			r.err = errors.New(fault.FaultString)
			return fault.FaultString, r.err
		}
	}
	if r.err == nil {
		r.err = errors.New(fmt.Sprintf("unspecific request problem HTTP status %s", r.statusMessage))
	}
	return fmt.Sprintf("Unspecific request problem HTTP status %s", r.statusMessage), r.err
}

func (r *Response) GetResponseBody() string {
	if r.response == nil {
		return r.body
	}
	err := r.responseReturnData()
	if err != nil {
		r.err = err
	}
	return r.body
}
