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

func (s *ScannerForSplitter) Seek(offset int) {
	s.scanner.lastScanOffset += offset
}

type Token struct {
	tokenType  int
	tokenValue *yySymType
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

func (s *ScannerForSplitter) handleInvalid() {
	if s.scanner.lastScanOffset == s.scanner.r.p.Offset {
		s.scanner.r.inc()
	}
}
