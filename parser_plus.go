package parser

import (
	"bytes"
	"github.com/pingcap/parser/ast"
)

// Parse parses a query string to raw ast.StmtNode.
// If charset or collation is "", default charset and collation will be used.
func (parser *Parser) MustParse(sql, charset, collation string) (stmt []ast.StmtNode, warns []error, err error) {
	_, warns, err = parser.Parse(sql, charset, collation)
	if err == nil {
		return parser.result, warns, nil
	}
	if len(parser.result) > 0 {
		for _, stmt := range parser.result {
			ast.SetFlag(stmt)
		}
		stmt = append(stmt, parser.result...)
	}

	buf := bytes.Buffer{}
	buf.WriteString(parser.lexer.r.s[parser.lexer.stmtStartPos:parser.lexer.lastScanOffset])

	// scan the remaining sql, there may be include unparseable sql.
	remainingSql := parser.lexer.r.s[parser.lexer.lastScanOffset:]
	l := NewScanner(remainingSql)
	var v yySymType
	var endOffset int
	for {
		result := l.Lex(&v)
		if result == 0 {
			endOffset = l.lastScanOffset - 1
			break
		}
		if result == ';' {
			endOffset = l.lastScanOffset
			break
		}
	}
	buf.WriteString(l.r.s[:endOffset+1])

	unknownSql := buf.String()
	if len(unknownSql) > 0 {
		un := &ast.UnknownNode{}
		un.SetText(unknownSql)
		stmt = append(stmt, un)
	}

	if l.lastScanOffset >= len(remainingSql) {
		return stmt, warns, nil
	}

	cStmt, cWarn, cErr := parser.MustParse(l.r.s[l.lastScanOffset+1:], charset, collation)
	warns = append(warns, cWarn...)
	if len(cStmt) > 0 {
		stmt = append(stmt, cStmt...)
	}
	if cErr == nil {
		return stmt, warns, cErr
	}
	return stmt, warns, nil
}
