package parser

import "github.com/pingcap/parser/ast"

type ParserForSplitter struct {
	Parser *Parser
}

func NewParserForSplitter() *ParserForSplitter {
	return &ParserForSplitter{
		Parser: New(),
	}
}

func (p ParserForSplitter) Result() []ast.StmtNode {
	return p.Parser.result
}
