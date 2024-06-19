package parser

import (
	"strings"

	"github.com/pingcap/parser/ast"
)

type splitter struct {
	parser    *Parser
	delimiter *Delimiter
}

func NewSplitter() *splitter {
	return &splitter{
		parser:    New(),
		delimiter: NewDelimiter(),
	}
}

func (s *splitter) ParseSqlText(sqlText string) ([]ast.StmtNode, error) {
	results, err := s.delimiter.SplitSqlText(sqlText)
	if err != nil {
		return nil, err
	}
	return s.processToExecutableNodes(results), nil
}

func (s *splitter) processToExecutableNodes(results []*sqlWithLineNumber) []ast.StmtNode {
	s.delimiter.setDelimiter(DefaultDelimiterString)

	var executableNodes []ast.StmtNode
	for _, result := range results {
		if matched, _ := s.delimiter.matchAndSetCustomDelimiter(result.sql); matched {
			continue
		}
		trimmedSQL := strings.TrimSuffix(result.sql, s.delimiter.delimiter())
		if trimmedSQL == "" {
			continue
		}
		result.sql = trimmedSQL + ";"
		s.parser.Parse(result.sql, "", "")
		if len(s.parser.result) == 1 {
			// 若结果集长度为1，则为单条且可解析的SQL
			stmt := s.parser.result[0]
			stmt.SetStartLine(result.line)
			executableNodes = append(executableNodes, stmt)
		} else {
			// 若结果集长度大于1，则为多条合并的SQL
			// 若结果集长度为0，则不可解析的SQL
			unParsedStmt := &ast.UnparsedStmt{}
			unParsedStmt.SetStartLine(result.line)
			unParsedStmt.SetText(result.sql)
			executableNodes = append(executableNodes, unParsedStmt)
		}
	}
	return executableNodes
}
