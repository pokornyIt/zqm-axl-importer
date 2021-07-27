package main

import (
	"encoding/xml"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

type FaultMessage struct {
	XMLName     xml.Name `xml:"Fault"`
	FaultCode   string   `xml:"faultcode"`
	FaultString string   `xml:"faultstring"`
	Detail      struct {
		Text     string `xml:",chardata"`
		AxlError struct {
			Text       string `xml:",chardata"`
			AxlCode    string `xml:"axlcode"`
			AxlMessage string `xml:"axlmessage"`
			Request    string `xml:"request"`
		} `xml:"axlError"`
	} `xml:"detail"`
}

func NewFaultMessage(response string) (*FaultMessage, error) {
	var data FaultMessage
	d := []byte(response)
	err := xml.Unmarshal(d, &data)
	if err != nil {
		log.WithField("error", err).Errorf("Problem unmarshal data from response")
		data = FaultMessage{}
	}
	return &data, err
}

func (f *FaultMessage) IsQueryTooLarge() bool {
	return strings.HasPrefix(f.FaultString, "Query request too large.")
}

func (f *FaultMessage) GetTotals() int {
	return f.getNumberFromString(1)
}

func (f *FaultMessage) GetFetchMax() int {
	return f.getNumberFromString(2)
}

func (f *FaultMessage) getNumberFromString(pos int) int {
	r := regexp.MustCompile(`^[^\d]+(\d+)[^\d]+(\d+).+$`)
	res := r.FindAllStringSubmatch(f.FaultString, -1)
	i, _ := strconv.Atoi(res[0][pos])
	return i
}
