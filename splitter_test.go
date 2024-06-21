package parser

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/pingcap/parser/ast"
)

func TestSplitSqlText(t *testing.T) {
	s := NewSplitter()
	// 读取文件内容
	testCases := []struct {
		filePath       string
		expectedLength int
	}{
		{"splitter_test_1.sql", 20},
		{"splitter_test_2.sql", 20},
		{"splitter_test_3.sql", 4},
	}
	for _, testCase := range testCases {
		sqls, err := os.ReadFile(testCase.filePath)
		if err != nil {
			t.Fatalf("无法读取文件: %v", err)
		}
		splitResults, err := s.splitSqlText(string(sqls))
		if err != nil {
			t.Fatalf(err.Error())
		}
		if len(splitResults) != testCase.expectedLength {
			t.FailNow()
		}
		for _, result := range splitResults {
			fmt.Print("\n------------------------------\n")
			fmt.Printf("SQL语句在第%v行\n", result.line)
			fmt.Printf("SQL语句为:\n%v", result.sql)
		}
	}
}

func TestSplitterProcess(t *testing.T) {
	s := NewSplitter()
	testCases := []struct {
		filePath       string
		expectedLength int
	}{
		{"splitter_test_1.sql", 14},
		// {"splitter_test_2.sql", 14},
	}
	for _, testCase := range testCases {
		// 读取文件内容
		sqlText, err := os.ReadFile(testCase.filePath)
		if err != nil {
			t.Fatalf("无法读取文件: %v", err)
		}
		executableNodes, err := s.ParseSqlText(string(sqlText))
		if err != nil {
			t.Fatalf(err.Error())
		}
		for _, node := range executableNodes {
			fmt.Print("\n------------------------------\n")
			fmt.Printf("SQL语句在第%v行\n", node.StartLine())
			fmt.Printf("SQL语句为:\n%v", node.Text())
		}
		if len(executableNodes) != testCase.expectedLength {
			t.FailNow()
		}
	}
}

func TestIsDelimiterReservedKeyWord(t *testing.T) {
	tests := []struct {
		delimiter string
		expected  bool
	}{
		// 非关键字
		{"id", false},
		{"$$", false},
		{";;", false},
		{"\\", false},
		{"Abscsd", false},
		{"%%", false},
		{"|", false},
		{"%", false},
		{"foo", false},
		{"column1", false},
		{"table_name", false},
		{"_underscore", false},
		// 关键字
		{"&&", true},
		{"=", true},
		{"!=", true},
		{"<=", true},
		{">=", true},
		{"||", true},
		{"<>", true},
		{"IN", true},
		{"AS", true},
		{"Update", true},
		{"Delete", true},
		{"not", true},
		{"Order", true},
		{"by", true},
		{"Select", true},
		{"From", true},
		{"Where", true},
		{"Join", true},
		{"Inner", true},
		{"Left", true},
		{"Right", true},
		{"Full", true},
		{"Group", true},
		{"Having", true},
		{"Insert", true},
		{"Into", true},
		{"Values", true},
		{"Create", true},
		{"Table", true},
		{"Alter", true},
		{"Drop", true},
		{"Truncate", true},
		{"Union", true},
		{"Exists", true},
		{"Like", true},
		{"Distinct", true},
		{"And", true},
		{"Or", true},
		{"Limit", true},
		{"ALL", true},
		{"ANY", true},
		{"BETWEEN", true},
	}

	for _, test := range tests {
		t.Run(test.delimiter, func(t *testing.T) {
			result := isReservedKeyWord(test.delimiter)
			if result != test.expected {
				t.Errorf("isDelimiterReservedKeyWord(%s) = %v; want %v", test.delimiter, result, test.expected)
			}
		})
	}
}

func TestSkipQuotedDelimiter(t *testing.T) {
	s := NewSplitter()
	// 读取文件内容
	sqls, err := os.ReadFile("splitter_test_skip_quoted_delimiter.sql")
	if err != nil {
		t.Fatalf("无法读取文件: %v", err)
	}
	splitResults, err := s.splitSqlText(string(sqls))
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, result := range splitResults {
		fmt.Print("------------------------------\n")
		fmt.Printf("SQL语句在第%v行\n", result.line)
		fmt.Printf("SQL语句为:\n%v\n", result.sql)
	}
	if len(splitResults) != 26 {
		t.FailNow()
	}
}

