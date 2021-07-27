package main

import (
	"fmt"
	"sort"
	"strings"
)

type UniqueList struct {
	name        string
	description string
	pkid        []string
}

type Duplicates struct {
	device map[string]*UniqueList
	line   map[string]*UniqueList
	user   map[string]*UniqueList
	errors []string
}

func (u *UniqueList) Add(pkid string) {
	if !ContainsString(u.pkid, pkid) {
		u.pkid = append(u.pkid, pkid)
	}
}

func (u *UniqueList) UserListString(user map[string]*UniqueList) string {
	sb := strings.Builder{}
	var comma string
	for _, pkid := range u.pkid {
		sb.WriteString(comma)
		sb.WriteString(user[pkid].name)
		//sb.WriteString(" ")
		//sb.WriteString(user[pkid].description)
		comma = ", "
	}
	return sb.String()
}

func (d *Duplicates) GenerateErrors() {
	d.errors = []string{}

	for key, val := range d.device {
		if len(d.device[key].pkid) > 1 {
			d.errors = append(d.errors, fmt.Sprintf("Device [%s - %s] associate to next User ID: [%s]", val.name, val.description, val.UserListString(d.user)))
		} else {
			delete(d.device, key)
		}
	}

	for key, val := range d.line {
		if len(d.line[key].pkid) > 1 {
			d.errors = append(d.errors, fmt.Sprintf("Line [%s - %s] associate to next User ID: [%s]", val.name, val.description, val.UserListString(d.user)))
		} else {
			delete(d.line, key)
		}
	}
}

func NewUniqueList(name string, description string, pkid string) *UniqueList {
	l := UniqueList{
		name:        name,
		description: description,
		pkid:        []string{pkid},
	}
	return &l
}

func ContainsString(s []string, search string) bool {
	t := append([]string{}, s...)
	sort.Strings(t)
	i := sort.SearchStrings(t, search)
	return i < len(s) && t[i] == search
}
