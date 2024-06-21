package parser

import (
	"errors"
	"strings"
)

const (
	BackSlash              int    = '\\'
	BackSlashString        string = "\\"
	BlankSpace             string = " "
	DefaultDelimiterString string = ";"
	DelimiterCommand       string = "DELIMITER"
	DelimiterCommandSort   string = `\d`
)

type Delimiter struct {
	FirstTokenTypeOfDelimiter  int
	FirstTokenValueOfDelimiter string
	DelimiterStr               string
	line                       int
	startPos                   int
}

func NewDelimiter() *Delimiter {
	return &Delimiter{}
}

/*
该方法检测sql文本开头是否是自定义分隔符语法，若是匹配并更新分隔符:

 1. 分隔符语法满足：delimiter str 或者 \d str
 2. 参考链接：https://dev.mysql.com/doc/refman/5.7/en/mysql-commands.html
*/
func (s *splitter) matchAndSetCustomDelimiter(sql string) (bool, error) {
	// 重置扫描器
	s.scanner.Reset(sql)

	var sqlAfterDelimiter string
	token := s.scanner.Lex()
	switch token.tokenType {
	case BackSlash:
		if s.delimiter.isSortDelimiterCommand(sql, s.scanner.Offset()) {
			sqlAfterDelimiter = sql[s.scanner.Offset()+2:] // \d的长度是2字节
			s.delimiter.startPos = s.scanner.Offset()
			s.scanner.Seek(2)
		}
	case identifier:
		if s.delimiter.isDelimiterCommand(token.tokenValue.ident) {
			sqlAfterDelimiter = sql[s.scanner.Offset()+9:] //DELIMITER的长度是9字节
			s.delimiter.startPos = s.scanner.Offset()
			s.scanner.Seek(9)
		}
	default:
		return false, nil
	}
	// 处理自定义分隔符
	if sqlAfterDelimiter != "" {
		end := strings.Index(sqlAfterDelimiter, "\n")
		if end == -1 {
			end = len(sqlAfterDelimiter)
		}
		newDelimiter := getDelimiter(sqlAfterDelimiter[:end])
		if err := s.delimiter.setDelimiter(newDelimiter); err != nil {
			return false, err
		}
		// 若识别到分隔符，则这一整行都为定义分隔符的sql，
		// 例如 delimiter ;; xx 其中;;为分隔符，而xx不产生任何影响，但属于这条语句
		s.scanner.Seek(end)
		return true, nil
	}
	return false, nil
}

// \\d会被识别为三个token \ \ d 不能使用Lex，Lex可能会跳过空格和注释，因此这里使用字符串匹配
func (d *Delimiter) isSortDelimiterCommand(sql string, index int) bool {
	return index+2 < len(sql) && sql[index+1] == 'd'
}

// DELIMITER会被识别为identifier，因此这里仅需识别其值是否相等
func (d *Delimiter) isDelimiterCommand(token string) bool {
	return strings.ToUpper(token) == DelimiterCommand
}

// 该函数翻译自MySQL Client获取delimiter值的代码，参考：https://github.com/mysql/mysql-server/blob/824e2b4064053f7daf17d7f3f84b7a3ed92e5fb4/client/mysql.cc#L4866
func getDelimiter(line string) string {
	ptr := 0
	start := 0
	quoted := false
	qtype := byte(0)

	// 跳过开头的空格
	for ptr < len(line) && isSpace(line[ptr]) {
		ptr++
	}

	if ptr == len(line) {
		return ""
	}

	// 检查是否为引号字符串
	if line[ptr] == '\'' || line[ptr] == '"' || line[ptr] == '`' {
		qtype = line[ptr]
		quoted = true
		ptr++
	}

	start = ptr

	// 找到字符串结尾
	for ptr < len(line) {
		if !quoted && line[ptr] == '\\' && ptr+1 < len(line) { // 跳过转义字符
			ptr += 2
		} else if (!quoted && isSpace(line[ptr])) || (quoted && line[ptr] == qtype) {
			break
		} else {
			ptr++
		}
	}

	return line[start:ptr]
}

