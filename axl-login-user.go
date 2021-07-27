package main

import (
	"encoding/xml"
	log "github.com/sirupsen/logrus"
	"strings"
)

type LoginUserList struct {
	XMLName xml.Name    `xml:"return"`
	Rows    []LoginUser `xml:"row"`
}

type LoginUser struct {
	XMLName      xml.Name `xml:"row"`
	UserPKID     string   `xml:"user_pkid" json:"user_pkid"`
	FirstName    string   `xml:"firstname" json:"firstname"`
	MiddleName   string   `xml:"middlename" json:"middlename"`
	LastName     string   `xml:"lastname" json:"lastname"`
	UserId       string   `xml:"userid" json:"userid"`
	Department   string   `xml:"department" json:"department"`
	Status       int      `xml:"status" json:"status"`
	IsLocalUser  bool     `xml:"islocaluser" json:"islocaluser"`
	Uccx         bool     `xml:"uccx" json:"uccx"`
	DirectoryUri string   `xml:"directoryuri" json:"directoryuri"`
	MailId       string   `xml:"mailid" json:"mailid"`
	ClusterName  string   `xml:"cluster_name" json:"cluster_name"`
}

func NewLoginUserList(response string) (*LoginUserList, error) {
	var data LoginUserList
	d := []byte(response)
	err := xml.Unmarshal(d, &data)
	if err != nil {
		log.WithField("error", err).Errorf("problem unmarshal data from response for LoginUser")
		data = LoginUserList{Rows: []LoginUser{}}
	} else {
		log.WithField("data", response).Tracef("use thi list of data")

		if !config.Processing.CoexistCcxImporter {
			for i, _ := range data.Rows {
				data.Rows[i].Uccx = false
			}
			log.Tracef("remove UCCX Integration value, because coexist CCX Importer not used")
		}
	}
	return &data, err
}

func (s *Connection) GetLoginUserList() *LoginUserList {
	log.WithField("id", s.id).Trace("get table with login user details from AXL")
	sql := NewLoginUserSql(config.Axl.AccessGroup)
	if !sql.IsParametersValid() {
		log.WithField("id", s.id).Errorf("Not valid request parameters for access control group name")
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
	data, err := NewLoginUserList(response.GetResponseBody())
	if err != nil {
		log.WithField("id", s.id).Errorf("Can't convert returned data to Version structure")
		return nil
	}
	log.WithField("id", s.id).Trace("Success read DB version from AXL")
	return data
}
