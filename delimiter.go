package parser

import (
	"errors"
	"regexp"
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
	Scanner        *Scanner
	DelimiterStr   string
	DelimiterBytes []byte
}

func NewDelimiter() *Delimiter {
	d := &Delimiter{}
	d.setDelimiter(DefaultDelimiterString)
	d.Scanner = NewScanner("")
	return d
}

func (d *Delimiter) ScanNextEndOfSql(sql string) (endOffset int, err error) {
	d.Scanner.reset(sql)
	var token *yySymType = &yySymType{}
	var tokenType int

	for d.Scanner.lastScanOffset < len(sql) {
		// 扫描下一个token
		tokenType = d.Scanner.Lex(token)
		// 当token无效且扫描偏移量未变时，增加偏移量
		if tokenType == invalid && d.Scanner.lastScanOffset == d.Scanner.r.p.Offset {
			d.Scanner.r.inc()
			continue
		}
		if err := d.detectAndSetCustomDelimiter(token, tokenType); err != nil {
			return 0, err
		}
		if d.matchedDelimiter(token, tokenType) {
			return d.Scanner.lastScanOffset, nil
		}
	}

	return d.Scanner.lastScanOffset, nil
}

/*
该方法检测自定义分隔符语法并更新分隔符:

 1. 分隔符语法满足：delimiter str 或者 \d str
 2. 参考链接：https://dev.mysql.com/doc/refman/5.7/en/mysql-commands.html
*/
func (d *Delimiter) detectAndSetCustomDelimiter(token *yySymType, tokenType int) error {
	switch tokenType {
	case BackSlash:
		// 如果token是反斜杠，尝试匹配简短分隔符命令，并设置自定义分隔符
		if matched, customDelimiter := matchDelimiterCommandSort(d.Scanner.r.s); matched {
			if err := d.setDelimiter(customDelimiter); err != nil {
				return err
			}
		}
	case identifier:
		// 如果token是DELIMITER，尝试匹配常规分隔符命令，并设置自定义分隔符
		if strings.ToUpper(token.ident) == DelimiterCommand {
			if matched, customDelimiter := matchDelimiterCommand(d.Scanner.r.s); matched {
				if err := d.setDelimiter(customDelimiter); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// 匹配分隔符定义语法 DELIMITER str
func matchDelimiterCommand(input string) (isMached bool, delimiter string) {
	/*
		1. (?i) 表示无视大小写
		2. \s* 匹配任意数量的空白字符(空格、制表符、换行符等)
		3. DELIMITER 用于匹配字符串DELIMITER，即定义分隔符的常规语法
		4. \x20+ 表示匹配1至多个空格
		5. (\S+)：这是一个捕获组。\S匹配任意非空白字符，+表示匹配前面的模式一次或多次
	*/
	re := regexp.MustCompile(`(?i)\s*DELIMITER\x20+` + `((?:"(.*)"|'(.*)'|` + "`(.*)`" + `))`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return true, matches[1]
	}
	re = regexp.MustCompile(`(?i)\s*DELIMITER\x20+(\S+)`)
	matches = re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return true, matches[1]
	}
	return false, ""
}

// 匹配分隔符定义语法 \d str
func matchDelimiterCommandSort(input string) (isMached bool, delimiter string) {
	/*
		1. \s*：匹配任意数量的空白字符(空格、制表符、换行符等)
		2. regexp.QuoteMeta("\\d") 用于匹配字符串\d，即定义分隔符的简短语法
		3. \x20+ 表示匹配1至多个空格
		4. (\S+)：这是一个捕获组。\S匹配任意非空白字符，+表示匹配前面的模式一次或多次
	*/
	re := regexp.MustCompile(`\s*` + regexp.QuoteMeta(DelimiterCommandSort) + `\x20+` + `((?:"(.*)"|'(.*)'|` + "`(.*)`" + `))`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return true, matches[1]
	}
	re = regexp.MustCompile(`\s*` + regexp.QuoteMeta(DelimiterCommandSort) + `\x20+(\S+)`)
	matches = re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return true, matches[1]
	}
	return false, ""
}

var ErrDelimiterIsCommentStyle = errors.New("please do not use c-style comment as delimiter")
var ErrDelimiterContainsBackslash = errors.New("DELIMITER cannot contain a backslash character")
var ErrDelimiterContainsBlankSpace = errors.New("DELIMITER should not contain blank space")
var ErrDelimiterMissing = errors.New("DELIMITER must be followed by a 'delimiter' character or string")
var ErrDelimiterReservedKeyword = errors.New("delimiter should not use a reserved keyword")

/*
该方法设置分隔符，对分隔符的内容有一定的限制：

 1. 不允许分隔符内部包含反斜杠
 2. 不允许分隔符为空字符串
 3. 不允许分隔符为C语言风格的注释，因为scanner在扫描token的时候会跳过注释内容，处理情况复杂
 4. 不允许分隔符为mysql的保留字，因为这样会被scanner扫描为其他类型的token，从而绕过判断分隔符的逻辑

注：其中1和2与MySQL客户端对分隔符内容一致，错误内容参考MySQL客户端源码中的com_delimiter函数
https://github.com/mysql/mysql-server/blob/824e2b4064053f7daf17d7f3f84b7a3ed92e5fb4/client/mysql.cc#L4621
*/
func (d *Delimiter) setDelimiter(delimiter string) (err error) {

	if isCommentLikeC(delimiter) {
		return ErrDelimiterIsCommentStyle
	}
	if strings.Contains(delimiter, BackSlashString) {
		return ErrDelimiterContainsBackslash
	}
	if strings.Contains(delimiter, BlankSpace) {
		return ErrDelimiterContainsBlankSpace
	}
	if delimiter = removeOuterQuotes(delimiter); delimiter == "" {
		return ErrDelimiterMissing
	}
	if isReservedKeyWord(delimiter) {
		return ErrDelimiterReservedKeyword
	}

	d.DelimiterStr = delimiter
	d.DelimiterBytes = []byte(delimiter)
	return nil
}

func isCommentLikeC(str string) bool {
	re := regexp.MustCompile(`^\/\*[\s\S]*?\*\/$`)
	return re.MatchString(str)
}

// 定义分隔符的时候如果使用引号将分隔符进行包裹，则需要自动去掉一层引号
func removeOuterQuotes(s string) string {
	// 匹配单引号
	singleQuoteRegex := regexp.MustCompile(`^'(.*)'$`)
	matches := singleQuoteRegex.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1]
	}
	// 匹配双引号
	doubleQuoteRegex := regexp.MustCompile(`^"(.*)"$`)
	matches = doubleQuoteRegex.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1]
	}
	// 匹配反引号
	backTickRegex := regexp.MustCompile("`(.*)`")
	matches = backTickRegex.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1]
	}
	return s
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

