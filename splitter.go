package parser

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"
)

type splitter struct {
	parser    *Parser
	delimiter *Delimiter
	blocker   *Blocker
	scanner   *ScannerForSplitter
}

func NewSplitter() *splitter {
	return &splitter{
		parser:    New(),
		delimiter: NewDelimiter(),
		blocker:   NewBlocker(),
		scanner:   NewScannerForSplitter(),
	}
}

func (s *splitter) ParseSqlText(sqlText string) ([]ast.StmtNode, error) {
	s.delimiter.reset()
	results, err := s.splitSqlText(sqlText)
	if err != nil {
		return nil, err
	}
	return s.processToExecutableNodes(results), nil
}

func (s *splitter) processToExecutableNodes(results []*sqlWithLineNumber) []ast.StmtNode {
	s.delimiter.reset()

	var executableNodes []ast.StmtNode
	for _, result := range results {
		if matched, _ := s.matchAndSetCustomDelimiter(result.sql); matched {
			continue
		}
		trimmedSQL := strings.TrimSuffix(result.sql, s.delimiter.DelimiterStr)
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

type sqlWithLineNumber struct {
	sql  string
	line int
}

func (s *splitter) splitSqlText(sqlText string) (results []*sqlWithLineNumber, err error) {
	result, err := s.getNextSql(sqlText)
	if err != nil {
		return nil, err
	}
	results = append(results, result)
	// 递归切分剩余SQL
	if s.scanner.Offset() < len(sqlText) {
		subResults, _ := s.splitSqlText(sqlText[s.scanner.Offset():])
		results = append(results, subResults...)
	}
	return results, nil
}

func (s *splitter) getNextSql(sqlText string) (*sqlWithLineNumber, error) {
	matcheDelimiterCommand, err := s.matchAndSetCustomDelimiter(sqlText)
	if err != nil {
		return nil, err
	}
	// 若匹配到自定义分隔符语法，则输出结果，否则匹配分隔符，输出结果
	if matcheDelimiterCommand || s.matcheSql(sqlText) {
		buff := bytes.Buffer{}
		buff.WriteString(sqlText[:s.scanner.Offset()])
		lineBeforeStart := strings.Count(sqlText[:s.delimiter.startPos], "\n")
		result := &sqlWithLineNumber{
			sql:  strings.TrimSpace(buff.String()),
			line: s.delimiter.line + lineBeforeStart + 1,
		}
		s.delimiter.line += s.scanner.ScannedLines() // pos().Line-1表示的是该SQL中有多少换行
		return result, nil
	}
	return &sqlWithLineNumber{
		sql:  strings.TrimSpace(sqlText),
		line: s.delimiter.line + strings.Count(sqlText[:s.delimiter.startPos], "\n") + 1,
	}, nil
}

func (s *splitter) matcheSql(sql string) bool {
	s.scanner.Reset(sql)
	token := &Token{}
	var isFirstToken bool = true

	for s.scanner.Offset() < len(sql) {
		token = s.scanner.Lex()
		if isFirstToken {
			s.delimiter.startPos = s.scanner.Offset()
			isFirstToken = false
		}
		token = s.skipBeginEndBlock(token)
		if s.isTokenMatchDelimiter(token) {
			return true
		}
	}
	return false
}

func (s *splitter) skipBeginEndBlock(token *Token) *Token {
	var blockStack []Block
	if token.tokenType == begin {
		blockStack = append(blockStack, BeginEndBlock{})
	}
	for len(blockStack) > 0 {
		token = s.scanner.Lex()
		for _, block := range allBlocks {
			if block.MatchBegin(token) {
				blockStack = append(blockStack, block)
				break
			}
		}
		// 如果匹配到END，则需要判断END后的token是否匹配当前的Block
		if token.tokenType == end {
			currentBlock := blockStack[len(blockStack)-1]
			token = s.scanner.Lex()
			if currentBlock.MatchEnd(token) {
				blockStack = blockStack[:len(blockStack)-1]
			}
		}
		if len(blockStack) == 0 {
			break
		}
	}
	return token
}
