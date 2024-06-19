package parser

import (
	"fmt"
	"os"
	"testing"
)

func TestSplitSqlText(t *testing.T) {
	d := NewDelimiter()
	// 读取文件内容
	testCases := []struct {
		filePath       string
		expectedLength int
	}{
		{"splitter_test_1.sql", 20},
		{"splitter_test_2.sql", 20},
	}
	for _, testCase := range testCases {
		sqls, err := os.ReadFile(testCase.filePath)
		if err != nil {
			t.Fatalf("无法读取文件: %v", err)
		}
		splitResults, err := d.SplitSqlText(string(sqls))
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
		{"splitter_test_2.sql", 14},
	}
	for _, testCase := range testCases {
		// 读取文件内容
		sqlText, err := os.ReadFile(testCase.filePath)
		if err != nil {
			t.Fatalf("无法读取文件: %v", err)
		}
		allNodes, err := s.ParseSqlText(string(sqlText))
		if err != nil {
			t.Fatalf(err.Error())
		}
		executableNodes := s.ProcessToExecutableNodes(allNodes)
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

func TestRemoveOuterQuotes(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// 使用引号
		{"'hello'", "hello"},
		{`"world"`, "world"},
		{"`go`", "go"},
		{"foo", "foo"},
		// 引号嵌套
		{"``", ""},
		{"`''`", "''"},
		// 不使用引号包裹
		{"'bar'baz", "'bar'baz"},
		{`"did"did`, `"did"did`},
		{`did`, `did`},
	}

	for _, tc := range testCases {
		result := removeOuterQuotes(tc.input)
		if result != tc.expected {
			t.Errorf("removeOuterQuotes(%q) = %q, expected %q", tc.input, result, tc.expected)
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

func TestIsCommentLikeC(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"/* This is a comment */", true},
		{"/* comment with special chars !@#$%^&*()_+ */", true},
		{"/* */", true},
		{"/**/", true},
		{"/* unclosed comment", false},
		{"just some text", false},
		{"// single line comment", false},
		{"/* This is a comment */ extra text", false},
		{" extra text /* This is a comment */", false},
	}

	for _, test := range tests {
		result := isCommentLikeC(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected %v but got %v", test.input, test.expected, result)
		}
	}
}