func (d *Delimiter) delimiter() string {
	return d.DelimiterStr
}

/*
该方法检测分隔符：

	由于scanner会把分隔符扫描为identifier或者其他单字符token类型，因此分为两种情况处理
	注意，若将SQL关键字定义为分隔符，目前未处理该情况
*/
func (d *Delimiter) matchedDelimiter(token *yySymType, tokenType int) bool {
	switch tokenType {
	case identifier:
		// 当token是当前分隔符时，更新扫描偏移量并返回true
		if strings.Contains(token.ident, d.delimiter()) {
			d.Scanner.lastScanOffset += len(d.delimiter()) + strings.Index(token.ident, d.DelimiterStr)
			return true
		}
	case d.firstAsciiValueOfDelimiter():
		// 检查当前扫描位置是否匹配当前分隔符的第一个字符
		expectedEnd := d.Scanner.lastScanOffset + len(d.delimiter())
		if expectedEnd >= len(d.Scanner.r.s) {
			return false
		}
		if d.Scanner.r.s[d.Scanner.lastScanOffset:expectedEnd] == d.delimiter() {
			d.Scanner.lastScanOffset = expectedEnd
			return true
		}
	}
	return false
}

func (d *Delimiter) firstAsciiValueOfDelimiter() int {
	if len(d.DelimiterBytes) > 0 {
		return int(d.DelimiterBytes[0])
	}
	return -1
}
