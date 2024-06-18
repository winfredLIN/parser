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

func (s *splitter) ParseSqlText(sqlText string) (allNodes []ast.StmtNode, err error) {
	results, err := s.delimiter.SplitSqlText(sqlText)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		s.parser.Parse(result.sql, "", "")

		if len(s.parser.result) == 1 {
			s.parser.result[0].SetStartLine(result.line)
			allNodes = append(allNodes, s.parser.result[0])
		} else {
			unParsedStmt := &ast.UnparsedStmt{}
			unParsedStmt.SetStartLine(result.line)
			unParsedStmt.SetText(result.sql)
			allNodes = append(allNodes, unParsedStmt)
		}
	}
	return allNodes, nil
}

func (s *splitter) ProcessToExecutableNodes(results []ast.StmtNode) (executableNodes []ast.StmtNode) {
	var currentDelimiter string = DefaultDelimiterString
	var sql string
	for _, node := range results {
		if stmt, isUnparsed := node.(*ast.UnparsedStmt); isUnparsed {
			if matched, delimiter := matchDelimiterCommand(stmt.Text()); matched {
				currentDelimiter = delimiter
				continue
			}
			if matched, delimiter := matchDelimiterCommandSort(stmt.Text()); matched {
				currentDelimiter = delimiter
				continue
			}
		}
		sql = strings.TrimSuffix(node.Text(), currentDelimiter)
		if sql == "" {
			continue
		}
		node.SetText(sql)
		executableNodes = append(executableNodes, node)
	}
	return executableNodes
}