func TestStartLine(t *testing.T) {
	// 测试用例第2个到第5个sql是解析器不能解析的sql
	p := NewSplitter()
	stmts, err := p.ParseSqlText(`grant all on point_trans_shard_00_part_202401 to kgoldpointapp;
create table point_trans_shard_00_part_202401(like point_trans_shard_00 including all) inherits(point_trans_shard_00);
Alter table point_trans_shard_00_part_202401 ADD CONSTRAINT chk_point_trans_shard_202401 CHECK (processedtime >= '1704038400000'::bigint AND processedtime < '1706716800000'::bigint );
create table point_trans_source_shard_00_part_202401(like point_trans_source_shard_00 including all) inherits(point_trans_source_shard_00);
Alter table point_trans_source_shard_00_part_202401 ADD CONSTRAINT chk_point_trans_source_shard_202401 CHECK (processedtime >= '1704038400000'::bigint AND processedtime < '1706716800000'::bigint );
grant select on point_trans_shard_00_part_202401 to prd_fin, dbsec, sec_db_scan;
grant all on point_trans_source_shard_00_part_202401 to kgoldpointapp;
grant select on point_trans_source_shard_00_part_202401 to prd_fin, dbsec, sec_db_scan;
`)
	if err != nil {
		t.Error(err)
		return
	}
	if len(stmts) != 8 {
		t.Errorf("expect 2 stmts, actual is %d", len(stmts))
		return
	}
	for i, stmt := range stmts {
		if stmt.StartLine() != i+1 {
			t.Errorf("expect start line is %d, actual is %d", i+1, stmt.StartLine())
		}
	}

	// 所有测试用例都是可以解析的sql
	stmts, err = p.ParseSqlText(`grant select on point_trans_shard_00_part_202401 to prd_fin, dbsec, sec_db_scan;
grant all on point_trans_source_shard_00_part_202401 to kgoldpointapp;
grant select on point_trans_source_shard_00_part_202401 to prd_fin, dbsec, sec_db_scan;
`)
	if err != nil {
		t.Error(err)
		return
	}
	if len(stmts) != 3 {
		t.Errorf("expect 3 nodes, actual is %d", len(stmts))
		return
	}
	for i, node := range stmts {
		if node.StartLine() != i+1 {
			t.Errorf("expect start line is %d, actual is %d", i+1, node.StartLine())
		}
	}

	// 所有测试用例都是不可以解析的sql
	stmts, err = p.ParseSqlText(`create table point_trans_shard_00_part_202401(like point_trans_shard_00 including all) inherits(point_trans_shard_00);
Alter table point_trans_shard_00_part_202401 ADD CONSTRAINT chk_point_trans_shard_202401 CHECK (processedtime >= '1704038400000'::bigint AND processedtime < '1706716800000'::bigint );
create table point_trans_source_shard_00_part_202401(like point_trans_source_shard_00 including all) inherits(point_trans_source_shard_00);`)
	if err != nil {
		t.Error(err)
		return
	}
	if len(stmts) != 3 {
		t.Errorf("expect 3 stmts, actual is %d", len(stmts))
		return
	}
	for i, stmt := range stmts {
		if stmt.StartLine() != i+1 {
			t.Errorf("expect start line is %d, actual is %d", i+1, stmt.StartLine())
		}
	}

	// 并排sql测试用例,备注:3个sql都不能被解析
	stmts, err = p.ParseSqlText(`create table point_trans_shard_00_part_202401(like point_trans_shard_00 including all) inherits(point_trans_shard_00);
Alter table point_trans_shard_00_part_202401 ADD CONSTRAINT chk_point_trans_shard_202401 CHECK (processedtime >= '1704038400000'::bigint AND processedtime < '1706716800000'::bigint );create table point_trans_source_shard_00_part_202401(like point_trans_source_shard_00 including all) inherits(point_trans_source_shard_00);`)
	if err != nil {
		t.Error(err)
		return
	}
	if len(stmts) != 3 {
		t.Errorf("expect 3 stmts, actual is %d", len(stmts))
		return
	}

	for i, stmt := range stmts {
		if i == 2 {
			if stmt.StartLine() != 2 {
				t.Errorf("expect start line is 2, actual is %d", stmt.StartLine())
			}
		} else {
			if stmt.StartLine() != i+1 {
				t.Errorf("expect start line is %d, actual is %d", i+1, stmt.StartLine())
			}
		}
	}
}

