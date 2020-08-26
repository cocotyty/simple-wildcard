package simple_wildcard

import "testing"

var testTable = []struct {
	pattern string
	valid   []string
	invalid []string
}{
	{"host01.idc01", []string{"host01.idc01"}, []string{"host03.idc01"}},
	{"host[01:30].idc01", []string{"host29.idc01", "host03.idc01"}, []string{"host00.idc01", "host00.idc00"}},
	{"host[01:30].*", []string{"host29.idc01", "host03.idc02"}, []string{"host00.idc01", "host99.idc02"}},
	{"host[01:130].*", []string{"host129.idc01", "host03.idc02"}, []string{"host00.idc01", "host199.idc02"}},
	{"host*", []string{"host29.idc01", "host03.idc02"}, []string{"not"}},
}

func TestMatch(t *testing.T) {
	for _, row := range testTable {
		for _, target := range row.valid {
			if !Match(row.pattern, target) {
				t.Fatalf("match %s <> %s  expected matched but is not", row.pattern, target)
			}
		}
		for _, target := range row.invalid {
			if Match(row.pattern, target) {
				t.Fatalf("match %s <> %s  expected not matched but matched", row.pattern, target)
			}
		}
	}
}
