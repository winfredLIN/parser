package parser

type ScannerForSplitter struct {
	scanner *Scanner
}

func NewScannerForSplitter() *ScannerForSplitter {
	return &ScannerForSplitter{
		scanner: NewScanner(""),
	}
}

func (s *ScannerForSplitter) Reset(sql string) {
	s.scanner.reset(sql)
	s.scanner.lastScanOffset = 0
}

func (s *ScannerForSplitter) Offset() int {
	return s.scanner.lastScanOffset
}

func (s *ScannerForSplitter) SetCursor(offset int) {
	s.scanner.lastScanOffset = offset
}

const (
	Identifier int = identifier
	YyEOFCode  int = yyEOFCode
	YyDefault  int = yyDefault
	IfKwd      int = ifKwd
	CaseKwd    int = caseKwd
	Repeat     int = repeat
	Begin      int = begin
	End        int = end
	StringLit  int = stringLit
	Invalid    int = invalid
)

type TokenValue yySymType

type Token struct {
	tokenType  int
	tokenValue *yySymType
}

func (t Token) Ident() string {
	return t.tokenValue.ident
}

func (t Token) TokenType() int {
	return t.tokenType
}

func (s *ScannerForSplitter) Lex() *Token {
	tokenValue := &yySymType{}
	tokenType := s.scanner.Lex(tokenValue)
	return &Token{
		tokenType:  tokenType,
		tokenValue: tokenValue,
	}
}

func (s *ScannerForSplitter) ScannedLines() int {
	return s.scanner.r.pos().Line - 1
}

func (s *ScannerForSplitter) ScannedText() string {
	return s.scanner.r.s
}

func (s *ScannerForSplitter) HandleInvalid() {
	if s.scanner.lastScanOffset == s.scanner.r.p.Offset {
		s.scanner.r.inc()
	}
}