// 辅助函数,判断字符是否为空格
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// ref:https://dev.mysql.com/doc/refman/8.4/en/flow-control-statements.html
func (s *splitter) isTokenMatchDelimiter(token *Token) bool {
	switch token.tokenType {
	case s.delimiter.FirstTokenTypeOfDelimiter:
		/*
			在mysql client的语法中需要跳过注释以及分隔符处于引号中的情况，由于scanner.Lex会自动跳过注释，因此，仅需要判断分隔符处于引号中的情况。对于该方法，以分隔符的第一个token作为特征仅需匹配，可能会匹配到由引号括起的情况，存在stringLit和identifier两种token需要进一步判断：
				1. 当匹配到identifier时，identifier有可能由反引号括起:
					1. 若identifier没有反引号括起，则不需要判断是否跳过
					2. 若identifier被反引号括起，匹配的字符串会带上反引号，能在匹配字符串时能够检查出是否需要跳过
				2. 当匹配到stringLit时，stringLit一定是由单引号或双引号括起:
					1. 当分隔符第一个token值与stringLit的token值不等，那么一定不是分隔符，则跳过
					2. 当分隔符第一个token值与stringLit的token值相等， 如："'abc'd" '"abc"d'会因为字符串不匹配而跳过
		*/
		// 1. 当分隔符第一个token值与stringLit的token值不等，那么一定不是分隔符，则跳过
		if token.tokenType == stringLit && token.tokenValue.ident != s.delimiter.FirstTokenValueOfDelimiter {
			return false
		}
		// 2. 定位特征的第一个字符所处的位置
		indexIntoken := strings.Index(token.tokenValue.ident, s.delimiter.FirstTokenValueOfDelimiter)
		if indexIntoken == -1 {
			return false
		}
		// 3. 字符串匹配
		begin := s.scanner.Offset() + indexIntoken
		end := begin + len(s.delimiter.DelimiterStr)
		if begin < 0 || end > len(s.scanner.ScannedText()) {
			return false
		}
		expected := s.scanner.ScannedText()[begin:end]
		if expected != s.delimiter.DelimiterStr {
			return false
		}
		s.scanner.Seek(end)
		return true

	case invalid:
		s.scanner.handleInvalid()
	}
	return false
}

var ErrDelimiterCanNotExtractToken = errors.New("sorry, we cannot extract any token form the delimiter you provide, please change a delimiter")
var ErrDelimiterContainsBackslash = errors.New("DELIMITER cannot contain a backslash character")
var ErrDelimiterContainsBlankSpace = errors.New("DELIMITER should not contain blank space")
var ErrDelimiterMissing = errors.New("DELIMITER must be followed by a 'delimiter' character or string")
var ErrDelimiterReservedKeyword = errors.New("delimiter should not use a reserved keyword")

/*
该方法设置分隔符，对分隔符的内容有一定的限制：

 1. 不允许分隔符内部包含反斜杠
 2. 不允许分隔符为空字符串
 3. 不允许分隔符为mysql的保留字，因为这样会被scanner扫描为其他类型的token，从而绕过判断分隔符的逻辑

注：其中1和2与MySQL客户端对分隔符内容一致，错误内容参考MySQL客户端源码中的com_delimiter函数
https://github.com/mysql/mysql-server/blob/824e2b4064053f7daf17d7f3f84b7a3ed92e5fb4/client/mysql.cc#L4621
*/
func (d *Delimiter) setDelimiter(delimiter string) (err error) {
	if delimiter == "" {
		return ErrDelimiterMissing
	}
	if strings.Contains(delimiter, BackSlashString) {
		return ErrDelimiterContainsBackslash
	}
	if strings.Contains(delimiter, BlankSpace) {
		return ErrDelimiterContainsBlankSpace
	}
	if isReservedKeyWord(delimiter) {
		return ErrDelimiterReservedKeyword
	}
	token := &yySymType{}
	d.FirstTokenTypeOfDelimiter = NewScanner(delimiter).Lex(token)
	if d.FirstTokenTypeOfDelimiter == 0 {
		return ErrDelimiterCanNotExtractToken
	}
	d.FirstTokenValueOfDelimiter = token.ident
	d.DelimiterStr = delimiter
	return nil
}

func isReservedKeyWord(input string) bool {
	s := NewScanner(input)
	var token *yySymType = &yySymType{}
	tokenType := s.Lex(token)
	if len(token.ident) < len(input) {
		// 如果分隔符无法识别为一个token，则一定不是关键字
		return false
	}
	// 如果分隔符识别为一个关键字，但不知道是哪个关键字，则为identifier，此时就非保留字
	return tokenType != identifier && tokenType > yyEOFCode && tokenType < yyDefault
}

func (d *Delimiter) reset() {
	d.line = 0
	d.startPos = 0
	d.setDelimiter(DefaultDelimiterString)
}
