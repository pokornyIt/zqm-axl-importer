package main

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Version struct {
	XMLName xml.Name `xml:"return"`
	Version string   `xml:"componentVersion>version"`
}

func VersionData(data string) (*Version, error) {
	var ver Version
	d := []byte(data)
	err := xml.Unmarshal(d, &ver)
	if err != nil {
		log.WithFields(log.Fields{"id": RandomString(), "error": err}).Errorf("Problem Unmarshal source data to Version structure.")
		ver = Version{
			Version: "",
		}
		return &ver, err
	}
	return &ver, nil
}

func (v *Version) ToString() string {
	return fmt.Sprintf("V: [%s}", v.Version)
}

func (v *Version) ToStringList() []string {
	var out []string
	out = append(out, v.ToString())
	return out
}

func (v *Version) IsValid() bool {
	return len(v.GetDbVersion()) > 0
}

func (v *Version) GetDbVersion() string {
	if v.Version == "" {
		return ""
	}
	ver := v.Version[:strings.Index(v.Version, ".")] + ".0"
	return ver
}

func (s *Connection) GetVersion() *Version {
	request := NewRequest(s.client, s)
	response := request.DbVersionRequest()
	msg, err := response.ResponseError()
	if err != nil {
		response.Close()
		log.WithField("id", s.id).Errorf("%s. HTTP Status [%s]", msg, response.statusMessage)
		return nil
	}
	data, err := VersionData(response.GetResponseBody())
	if err != nil {
		log.WithField("id", s.id).Errorf("Can't convert returned data to Version structure")
		return nil
	}
	log.WithField("id", s.id).Trace("Success read DB version from AXL")
	return data
}
