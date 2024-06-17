package parser

import (
	"bytes"
	"fmt"
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

func (s *splitter) SplitSqlText(sql, charset, collation string) (allNodes []ast.StmtNode, warns []error, err error) {
	// 先使用parser尝试解析能解析的SQL
	var stmtNodes []ast.StmtNode
	stmtNodes, warns, err = s.resloveSqlsByParser(sql, charset, collation)
	allNodes = append(allNodes, stmtNodes...)
	if err == nil {
		return allNodes, warns, nil
	}
	// 若未解析出任何SQL，则需要重置起始位置
	if len(stmtNodes) == 0 {
		s.parser.lexer.stmtStartPos = 0
	}
	// 根据parser的解析结果，使用未被解析的部分，解析出一条不支持解析的SQL
	var unParsedNode ast.StmtNode
	unParsedNode, err = s.resolveSqlsByDelimiter(sql[s.parser.lexer.stmtStartPos:])
	if err != nil {
		return allNodes, warns, err
	}
	allNodes = append(allNodes, unParsedNode)
	// 递归切分剩余SQL
	if s.parser.lexer.stmtStartPos < len(sql) {
		nodes, _, _ := s.SplitSqlText(sql[s.parser.lexer.stmtStartPos:], charset, collation)
		allNodes = append(allNodes, nodes...)
	}
	return allNodes, warns, nil
}

func (s *splitter) resloveSqlsByParser(sql, charset, collation string) (stmtNodes []ast.StmtNode, warns []error, err error) {
	_, warns, err = s.parser.Parse(sql, charset, collation)
	parseResults := s.parser.result
	s.parser.updateStartLineWithOffset(parseResults)
	if len(parseResults) > 0 {
		for _, stmtNode := range parseResults {
			ast.SetFlag(stmtNode)
		}
		stmtNodes = append(stmtNodes, parseResults...)
	}
	return stmtNodes, warns, err
}

func (s *splitter) resolveSqlsByDelimiter(sql string) (stmtNode ast.StmtNode, err error) {
	endOffset, err := s.delimiter.ScanNextEndOfSql(sql)
	if err != nil {
		return nil, err
	}
	unparsedStmtBuf := bytes.Buffer{}
	unparsedStmtBuf.WriteString(sql[:endOffset])
	unparsedSql := unparsedStmtBuf.String()
	unparsedSql = strings.TrimSpace(unparsedSql)
	if len(unparsedSql) > 0 {
		unParsedStmt := &ast.UnparsedStmt{}
		unParsedStmt.SetStartLine(s.parser.startLineOffset)
		unParsedStmt.SetText(unparsedSql)
		s.parser.lexer.stmtStartPos += endOffset
		return unParsedStmt, nil
	}
	return nil, fmt.Errorf("cannot reslove unparsed SQL")
}

func (s *splitter) ProcessToExecutableNodes(allNodes []ast.StmtNode) (executableNodes []ast.StmtNode) {
	var currentDelimiter string = DefaultDelimiterString
	var sql string
	for _, node := range allNodes {
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
