package parser

import (
	"fmt"
	"os"
	"testing"
)

func TestSplitter(t *testing.T) {
	s := NewSplitter()
	// 读取文件内容
	testCases := []struct {
		filePath       string
		expectedLength int
	}{
		{"splitter_test_1.sql", 14},
	}
	for _, testCase := range testCases {
		sqls, err := os.ReadFile(testCase.filePath)
		if err != nil {
			t.Fatalf("无法读取文件: %v", err)
		}
		splitResults, _, err := s.SplitSqlText(string(sqls), "", "")
		if err != nil {
			t.Fatalf(err.Error())
		}
		if len(splitResults) != testCase.expectedLength {
			t.FailNow()
		}
		for _, result := range splitResults {
			fmt.Print(result.Text())
			fmt.Print("\n-----------\n")
		}
	}
}

func TestSplitterProcess(t *testing.T) {
	s := NewSplitter()
	testCases := []struct {
		filePath       string
		expectedLength int
	}{
		{"splitter_test_1.sql", 9},
	}
	for _, testCase := range testCases {
		// 读取文件内容
		sql, err := os.ReadFile(testCase.filePath)
		if err != nil {
			t.Fatalf("无法读取文件: %v", err)
		}
		splitResults, _, err := s.SplitSqlText(string(sql), "", "")
		if err != nil {
			t.Fatalf(err.Error())
		}
		executableNodes := s.ProcessToExecutableNodes(splitResults)
		for _, node := range executableNodes {
			fmt.Println(node.Text())
			fmt.Print("\n-----------\n")
		}
		if len(executableNodes) != testCase.expectedLength {
			t.FailNow()
		}
	}
}

func TestMatchDelimiterCommand(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// 匹配到空字符串
		{"DELIMITER", ""},
		{"delimiter", ""},
		{"DELIMITER \n 'xx'", ""}, // 不允许换行，若换行则为空字符串
		// 一般情况
		{"DELIMITER -- use test", "--"},
		{"DELIMITER AbC123", "AbC123"},
		{"delimiter     ghi789  ", "ghi789"},
		// 使用引号，但无值
		{"DELIMITER ''", "''"},
		{"delimiter ``", "``"},
		{`delimiter ""`, `""`},
		// 使用引号，且有值
		{`DELIMITER "s s"`, `"s s"`},
		{`DELIMITER 'xx'`, `'xx'`},
		{"DELIMITER `aa`", "`aa`"},
	}

	for _, tc := range testCases {
		_, result := matchDelimiterCommand(tc.input)
		if result != tc.expected {
			t.Errorf("For input '%s', expected '%s', but got '%s'", tc.input, tc.expected, result)
		}
	}
}

func TestMatchDelimiterCommandSort(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// 匹配到空字符串
		{`\d`, ""},
		{`\d` + "\n 'xx'", ""}, // 不允许换行，若换行则为空字符串
		// 一般情况
		{`\d` + " -- use test", "--"},
		{`\d` + " AbC123", "AbC123"},
		{`\d` + "     ghi789  ", "ghi789"},
		// 使用引号，但无值
		{`\d` + " ''", "''"},
		{`\d` + " ``", "``"},
		{`\d` + ` ""`, `""`},
		// 使用引号，且有值
		{`\d` + ` "s s"`, `"s s"`},
		{`\d` + ` 'xx'`, `'xx'`},
		{`\d` + " `aa`", "`aa`"},
	}

	for _, tc := range testCases {
		_, result := matchDelimiterCommandSort(tc.input)
		if result != tc.expected {
			t.Errorf("For input '%s', expected '%s', but got '%s'", tc.input, tc.expected, result)
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