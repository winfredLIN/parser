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

	var allNodes []ast.StmtNode
	for _, result := range results {
		s.parser.Parse(result.sql, "", "")
		if len(s.parser.result) == 1 {
			// 若结果集长度为1，则为单条且可解析的SQL
			stmt := s.parser.result[0]
			stmt.SetStartLine(result.line)
			allNodes = append(allNodes, stmt)
		} else {
			// 若结果集长度大于1，则为多条合并的SQL
			// 若结果集长度为0，则不可解析的SQL
			unParsedStmt := &ast.UnparsedStmt{}
			unParsedStmt.SetStartLine(result.line)
			unParsedStmt.SetText(result.sql)
			allNodes = append(allNodes, unParsedStmt)
		}
	}

	return allNodes, nil
}

func (s *splitter) ProcessToExecutableNodes(results []ast.StmtNode) []ast.StmtNode {

	var executableNodes []ast.StmtNode
	for _, node := range results {
		if stmt, isUnparsed := node.(*ast.UnparsedStmt); isUnparsed {
			if matched, _ := s.delimiter.matchAndSetCustomDelimiter(stmt.Text()); matched {
				continue
			}
		}
		trimmedSQL := strings.TrimSuffix(node.Text(), s.delimiter.delimiter())
		if trimmedSQL == "" {
			continue
		}
		node.SetText(trimmedSQL)
		executableNodes = append(executableNodes, node)
	}

	return executableNodes
}
