package parser

import (
	"fmt"
	"os"
	"testing"
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