package parser_test

import (
	"bytes"
	"testing"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/format"
)

func TestPerfectParse(t *testing.T) {
	parser := parser.New()

	stmt, _, err := parser.PerfectParse("OPTIMIZE TABLE foo;", "", "")
	if err != nil {
		t.Error(err)
		return
	}
	if _, ok := stmt[0].(*ast.UnparsedStmt); !ok {
		t.Errorf("expect stmt type is unparsedStmt, actual is %T", stmt)
		return
	}

	type testCase struct {
		sql    string
		expect []string
	}

	tc := []testCase{
		{
			sql: `SELECT * FROM db1.t1`,
			expect: []string{
				`SELECT * FROM db1.t1`,
			},
		},
		{
			sql: `SELECT * FROM db1.t1;SELECT * FROM db2.t2`,
			expect: []string{
				"SELECT * FROM db1.t1;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "SELECT * FROM db1.t1;OPTIMIZE TABLE foo;SELECT * FROM db2.t2",
			expect: []string{
				"SELECT * FROM db1.t1;",
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "OPTIMIZE TABLE foo;SELECT * FROM db1.t1;SELECT * FROM db2.t2",
			expect: []string{
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db1.t1;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "SELECT * FROM db1.t1;SELECT * FROM db2.t2;OPTIMIZE TABLE foo",
			expect: []string{
				"SELECT * FROM db1.t1;",
				"SELECT * FROM db2.t2;",
				"OPTIMIZE TABLE foo",
			},
		},
		{
			sql: "SELECT FROM db2.t2 where a=\"asd;\"; SELECT * FROM db1.t1;",
			expect: []string{
				"SELECT FROM db2.t2 where a=\"asd;\";",
				" SELECT * FROM db1.t1;",
			},
		},
		{
			sql: "SELECT * FROM db1.t1;OPTIMIZE TABLE foo;OPTIMIZE TABLE foo;SELECT * FROM db2.t2",
			expect: []string{
				"SELECT * FROM db1.t1;",
				"OPTIMIZE TABLE foo;",
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "OPTIMIZE TABLE foo;SELECT * FROM db1.t1;OPTIMIZE TABLE foo;SELECT * FROM db2.t2",
			expect: []string{
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db1.t1;",
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db2.t2",
			},
		},
		{
			sql: "SELECT * FROM db1.t1;OPTIMIZE TABLE foo;SELECT * FROM db2.t2;OPTIMIZE TABLE foo",
			expect: []string{
				"SELECT * FROM db1.t1;",
				"OPTIMIZE TABLE foo;",
				"SELECT * FROM db2.t2;",
				"OPTIMIZE TABLE foo",
			},
		},
		{
			sql: `
CREATE PROCEDURE proc1(OUT s int)
BEGIN
END;
`,
			expect: []string{
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
END;`,
			},
		},
		{
			sql: `
CREATE PROCEDURE proc1(OUT s int)
BEGIN
SELECT COUNT(*)  FROM user;
END;
`,
			expect: []string{
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
SELECT COUNT(*)  FROM user;
END;`,
			},
		},
		{
			sql: `
CREATE PROCEDURE proc1(OUT s int)
BEGIN
SELECT COUNT(*)  FROM user;
SELECT COUNT(*)  FROM user;
END;
`,
			expect: []string{
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
SELECT COUNT(*)  FROM user;
SELECT COUNT(*)  FROM user;
END;`,
			},
		},
		{
			sql: `
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
END;
`,
			expect: []string{
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
END;`,
			},
		},
		{
			sql: `
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
`,
			expect: []string{
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
			},
		},
		{
			sql: `
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
SELECT * FROM db1.t1;
`,
			expect: []string{
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`SELECT * FROM db1.t1;`,
			},
		},
		{
			sql: `
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
SELECT * FROM db1.t1;
`,
			expect: []string{
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`SELECT * FROM db1.t1;`,
			},
		},
		{
			sql: `
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
SELECT * FROM db1.t1;
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;
SELECT * FROM db1.t1;
`,
			expect: []string{
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`SELECT * FROM db1.t1;`,
				`
CREATE PROCEDURE proc1(OUT s int)
BEGIN
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
 SELECT COUNT(*)  FROM user;
END;`,
				`SELECT * FROM db1.t1;`,
			},
		},
		{ // 匹配特殊字符结束
			sql: "select * from  �E",
			expect: []string{
				`select * from  �E`,
			},
		},
		{ // 匹配特殊字符后是;
			sql: "select * from  �E;select * from t1",
			expect: []string{
				`select * from  �E;`,
				"select * from t1",
			},
		},
		{ // 匹配特殊字符在中间
			sql: "select * from  �E where id = 1;select * from  �E ",
			expect: []string{
				`select * from  �E where id = 1;`,
				`select * from  �E `,
			},
		},
		{ // 匹配特殊字符在开头
			sql: " where id = 1;select * from  �E ",
			expect: []string{
				` where id = 1;`,
				`select * from  �E `,
			},
		},
		{ // 匹配特殊字符在SQL开头
			sql: "select * from  �E ; where id = 1",
			expect: []string{
				`select * from  �E ;`,
				` where id = 1`,
			},
		},
		{ // 匹配其他invalid场景
			sql: "@`",
			expect: []string{
				"@`",
			},
		},
		{ // 匹配其他invalid场景
			sql: "@` ;select * from t1",
			expect: []string{
				"@` ;select * from t1",
			},
		},
	}
	for _, c := range tc {
		stmt, _, err := parser.PerfectParse(c.sql, "", "")
		if err != nil {
			t.Error(err)
			return
		}
		if len(c.expect) != len(stmt) {
			t.Errorf("expect sql length is %d, actual is %d, sql is [%s]", len(c.expect), len(stmt), c.sql)
		} else {
			for i, s := range stmt {
				if s.Text() != c.expect[i] {
					t.Errorf("expect sql is [%s], actual is [%s]", c.expect[i], s.Text())
				}
			}
		}
	}
}

func TestCharset(t *testing.T) {
	parser := parser.New()
	type testCase struct {
		sql       string
		formatSQL string
		noError   bool
		errMsg    string
	}

	tc := []testCase{
		{
			sql:       `create table t1(id int, name varchar(255) CHARACTER SET armscii8)`,
			formatSQL: `CREATE TABLE t1 (id INT,name VARCHAR(255) CHARACTER SET ARMSCII8)`,
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255) CHARACTER SET armscii8 COLLATE armscii8_general_ci)`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255) CHARACTER SET ARMSCII8 COLLATE armscii8_general_ci)",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255)) DEFAULT CHARACTER SET armscii8`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255)) DEFAULT CHARACTER SET = ARMSCII8",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255)) DEFAULT CHARACTER SET armscii8 COLLATE greek_general_ci`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255)) DEFAULT CHARACTER SET = ARMSCII8 DEFAULT COLLATE = GREEK_GENERAL_CI",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255)) DEFAULT CHARACTER SET utf8mb3`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255)) DEFAULT CHARACTER SET = UTF8",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255)) DEFAULT CHARACTER SET utf8mb3 COLLATE utf8mb3_bin`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255)) DEFAULT CHARACTER SET = UTF8 DEFAULT COLLATE = UTF8_BIN",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255)) DEFAULT CHARACTER SET utf8 COLLATE utf8mb3_bin`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255)) DEFAULT CHARACTER SET = UTF8 DEFAULT COLLATE = UTF8_BIN",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255) CHARACTER SET utf8mb3)`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255) CHARACTER SET UTF8)",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255) CHARACTER SET utf8mb3 COLLATE cp852_general_ci)`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255) CHARACTER SET UTF8 COLLATE cp852_general_ci)",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255))default character set utf8mb3 COLLATE utf8mb3_unicode_ci;`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255)) DEFAULT CHARACTER SET = UTF8 DEFAULT COLLATE = UTF8_UNICODE_CI",
			noError:   true,
		},
		{
			sql:       `create table t1(id int, name varchar(255))default character set utf8mb3 COLLATE big5_chinese_ci;`,
			formatSQL: "CREATE TABLE t1 (id INT,name VARCHAR(255)) DEFAULT CHARACTER SET = UTF8 DEFAULT COLLATE = BIG5_CHINESE_CI",
			noError:   true,
		},
		{
			sql:     `create table t1(id int, name varchar(255)) DEFAULT CHARACTER SET aaa`,
			noError: false,
			errMsg:  "[parser:1115]Unknown character set: 'aaa'",
		},
		{
			sql:     `create table t1(id int, name varchar(255)) DEFAULT CHARACTER SET utf8mb3 COLLATE bbb`,
			noError: false,
			errMsg:  "[ddl:1273]Unknown collation: 'bbb'",
		},

		// 原生测试用例，预期从报错调整为不报错。
		{
			sql:       `create table t (a longtext unicode);`,
			formatSQL: "CREATE TABLE t (a LONGTEXT CHARACTER SET UCS2)",
			noError:   true,
		},
		{
			sql:       `create table t (a long byte, b text unicode);`,
			formatSQL: "CREATE TABLE t (a MEDIUMTEXT,b TEXT CHARACTER SET UCS2)",
			noError:   true,
		},
		{
			sql:       `create table t (a long ascii, b long unicode);`,
			formatSQL: "CREATE TABLE t (a MEDIUMTEXT CHARACTER SET LATIN1,b MEDIUMTEXT CHARACTER SET UCS2)",
			noError:   true,
		},
		{
			sql:       `create table t (a text unicode, b mediumtext ascii, c int);`,
			formatSQL: "CREATE TABLE t (a TEXT CHARACTER SET UCS2,b MEDIUMTEXT CHARACTER SET LATIN1,c INT)",
			noError:   true,
		},
	}

	for _, c := range tc {
		stmt, err := parser.ParseOneStmt(c.sql, "", "")
		if err != nil {
			if c.noError {
				t.Error(err)
				continue
			}
			if err.Error() != c.errMsg {
				t.Errorf("expect error message: %s; actual error message: %s", c.errMsg, err.Error())
				continue
			}
			continue
		} else {
			if !c.noError {
				t.Errorf("expect need error, but no error")
				continue
			}
			buf := new(bytes.Buffer)
			restoreCtx := format.NewRestoreCtx(format.RestoreKeyWordUppercase, buf)
			err = stmt.Restore(restoreCtx)
			if nil != err {
				t.Error(err)
				continue
			}
			if buf.String() != c.formatSQL {
				t.Errorf("expect sql format: %s; actual sql format: %s", c.formatSQL, buf.String())
			}
		}
	}
}