func TestPerfectParse(t *testing.T) {
	parser := NewSplitter()

	stmt, err := parser.ParseSqlText("OPTIMIZE TABLE foo;")
	if err != nil {
		t.Error(err)
		return
	}
	if _, ok := stmt[0].(*ast.UnparsedStmt); !ok {
		t.Errorf("expect stmt type is unparsedStmt, actual is %T", stmt)
		return
	}

	type testCase struct {
		sql    string
		expect []string
	}

	tc := []testCase{
		{
			sql: `SELECT * FROM db1.t1`,
			expect: []string{
				`SELECT * FROM db1.t1`,
			},
		},
		{
			sql: `SELECT * FROM db1.t1;SELECT * FROM db2.t2`,
			expect: []string{
				"SELECT * FROM db1.t1",
				"SELECT * FROM db2.t2",
			},
		},
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
		{
			sql: "SELECT FROM db2.t2 where a=\"asd;\"; SELECT * FROM db1.t1;",
			expect: []string{
				"SELECT FROM db2.t2 where a=\"asd;\";",
				" SELECT * FROM db1.t1;",
			},
		},
		{
			sql: "SELECT * FROM db1.t1;OPTIMIZE TABLE foo;OPTIMIZE TABLE foo;SELECT * FROM db2.t2",
			expect: []string{
				"SELECT * FROM db1.t1;",
				"OPTIMIZE TABLE foo;",
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "OPTIMIZE TABLE foo;SELECT * FROM db1.t1;OPTIMIZE TABLE foo;SELECT * FROM db2.t2",
			expect: []string{
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db1.t1;",
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "SELECT * FROM db1.t1;OPTIMIZE TABLE foo;SELECT * FROM db2.t2;OPTIMIZE TABLE foo",
			expect: []string{
				"SELECT * FROM db1.t1;",
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db2.t2;",
				"OPTIMIZE TABLE foo",
			},
		},
		{
			sql: `
CREATE PROCEDURE proc1(OUT s int)
BEGIN
END;
`,
			expect: []string{
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
END;`,
			},
		},
		{
			sql: `
CREATE PROCEDURE proc1(OUT s int)
BEGIN
SELECT COUNT(*)  FROM user;
END;
`,
			expect: []string{
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
SELECT COUNT(*)  FROM user;
END;`,
			},
		},
		{
			sql: `
CREATE PROCEDURE proc1(OUT s int)
BEGIN
SELECT COUNT(*)  FROM user;
SELECT COUNT(*)  FROM user;
END;
`,
			expect: []string{
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
SELECT COUNT(*)  FROM user;
SELECT COUNT(*)  FROM user;
END;`,
			},
		},
		{
			sql: `
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
END;
`,
			expect: []string{
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
END;`,
			},
		},
		{
			sql: `
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
`,
			expect: []string{
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
			},
		},
		{
			sql: `
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
SELECT * FROM db1.t1;
`,
			expect: []string{
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`SELECT * FROM db1.t1;`,
			},
		},
		{
			sql: `
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
SELECT * FROM db1.t1;
`,
			expect: []string{
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`SELECT * FROM db1.t1;`,
			},
		},
		{
			sql: `
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
SELECT * FROM db1.t1;
`,
			expect: []string{
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`SELECT * FROM db1.t1;`,
			},
		},
		{ // 匹配特殊字符结束
			sql: "select * from  �E",
			expect: []string{
				`select * from  �E`,
			},
		},
		{ // 匹配特殊字符后是;
			sql: "select * from  �E;select * from t1",
			expect: []string{
				`select * from  �E;`,
				"select * from t1",
			},
		},
		{ // 匹配特殊字符在中间
			sql: "select * from  �E where id = 1;select * from  �E ",
			expect: []string{
				`select * from  �E where id = 1;`,
				`select * from  �E `,
			},
		},
		{ // 匹配特殊字符在开头
			sql: " where id = 1;select * from  �E ",
			expect: []string{
				` where id = 1;`,
				`select * from  �E `,
			},
		},
		{ // 匹配特殊字符在SQL开头
			sql: "select * from  �E ; where id = 1",
			expect: []string{
				`select * from  �E ;`,
				` where id = 1`,
			},
		},
		{ // 匹配其他invalid场景
			sql: "@`",
			expect: []string{
				"@`",
			},
		},
		{ // 匹配其他invalid场景
			sql: "@` ;select * from t1",
			expect: []string{
				"@` ;select * from t1",
			},
		},
	}
	for _, c := range tc {
		stmt, err := parser.splitSqlText(c.sql)
		if err != nil {
			t.Error(err)
			return
		}
		if len(c.expect) != len(stmt) {
			t.Errorf("expect sql length is %d, actual is %d, sql is [%s]", len(c.expect), len(stmt), c.sql)
		} else {
			for i, s := range stmt {
				// 之前的测试用例预期对SQL的切分会保留SQL语句的前后的空格
				// 现在的切分会将SQL前后的空格去掉
				// 这里统一修改为匹配SQL语句，除去分隔符后的内容是否相等
				if strings.TrimSuffix(s.sql, ";") != strings.TrimSuffix(strings.TrimSpace(c.expect[i]), ";") {
					t.Errorf("expect sql is [%s], actual is [%s]", c.expect[i], s.sql)
				}
			}
		}
	}
}
