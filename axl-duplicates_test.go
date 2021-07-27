package main

import (
	"math/rand"
	"testing"
)

type anyListTables = struct {
	name      string
	pkid      string
	addId     string
	duplicate bool
}

type uniqueTables struct {
	t *UniqueList
}

func processAddTest(t *testing.T, table anyListTables) {
	ul := NewUniqueList(table.name, "desc "+table.name, table.pkid)
	ul.Add(table.addId)
	if table.duplicate {
		if len(ul.pkid) != 1 {
			t.Errorf("duplicate pkid added [%s / %d]\r\n%s\r\n%s", table.name, len(ul.pkid), table.pkid, table.addId)
		}
	} else {
		if len(ul.pkid) != 2 {
			t.Errorf("unique pkid not added [%s / %d]\r\n%s\r\n%s", table.name, len(ul.pkid), table.pkid, table.addId)
		}
	}
}

func TestUniqueList_Add(t *testing.T) {
	t.Parallel()
	for _, table := range userListData {
		processAddTest(t, table)
	}
	for _, table := range deviceListData {
		processAddTest(t, table)
	}
	for _, table := range lineListData {
		processAddTest(t, table)
	}
}

func TestUniqueList_UserListString(t *testing.T) {

}

func TestContainsString(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t    []string
		a    string
		e    bool
		name string
	}{
		{t: []string{"aaa", "bbb", "aba"}, a: "ccc", e: false, name: "new last"},
		{t: []string{"ddd", "bbb", "aba"}, a: "aaa", e: false, name: "new first"},
		{t: []string{"aaa", "bbb", "ccc"}, a: "aba", e: false, name: "new middle"},
		{t: []string{"ddd", "bbb", "ccc"}, a: "ddd", e: true, name: "contains first"},
		{t: []string{"ddd", "bbb", "ccc"}, a: "ccc", e: true, name: "contains last"},
		{t: []string{"ddd", "bbb", "ccc"}, a: "bbb", e: true, name: "contains middle"},
		{t: []string{"ddd", "bbb", "ccc"}, a: "", e: false, name: "empty string"},
	}
	for _, table := range tables {
		if table.e != ContainsString(table.t, table.a) {
			t.Errorf("contains test fail [%s]", table.name)
		}
	}
}
func BenchmarkContainsString(b *testing.B) {
	benchmarkContainsString(100, false, b)
}
func BenchmarkContainsString2(b *testing.B) {
	benchmarkContainsString(100, true, b)
}

func benchmarkContainsString(size int, exist bool, b *testing.B) {
	var list []string
	test := RandomString()
	list = []string{}
	for i := 0; i < size; i++ {
		list = append(list, RandomString())
	}
	if exist {
		test = list[rand.Intn(size)]
	}
	for n := 0; n < b.N; n++ {
		_ = ContainsString(list, test)
	}
}

func generateUniqueTables() (tbl []uniqueTables) {
	var ut []uniqueTables
	ut = []uniqueTables{}
	for _, table := range userListData {
		ut = append(ut, uniqueTables{t: NewUniqueList(table.name, "desc "+table.name, table.pkid)})
	}

	return
}

var userListData = []anyListTables{
	{name: "agent01", pkid: "aaaaa-aaa-01", duplicate: false, addId: "baaaa-aaa-01"},
	{name: "agent02", pkid: "aaaaa-aaa-02", duplicate: false, addId: "baaaa-aaa-02"},
	{name: "agent03", pkid: "aaaaa-aaa-03", duplicate: true, addId: "aaaaa-aaa-03"},
}

var deviceListData = []anyListTables{
	{name: "device01", pkid: "ddddd-ddd-01", duplicate: false, addId: "bdddd-ddd-01"},
	{name: "device02", pkid: "ddddd-ddd-02", duplicate: false, addId: "bdddd-ddd-02"},
	{name: "device03", pkid: "ddddd-ddd-03", duplicate: true, addId: "ddddd-ddd-03"},
}

var lineListData = []anyListTables{
	{name: "line01", pkid: "lllll-lll-01", duplicate: false, addId: "bllll-lll-01"},
	{name: "line02", pkid: "lllll-lll-02", duplicate: false, addId: "bllll-lll-02"},
	{name: "line03", pkid: "lllll-lll-03", duplicate: true, addId: "lllll-lll-03"},
}