func TestGeometryColumn(t *testing.T) {
	parser := parser.New()
	type testCase struct {
		sql       string
		formatSQL string
		noError   bool
		errMsg    string
	}

	tc := []testCase{
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,g POINT)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,g POINT)`,
			noError:   true,
		},
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY, g GEOMETRY)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,g GEOMETRY)`,
			noError:   true,
		},
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY, g LINESTRING)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,g LINESTRING)`,
			noError:   true,
		},
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY, g POLYGON)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,g POLYGON)`,
			noError:   true,
		},
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY, g MULTIPOINT)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,g MULTIPOINT)`,
			noError:   true,
		},
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY, g MULTILINESTRING)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,g MULTILINESTRING)`,
			noError:   true,
		},
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY, g MULTIPOLYGON)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,g MULTIPOLYGON)`,
			noError:   true,
		},
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY, g GEOMETRYCOLLECTION)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,g GEOMETRYCOLLECTION)`,
			noError:   true,
		},
		{
			sql:       `ALTER TABLE t ADD COLUMN g GEOMETRY`,
			formatSQL: `ALTER TABLE t ADD COLUMN g GEOMETRY`,
			noError:   true,
		},
		{
			sql:       `ALTER TABLE t ADD COLUMN g POINT`,
			formatSQL: `ALTER TABLE t ADD COLUMN g POINT`,
			noError:   true,
		},
		{
			sql:       `ALTER TABLE t ADD COLUMN g LINESTRING`,
			formatSQL: `ALTER TABLE t ADD COLUMN g LINESTRING`,
			noError:   true,
		},
		{
			sql:       `ALTER TABLE t ADD COLUMN g POLYGON`,
			formatSQL: `ALTER TABLE t ADD COLUMN g POLYGON`,
			noError:   true,
		},
		{
			sql:       `ALTER TABLE t ADD COLUMN g MULTIPOINT`,
			formatSQL: `ALTER TABLE t ADD COLUMN g MULTIPOINT`,
			noError:   true,
		},
		{
			sql:       `ALTER TABLE t ADD COLUMN g MULTILINESTRING`,
			formatSQL: `ALTER TABLE t ADD COLUMN g MULTILINESTRING`,
			noError:   true,
		},
		{
			sql:       `ALTER TABLE t ADD COLUMN g MULTIPOLYGON`,
			formatSQL: `ALTER TABLE t ADD COLUMN g MULTIPOLYGON`,
			noError:   true,
		},
		{
			sql:       `ALTER TABLE t ADD COLUMN g GEOMETRYCOLLECTION`,
			formatSQL: `ALTER TABLE t ADD COLUMN g GEOMETRYCOLLECTION`,
			noError:   true,
		},
	}

	for _, c := range tc {
		stmt, err := parser.ParseOneStmt(c.sql, "", "")
		if err != nil {
			if c.noError {
				t.Error(err)
				continue
			}
			if err.Error() != c.errMsg {
				t.Errorf("expect error message: %s; actual error message: %s", c.errMsg, err.Error())
				continue
			}
			continue
		} else {
			if !c.noError {
				t.Errorf("expect need error, but no error")
				continue
			}
			buf := new(bytes.Buffer)
			restoreCtx := format.NewRestoreCtx(format.RestoreKeyWordUppercase, buf)
			err = stmt.Restore(restoreCtx)
			if nil != err {
				t.Error(err)
				continue
			}
			if buf.String() != c.formatSQL {
				t.Errorf("expect sql format: %s; actual sql format: %s", c.formatSQL, buf.String())
			}
		}
	}
}

