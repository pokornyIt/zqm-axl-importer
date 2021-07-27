package main

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type UserDeviceLineList struct {
	XMLName xml.Name         `xml:"return"`
	Rows    []UserDeviceLine `xml:"row"`
}

type UserDeviceLine struct {
	XMLName           xml.Name `xml:"row"`
	UserPKID          string   `xml:"user_pkid" json:"user_pkid"`
	DevicePKID        string   `xml:"device_pkid" json:"device_pkid"`
	LinePKID          string   `xml:"line_pkid" json:"line_pkid"`
	FirstName         string   `xml:"firstname" json:"firstname"`
	MiddleName        string   `xml:"middlename" json:"middlename"`
	LastName          string   `xml:"lastname" json:"lastname"`
	UserId            string   `xml:"userid" json:"userid"`
	Department        string   `xml:"department" json:"department"`
	Status            int      `xml:"status" json:"status"`
	IsLocalUser       bool     `xml:"islocaluser" json:"islocaluser"`
	Uccx              bool     `xml:"uccx" json:"uccx"`
	DirectoryUri      string   `xml:"directoryuri" json:"directoryuri"`
	MailId            string   `xml:"mailid" json:"mailid"`
	DeviceName        string   `xml:"devicename" json:"devicename"`
	DeviceDescription string   `xml:"devicedescrition" json:"devicedescrition"`
	LineNumber        string   `xml:"dnorpattern" json:"dnorpattern"`
	LineAlertingName  string   `xml:"alertingnameascii" json:"alertingnameascii"`
	LineDescription   string   `xml:"line_description" json:"line_description"`
	ClusterName       string   `xml:"cluster_name" json:"cluster_name"`
}

func NewUserDeviceLineList(response string) (*UserDeviceLineList, error) {
	var data UserDeviceLineList
	d := []byte(response)
	err := xml.Unmarshal(d, &data)
	if err != nil {
		log.WithField("error", err).Errorf("problem unmarshal data from response")
		data = UserDeviceLineList{Rows: []UserDeviceLine{}}
	} else {
		log.WithField("data", response).Tracef("use this list of data")
		if !config.Processing.CoexistCcxImporter {
			for i, _ := range data.Rows {
				data.Rows[i].Uccx = false
			}
			log.Tracef("remove UCCX Integration value")
		}
	}
	return &data, err
}

func (u *UserDeviceLineList) GetDuplicateDevices() *Duplicates {

	data := Duplicates{
		device: make(map[string]*UniqueList),
		line:   make(map[string]*UniqueList),
		user:   make(map[string]*UniqueList),
		errors: []string{},
	}

	for _, r := range u.Rows {
		if _, ok := data.device[r.DevicePKID]; ok {
			data.device[r.DevicePKID].Add(r.UserPKID)
		} else {
			data.device[r.DevicePKID] = NewUniqueList(r.DeviceName, r.DeviceDescription, r.UserPKID)
		}

		if _, ok := data.line[r.LinePKID]; ok {
			data.line[r.LinePKID].Add(r.UserPKID)
		} else {
			data.line[r.LinePKID] = NewUniqueList(r.LineNumber, r.LineAlertingName, r.UserPKID)
		}

		if _, ok := data.user[r.UserPKID]; ok {
		} else {
			data.user[r.UserPKID] = NewUniqueList(r.UserId, fmt.Sprintf("%s %s", r.FirstName, r.LastName), "x")
		}
	}
	data.GenerateErrors()

	return &data
}

func (u *UserDeviceLineList) cleanDeviceLineList() []UserDeviceLine {
	log.WithField("rows", len(u.Rows)).Debugf("from AXL select %d rows combination user/device/line", len(u.Rows))
	duplicates := u.GetDuplicateDevices()
	if len(duplicates.errors) > 0 {
		log.Error("all duplicate association remove from source data")
		for _, d := range duplicates.errors {
			log.Error(d)
		}
	}
	ret := u.removeDuplicates(duplicates)
	return ret
}

func (u *UserDeviceLine) inDuplicates(dup *Duplicates) bool {
	if _, ok := dup.device[u.DevicePKID]; ok {
		return true
	}
	if _, ok := dup.line[u.LinePKID]; ok {
		return true
	} else {
		return false
	}
}

func (u *UserDeviceLineList) removeDuplicates(dup *Duplicates) []UserDeviceLine {
	var response []UserDeviceLine

	for idx := 0; idx < len(u.Rows); idx++ {
		if !u.Rows[idx].inDuplicates(dup) {
			response = append(response, u.Rows[idx])
		}
	}
	log.WithField("removedRows", len(u.Rows)-len(response)).Infof("From source AXL table remove %d rows", len(u.Rows)-len(response))
	return response
}

func (s *Connection) GetUserDeviceLineList() *UserDeviceLineList {
	log.WithField("id", s.id).Trace("get table with user/device/line details from AXL")
	var jtapi []string
	users := strings.Join(config.Zqm.JtapiUser, "','")
	jtapi = append(jtapi, strings.ToLower(users))
	sql := NewUserDeviceLineSql(jtapi)
	if !sql.IsParametersValid() {
		log.WithField("id", s.id).Errorf("Not valid request parameters")
		return nil
	}
	request := NewRequest(s.client, s)
	log.WithFields(log.Fields{"id": s.id, "sql": sql.ToString()}).Debugf("Request for %s", strings.Join(config.Zqm.JtapiUser, ","))
	response := request.SqlRequest(sql.ToString())
	msg, err := response.ResponseError()
	if err != nil {
		response.Close()
		log.WithField("id", s.id).Errorf("%s. HTTP Status [%s]", msg, response.statusMessage)
		return nil
	}
	data, err := NewUserDeviceLineList(response.GetResponseBody())
	if err != nil {
		log.WithField("id", s.id).Errorf("Can't convert returned data to Version structure")
		return nil
	}
	log.WithField("id", s.id).Trace("Success read DB version from AXL")
	return data
}
