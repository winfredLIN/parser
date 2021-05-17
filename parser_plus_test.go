package parser_test

import (
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"testing"
)

func TestParserPlus(t *testing.T) {
	parser := parser.New()

	stmt, _, err := parser.MustParse("OPTIMIZE TABLE foo;", "", "")
	if err != nil {
		t.Error(err)
		return
	}
	if _, ok := stmt[0].(*ast.UnknownNode); !ok {
		t.Errorf("expect stmt type is unknow, actual is %T", stmt)
		return
	}

	type testCase struct {
		sql    string
		expect []string
	}

	tc := []testCase{
		{
			sql: "SELECT * FROM db1.t1;OPTIMIZE TABLE foo;SELECT * FROM db2.t2",
			expect: []string{
				"SELECT * FROM db1.t1;",
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "OPTIMIZE TABLE foo;SELECT * FROM db1.t1;SELECT * FROM db2.t2",
			expect: []string{
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db1.t1;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "SELECT * FROM db1.t1;SELECT * FROM db2.t2;OPTIMIZE TABLE foo",
			expect: []string{
				"SELECT * FROM db1.t1;",
				"SELECT * FROM db2.t2;",
				"OPTIMIZE TABLE foo",
			},
		},
	}
	for _, c := range tc {
		stmt, _, err := parser.MustParse(c.sql, "", "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(c.expect) != len(stmt) {
			t.Errorf("expect sql length is %d, actual is %d", len(c.expect), len(stmt))
		}
		for i, s := range stmt {
			if s.Text() != c.expect[i] {
				t.Errorf("expect sql is %s, actual is %s", s.Text(), c.expect[i])
			}
		}
	}
}