func TestIndexConstraint(t *testing.T) {
	parser := parser.New()
	type testCase struct {
		sql             string
		indexConstraint interface{}
	}
	tc := []testCase{
		{
			sql:             "CREATE TABLE t (id INT PRIMARY KEY, g POINT, SPATIAL INDEX(g))",
			indexConstraint: ast.ConstraintSpatial,
		},
		{
			sql:             "ALTER TABLE geom ADD SPATIAL INDEX(g)",
			indexConstraint: ast.ConstraintSpatial,
		},
		{
			sql:             "CREATE SPATIAL INDEX g ON geom (g)",
			indexConstraint: ast.IndexKeyTypeSpatial,
		},
	}

	for _, c := range tc {
		isRight := false
		stmt, err := parser.ParseOneStmt(c.sql, "", "")
		if err != nil {
			t.Error(err)
			continue
		} else {
			switch stmt := stmt.(type) {
			case *ast.CreateTableStmt:
				indexConstraint, ok := c.indexConstraint.(ast.ConstraintType)
				if !ok {
					t.Errorf("sql: %s, indexConstraint is not ConstraintType", c.sql)
				}
				for _, constraint := range stmt.Constraints {
					if constraint.Tp == indexConstraint {
						isRight = true
					}
				}
			case *ast.AlterTableStmt:
				indexConstraint, ok := c.indexConstraint.(ast.ConstraintType)
				if !ok {
					t.Errorf("sql: %s, indexConstraint is not ConstraintType", c.sql)
				}
				for _, spec := range stmt.Specs {
					if spec.Tp != ast.AlterTableAddConstraint || spec.Constraint == nil {
						continue
					}
					if spec.Constraint.Tp == indexConstraint {
						isRight = true
					}
				}
			case *ast.CreateIndexStmt:
				indexKey, ok := c.indexConstraint.(ast.IndexKeyType)
				if !ok {
					t.Errorf("sql: %s, indexConstraint is not indexKey", c.sql)
				}
				if stmt.KeyType == indexKey {
					isRight = true
				}
			}
		}
		if !isRight {
			t.Errorf("sql: %s, do not get expect indexConstraint: %v", c.sql, c.indexConstraint)
		}
	}
}

