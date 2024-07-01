package parser

import "github.com/pingcap/parser/ast"

type ParserForSplitter struct {
	parser *Parser
}

func NewParserForSplitter() *ParserForSplitter {
	return &ParserForSplitter{
		parser: New(),
	}
}

func (p *ParserForSplitter) Parse(sql string) (stmt []ast.StmtNode, err error) {
	stmt, _, err = p.parser.Parse(sql, "", "")
	return
}

func (p *ParserForSplitter) Result() []ast.StmtNode {
	return p.parser.result
}
