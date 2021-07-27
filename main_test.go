package main

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const compareGeneratedString = 10000

func TestRandomString(t *testing.T) {
	t.Parallel()
	var list []string
	list = []string{}
	for i := 0; i < compareGeneratedString; i++ {
		r := RandomString()
		if len(r) != maxRandomSize {
			t.Errorf("random string has bad lenghth %d/%d", len(r), maxRandomSize)
		} else {
			list = append(list, r)
		}
	}
	if len(list) != len(uniqueString(list)) {
		t.Errorf("in %d compared strings are %d duplication", compareGeneratedString, len(list)-len(uniqueString(list)))
	}
}

func TestShortBody(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t    string
		e    string
		name string
	}{
		{"", "", "empty"},
		{strings.Repeat("a", shortBodyChars), strings.Repeat("a", shortBodyChars-2) + " ...", fmt.Sprintf("%d chars", shortBodyChars)},
		{strings.Repeat("a", shortBodyChars-1), strings.Repeat("a", shortBodyChars-1), fmt.Sprintf("%d chars", shortBodyChars-1)},
		{"test test\nnot show\ndata", "not show\ndata", "split FL"},
	}
	for _, table := range tables {
		if ShortBody(table.t) != table.e {
			t.Errorf("unexpected ShortBody for [%d/%d]", len(ShortBody(table.t)), len(table.e))
		}
	}
}

func TestIsTimeToAxlUpdate(t *testing.T) {
	t.Parallel()
	tables := []struct {
		t       int
		hours   []int
		success bool
	}{
		{t: 4, hours: []int{4}, success: true},
		{t: 4, hours: []int{2, 4, 6}, success: true},
		{t: 4, hours: []int{1, 3, 5}, success: false},
	}
	config = NewConfig()
	for _, table := range tables {
		config.Processing.UserImportHour = table.hours
		for _, tim := range timeTable() {
			usedTime := time.Date(tim.Year(), tim.Month(), tim.Day(), table.t, tim.Minute(), tim.Second(), tim.Nanosecond(), tim.Location())
			ret := IsTimeToAxlUpdate(usedTime)
			if ret != table.success {
				t.Errorf("analyze for time %s is not valid in table [%s]", usedTime.Format("2006-01-02 15:04:05"), arrayToString(table.hours, ", "))
			}
		}
	}

}

func BenchmarkRandomString5(b *testing.B) {
	benchmarkRandomString(5, b)
}

func BenchmarkRandomString10(b *testing.B) {
	benchmarkRandomString(10, b)
}

func BenchmarkRandomString20(b *testing.B) {
	benchmarkRandomString(20, b)
}

func BenchmarkRandomString30(b *testing.B) {
	benchmarkRandomString(30, b)
}

func benchmarkRandomString(size int, b *testing.B) {
	maxRandomSize = size
	for n := 0; n < b.N; n++ {
		_ = RandomString()
	}
}

func BenchmarkShortBody100F(b *testing.B) {
	benchmarkShortBody(100, false, b)
}

func BenchmarkShortBody100T(b *testing.B) {
	benchmarkShortBody(100, true, b)
}

func BenchmarkShortBody200F(b *testing.B) {
	benchmarkShortBody(200, false, b)
}

func BenchmarkShortBody200T(b *testing.B) {
	benchmarkShortBody(200, true, b)
}

func benchmarkShortBody(size int, incLn bool, b *testing.B) {
	script := strings.Repeat("a", rand.Intn(size*5))
	if incLn {
		script = strings.Repeat("s", size/2) + "\n" + script
	}
	shortBodyChars = size
	for n := 0; n < b.N; n++ {
		_ = ShortBody(script)
	}
}

func uniqueString(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func timeTable() (t []time.Time) {
	for i := 0; i < 20; i++ {
		rnd := rand.Int63n(time.Now().Unix()-94608000) + 94608000
		t = append(t, time.Unix(rnd, 0))
	}
	return t
}

func arrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}