func TestGeometryColumnIsNotReserved(t *testing.T) {
	parser := parser.New()
	type testCase struct {
		sql       string
		formatSQL string
		noError   bool
		errMsg    string
	}

	tc := []testCase{
		// point
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,point INT(8) NOT NULL)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,point INT(8) NOT NULL)`,
			noError:   true,
		},
		{
			sql:       `SELECT point FROM t`,
			formatSQL: `SELECT point FROM t`,
			noError:   true,
		},
		{
			sql:       `INSERT INTO t (point) VALUES (1)`,
			formatSQL: `INSERT INTO t (point) VALUES (1)`,
			noError:   true,
		},
		{
			sql:       `UPDATE t SET point=1`,
			formatSQL: `UPDATE t SET point=1`,
			noError:   true,
		},
		{
			sql:       `DELETE FROM t WHERE point=1`,
			formatSQL: `DELETE FROM t WHERE point=1`,
			noError:   true,
		},
		// geometry
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,geometry INT(8) NOT NULL)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,geometry INT(8) NOT NULL)`,
			noError:   true,
		},
		{
			sql:       `SELECT geometry FROM t`,
			formatSQL: `SELECT geometry FROM t`,
			noError:   true,
		},
		{
			sql:       `INSERT INTO t (geometry) VALUES (1)`,
			formatSQL: `INSERT INTO t (geometry) VALUES (1)`,
			noError:   true,
		},
		{
			sql:       `UPDATE t SET geometry=1`,
			formatSQL: `UPDATE t SET geometry=1`,
			noError:   true,
		},
		{
			sql:       `DELETE FROM t WHERE geometry=1`,
			formatSQL: `DELETE FROM t WHERE geometry=1`,
			noError:   true,
		},
		// LINESTRING
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,linestring INT(8) NOT NULL)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,linestring INT(8) NOT NULL)`,
			noError:   true,
		},
		{
			sql:       `SELECT linestring FROM t`,
			formatSQL: `SELECT linestring FROM t`,
			noError:   true,
		},
		{
			sql:       `INSERT INTO t (linestring) VALUES (1)`,
			formatSQL: `INSERT INTO t (linestring) VALUES (1)`,
			noError:   true,
		},
		{
			sql:       `UPDATE t SET linestring=1`,
			formatSQL: `UPDATE t SET linestring=1`,
			noError:   true,
		},
		{
			sql:       `DELETE FROM t WHERE linestring=1`,
			formatSQL: `DELETE FROM t WHERE linestring=1`,
			noError:   true,
		},
		// POLYGON
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,polygon INT(8) NOT NULL)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,polygon INT(8) NOT NULL)`,
			noError:   true,
		},
		{
			sql:       `SELECT polygon FROM t`,
			formatSQL: `SELECT polygon FROM t`,
			noError:   true,
		},
		{
			sql:       `INSERT INTO t (polygon) VALUES (1)`,
			formatSQL: `INSERT INTO t (polygon) VALUES (1)`,
			noError:   true,
		},
		{
			sql:       `UPDATE t SET polygon=1`,
			formatSQL: `UPDATE t SET polygon=1`,
			noError:   true,
		},
		{
			sql:       `DELETE FROM t WHERE polygon=1`,
			formatSQL: `DELETE FROM t WHERE polygon=1`,
			noError:   true,
		},
		// MULTIPOINT
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,multipoint INT(8) NOT NULL)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,multipoint INT(8) NOT NULL)`,
			noError:   true,
		},
		{
			sql:       `SELECT multipoint FROM t`,
			formatSQL: `SELECT multipoint FROM t`,
			noError:   true,
		},
		{
			sql:       `INSERT INTO t (multipoint) VALUES (1)`,
			formatSQL: `INSERT INTO t (multipoint) VALUES (1)`,
			noError:   true,
		},
		{
			sql:       `UPDATE t SET multipoint=1`,
			formatSQL: `UPDATE t SET multipoint=1`,
			noError:   true,
		},
		{
			sql:       `DELETE FROM t WHERE multipoint=1`,
			formatSQL: `DELETE FROM t WHERE multipoint=1`,
			noError:   true,
		},
		// MULTILINESTRING
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,multilinestring INT(8) NOT NULL)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,multilinestring INT(8) NOT NULL)`,
			noError:   true,
		},
		{
			sql:       `SELECT multilinestring FROM t`,
			formatSQL: `SELECT multilinestring FROM t`,
			noError:   true,
		},
		{
			sql:       `INSERT INTO t (multilinestring) VALUES (1)`,
			formatSQL: `INSERT INTO t (multilinestring) VALUES (1)`,
			noError:   true,
		},
		{
			sql:       `UPDATE t SET multilinestring=1`,
			formatSQL: `UPDATE t SET multilinestring=1`,
			noError:   true,
		},
		{
			sql:       `DELETE FROM t WHERE multilinestring=1`,
			formatSQL: `DELETE FROM t WHERE multilinestring=1`,
			noError:   true,
		},
		// MULTIPOLYGON
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,multipolygon INT(8) NOT NULL)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,multipolygon INT(8) NOT NULL)`,
			noError:   true,
		},
		{
			sql:       `SELECT multipolygon FROM t`,
			formatSQL: `SELECT multipolygon FROM t`,
			noError:   true,
		},
		{
			sql:       `INSERT INTO t (multipolygon) VALUES (1)`,
			formatSQL: `INSERT INTO t (multipolygon) VALUES (1)`,
			noError:   true,
		},
		{
			sql:       `UPDATE t SET multipolygon=1`,
			formatSQL: `UPDATE t SET multipolygon=1`,
			noError:   true,
		},
		{
			sql:       `DELETE FROM t WHERE multipolygon=1`,
			formatSQL: `DELETE FROM t WHERE multipolygon=1`,
			noError:   true,
		},
		// GEOMETRYCOLLECTION
		{
			sql:       `CREATE TABLE t (id INT PRIMARY KEY,geometrycollection INT(8) NOT NULL)`,
			formatSQL: `CREATE TABLE t (id INT PRIMARY KEY,geometrycollection INT(8) NOT NULL)`,
			noError:   true,
		},
		{
			sql:       `SELECT geometrycollection FROM t`,
			formatSQL: `SELECT geometrycollection FROM t`,
			noError:   true,
		},
		{
			sql:       `INSERT INTO t (geometrycollection) VALUES (1)`,
			formatSQL: `INSERT INTO t (geometrycollection) VALUES (1)`,
			noError:   true,
		},
		{
			sql:       `UPDATE t SET geometrycollection=1`,
			formatSQL: `UPDATE t SET geometrycollection=1`,
			noError:   true,
		},
		{
			sql:       `DELETE FROM t WHERE geometrycollection=1`,
			formatSQL: `DELETE FROM t WHERE geometrycollection=1`,
			noError:   true,
		},
	}

	for _, c := range tc {
		stmt, err := parser.ParseOneStmt(c.sql, "", "")
		if err != nil {
			if c.noError {
				t.Error(err)
				continue
			}
			if err.Error() != c.errMsg {
				t.Errorf("expect error message: %s; actual error message: %s", c.errMsg, err.Error())
				continue
			}
			continue
		} else {
			if !c.noError {
				t.Errorf("expect need error, but no error")
				continue
			}
			buf := new(bytes.Buffer)
			restoreCtx := format.NewRestoreCtx(format.RestoreKeyWordUppercase, buf)
			err = stmt.Restore(restoreCtx)
			if nil != err {
				t.Error(err)
				continue
			}
			if buf.String() != c.formatSQL {
				t.Errorf("expect sql format: %s; actual sql format: %s", c.formatSQL, buf.String())
			}
		}
	}
}
