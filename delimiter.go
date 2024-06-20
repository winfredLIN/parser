package parser

import (
	"bytes"
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
	Scanner                    *Scanner
	FirstTokenTypeOfDelimiter  int
	FirstTokenValueOfDelimiter string
	DelimiterStr               string
	line                       int
	startPos                   int
}

func NewDelimiter() *Delimiter {
	d := &Delimiter{}
	d.Scanner = NewScanner("")
	return d
}

func (d *Delimiter) SplitSqlText(sqlText string) (results []*sqlWithLineNumber, err error) {
	d.line = 0
	d.startPos = 0
	d.setDelimiter(DefaultDelimiterString)
	return d.splitSqlText(sqlText)
}

func (d *Delimiter) splitSqlText(sqlText string) (results []*sqlWithLineNumber, err error) {
	result, err := d.getNextSql(sqlText)
	if err != nil {
		return nil, err
	}
	results = append(results, result)
	// 递归切分剩余SQL
	if d.Scanner.lastScanOffset < len(sqlText) {
		subResults, _ := d.splitSqlText(sqlText[d.Scanner.lastScanOffset:])
		results = append(results, subResults...)
	}
	return results, nil
}

type sqlWithLineNumber struct {
	sql  string
	line int
}

func (d *Delimiter) getNextSql(sqlText string) (*sqlWithLineNumber, error) {
	matched, err := d.matchAndSetCustomDelimiter(sqlText)
	if err != nil {
		return nil, err
	}
	// 若匹配到自定义分隔符语法，则输出结果，否则匹配分隔符，输出结果
	if matched || (d.matcheDelimiter(sqlText) && d.Scanner.lastScanOffset > 0) {
		buff := bytes.Buffer{}
		buff.WriteString(sqlText[:d.Scanner.lastScanOffset])
		lineBeforeStart := strings.Count(sqlText[:d.startPos], "\n")
		result := &sqlWithLineNumber{
			sql:  strings.TrimSpace(buff.String()),
			line: d.line + lineBeforeStart + 1,
		}
		d.line += d.Scanner.r.pos().Line - 1 // 表示的是该SQL中有多少换行
		return result, nil
	}
	return &sqlWithLineNumber{
		sql:  strings.TrimSpace(sqlText),
		line: d.line + strings.Count(sqlText[:d.startPos], "\n") + 1,
	}, nil
}

/*
该方法检测sql文本开头是否是自定义分隔符语法，若是匹配并更新分隔符:

 1. 分隔符语法满足：delimiter str 或者 \d str
 2. 参考链接：https://dev.mysql.com/doc/refman/5.7/en/mysql-commands.html
*/
func (d *Delimiter) matchAndSetCustomDelimiter(sql string) (bool, error) {
	// 重置扫描器
	token := &yySymType{}
	d.Scanner.reset(sql)
	d.Scanner.lastScanOffset = 0

	var sqlAfterDelimiter string

	switch d.Scanner.Lex(token) {
	case BackSlash:
		if d.isSortDelimiterCommand(sql) {
			sqlAfterDelimiter = sql[d.Scanner.lastScanOffset+2:] // \d的长度是2字节
			d.startPos = d.Scanner.lastScanOffset
			d.Scanner.lastScanOffset += 2
		}
	case identifier:
		if d.isDelimiterCommand(token.ident) {
			sqlAfterDelimiter = sql[d.Scanner.lastScanOffset+9:] //DELIMITER的长度是9字节
			d.startPos = d.Scanner.lastScanOffset
			d.Scanner.lastScanOffset += 9
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
		if err := d.setDelimiter(newDelimiter); err != nil {
			return false, err
		}
		// 若识别到分隔符，则这一整行都为定义分隔符的sql，
		// 例如 delimiter ;; xx 其中;;为分隔符，而xx不产生任何影响，但属于这条语句
		d.Scanner.lastScanOffset += end
		return true, nil
	}
	return false, nil
}

// \\d会被识别为三个token \ \ d 不能使用Lex，Lex可能会跳过空格和注释，因此这里使用字符串匹配
func (d *Delimiter) isSortDelimiterCommand(sql string) bool {
	return d.Scanner.lastScanOffset+2 < len(sql) && sql[d.Scanner.lastScanOffset+1] == 'd'
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

/*
该方法检测分隔符：

	由于scanner会把分隔符扫描为identifier或者其他单字符token类型，因此分为两种情况处理
	注意，若将SQL关键字定义为分隔符，目前未处理该情况
*/
func (d *Delimiter) matcheDelimiter(sql string) bool {

	d.Scanner.reset(sql)
	d.Scanner.lastScanOffset = 0
	token := &yySymType{}
	var isFirstToken bool = true

	for d.Scanner.lastScanOffset < len(sql) {
		tokenType := d.Scanner.Lex(token)
		if isFirstToken {
			d.startPos = d.Scanner.lastScanOffset
			isFirstToken = false
		}
		if d.isTokenMatchDelimiter(tokenType, token) {
			return true
		}
	}
	return false
}

func (d *Delimiter) isTokenMatchDelimiter(tokenType int, token *yySymType) bool {
	switch tokenType {
	case d.FirstTokenTypeOfDelimiter:
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
		if tokenType == stringLit && token.ident != d.FirstTokenValueOfDelimiter {
			return false
		}
		// 2. 定位特征的第一个字符所处的位置
		indexIntoken := strings.Index(token.ident, d.FirstTokenValueOfDelimiter)
		if indexIntoken == -1 {
			return false
		}
		// 3. 字符串匹配
		begin := d.Scanner.lastScanOffset + indexIntoken
		end := begin + len(d.DelimiterStr)
		if begin < 0 || end > len(d.Scanner.r.s) {
			return false
		}
		expected := d.Scanner.r.s[begin:end]
		if expected != d.DelimiterStr {
			return false
		}
		d.Scanner.lastScanOffset = end
		return true

	case invalid:
		if d.Scanner.lastScanOffset == d.Scanner.r.p.Offset {
			d.Scanner.r.inc()
		}
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
