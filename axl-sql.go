package main

import (
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

const (
	token01 = "TOKEN_01"
	token02 = "TOKEN_02"
	token03 = "TOKEN_03"
)

const SelectCompleteTable = `select eu.pkid as user_pkid,
       d.pkid as device_pkid,
       np.pkid as line_pkid,
       eu.firstname,
       eu.middlename,
       eu.lastname,
       eu.userid,
       eu.department,
       eu.status,
       eu.islocaluser,
       eunp.uccx,
       eu.directoryuri,
       eu.mailid,
       d.name as devicename,
       d.description as devicedescrition,
       np.dnorpattern,
       np.alertingnameascii,
       (select paramvalue from processconfig where paramname = 'ClusterID') as cluster_name,
       np.description as line_description
from enduser eu
         LEFT OUTER JOIN (SELECT fkenduser, max(CASE tkdnusage WHEN 2 THEN tkdnusage ELSE null END) is not null AS uccx
                          FROM endusernumplanmap
                          GROUP BY fkenduser
) AS eunp ON eunp.fkenduser = eu.pkid
         INNER JOIN enduserdevicemap eudm ON eudm.fkenduser = eu.pkid
         INNER JOIN device d ON d.pkid = eudm.fkdevice
         INNER JOIN devicenumplanmap dnpm ON d.pkid = dnpm.fkdevice
         INNER JOIN numplan np ON np.pkid = dnpm.fknumplan
WHERE d.pkid IN (
    select fkdevice
    from applicationuserdevicemap
    where fkapplicationuser in (select au.pkid from applicationuser au where lower(name) in (` + token01 + `))
    union
    select fkdevice
    from enduserdevicemap
    where fkenduser in (select au.pkid from enduser au where lower(userid) in (` + token01 + `))
)
ORDER BY eu.pkid, d.pkid, np.pkid`

const SelectLoginUsers = `select enduser.pkid as user_pkid,
       enduser.firstname,
       enduser.middlename,
       enduser.lastname,
       enduser.userid,
       enduser.department,
       enduser.status,
       enduser.islocaluser,
       eunp.uccx,
       enduser.directoryuri,
       enduser.mailid,
       (select paramvalue from processconfig where paramname = 'ClusterID') as cluster_name
from enduser
         LEFT OUTER JOIN (SELECT fkenduser, max(CASE tkdnusage WHEN 2 THEN tkdnusage ELSE null END) is not null AS uccx
                          FROM endusernumplanmap
                          GROUP BY fkenduser
) AS eunp ON eunp.fkenduser = enduser.pkid
where enduser.pkid in (
    select e.fkenduser
    from enduserdirgroupmap as e
             inner join dirgroup as dg on e.fkdirgroup = dg.pkid
    where dg.name = '` + token01 + `')`

const SelectCompleteTableMax = "select * from device"

var tokens = []string{token01, token02, token03}

type ApiSqlBody interface {
	sqlParameterClean(data string) string
	addParameter(data string) int
	ToString() string
	IsParametersValid() bool
}

type SqlBody struct {
	id         string
	sql        string
	parameter  []string
	needParams int
}

func NewUserDeviceLineSql(users []string) *SqlBody {
	return stringListParameterSql(users, SelectCompleteTable)
}

func NewLoginUserSql(accessGroup string) *SqlBody {
	n := newSqlBody(SelectLoginUsers, 1)
	n.addParameter(accessGroup)
	return n
}

func stringListParameterSql(devices []string, sql string) *SqlBody {
	listId := ""
	addComa := ""
	for _, r := range devices {
		listId += addComa + "'" + r + "'"
		addComa = ","
	}
	n := newSqlBody(sql, 1)
	n.addParameter(listId)
	return n
}

func newSqlBody(sql string, parameters int) *SqlBody {
	a := RandomString()
	log.WithField("id", a).Tracef("Prepare SQL request with %d parameters", parameters)
	return &SqlBody{
		id:         a,
		sql:        sql,
		parameter:  nil,
		needParams: parameters,
	}
}

func (a *SqlBody) sqlParameterClean(data string) string {
	invalid := []string{"%", "*", "?", ";"}
	if data == "" {
		return ""
	}
	for _, s := range invalid {
		data = strings.ReplaceAll(data, s, " ")
	}
	return data
}

func (a *SqlBody) IsParametersValid() bool {
	var re = regexp.MustCompile(`(?mi)[\s;]+(insert|update|select|delete)[\s;]+`)
	ret := true
	for _, s := range a.parameter {
		match := re.MatchString(s)
		ret = ret && !match
	}
	return ret
}

func (a *SqlBody) addParameter(data string) int {
	data = a.sqlParameterClean(data)
	if data != "" {
		a.parameter = append(a.parameter, data)
	}
	return len(a.parameter)
}

func (a *SqlBody) ToString() string {
	if a.needParams > len(a.parameter) || !a.IsParametersValid() {
		return ""
	}
	data := a.sql
	for i, parameter := range a.parameter {
		data = strings.ReplaceAll(data, tokens[i], parameter)
	}
	return data
}
