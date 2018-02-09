package xorm

import (
	"reflect"
	"testing"

	"github.com/go-xorm/core"
)

func TestParsePostgres(t *testing.T) {
	tests := []struct {
		in       string
		expected string
		valid    bool
	}{
		{"postgres://auser:password@localhost:5432/db?sslmode=disable", "db", true},
		{"postgresql://auser:password@localhost:5432/db?sslmode=disable", "db", true},
		{"postg://auser:password@localhost:5432/db?sslmode=disable", "db", false},
		//{"postgres://auser:pass with space@localhost:5432/db?sslmode=disable", "db", true},
		//{"postgres:// auser : password@localhost:5432/db?sslmode=disable", "db", true},
		{"postgres://%20auser%20:pass%20with%20space@localhost:5432/db?sslmode=disable", "db", true},
		//{"postgres://auser:パスワード@localhost:5432/データベース?sslmode=disable", "データベース", true},
		{"dbname=db sslmode=disable", "db", true},
		{"user=auser password=password dbname=db sslmode=disable", "db", true},
		{"", "db", false},
		{"dbname=db =disable", "db", false},
	}

	driver := core.QueryDriver("postgres")

	for _, test := range tests {
		uri, err := driver.Parse("postgres", test.in)

		if err != nil && test.valid {
			t.Errorf("%q got unexpected error: %s", test.in, err)
		} else if err == nil && !reflect.DeepEqual(test.expected, uri.DbName) {
			t.Errorf("%q got: %#v want: %#v", test.in, uri.DbName, test.expected)
		}
	}
}
