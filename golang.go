// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refactor

import (
	"errors"
	"fmt"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"time"

	"gitee.com/azhai/xorm-refactor/config"
	"gitee.com/azhai/xorm-refactor/rewrite"
	"github.com/azhai/gozzo-utils/filesystem"
	"xorm.io/xorm/schemas"
)

const FIXED_STR_MAX_SIZE = 255 // 固定字符串最大长度

var (
	TypeOfTime = reflect.TypeOf(time.Time{})

	Bit       = "BIT"
	TinyInt   = "TINYINT"
	SmallInt  = "SMALLINT"
	MediumInt = "MEDIUMINT"
	Int       = "INT"
	Integer   = "INTEGER"
	BigInt    = "BIGINT"

	Enum = "ENUM"
	Set  = "SET"

	Char             = "CHAR"
	Varchar          = "VARCHAR"
	NChar            = "NCHAR"
	NVarchar         = "NVARCHAR"
	TinyText         = "TINYTEXT"
	Text             = "TEXT"
	NText            = "NTEXT"
	Clob             = "CLOB"
	MediumText       = "MEDIUMTEXT"
	LongText         = "LONGTEXT"
	Uuid             = "UUID"
	UniqueIdentifier = "UNIQUEIDENTIFIER"
	SysName          = "SYSNAME"

	Date          = "DATE"
	DateTime      = "DATETIME"
	SmallDateTime = "SMALLDATETIME"
	Time          = "TIME"
	TimeStamp     = "TIMESTAMP"
	TimeStampz    = "TIMESTAMPZ"
	Year          = "YEAR"

	Decimal    = "DECIMAL"
	Numeric    = "NUMERIC"
	Money      = "MONEY"
	SmallMoney = "SMALLMONEY"

	Real   = "REAL"
	Float  = "FLOAT"
	Double = "DOUBLE"

	Binary     = "BINARY"
	VarBinary  = "VARBINARY"
	TinyBlob   = "TINYBLOB"
	Blob       = "BLOB"
	MediumBlob = "MEDIUMBLOB"
	LongBlob   = "LONGBLOB"
	Bytea      = "BYTEA"

	Bool    = "BOOL"
	Boolean = "BOOLEAN"

	Serial    = "SERIAL"
	BigSerial = "BIGSERIAL"

	Json  = "JSON"
	Jsonb = "JSONB"

	Array = "ARRAY"
)

// Golang represents a golang language
var Golang = Language{
	Name:     "golang",
	Template: golangModelTemplate,
	Types:    map[string]string{},
	Funcs: template.FuncMap{
		"Type": type2string,
		"Tag":  tag2string,
	},
	Formatter: rewrite.CleanImportsWriteGolangFile,
	Importter: genGoImports,
	Packager:  genNameSpace,
	ExtName:   ".go",
}

func init() {
	RegisterLanguage(&Golang)
}

var (
	errBadComparisonType = errors.New("invalid type for comparison")
	errBadComparison     = errors.New("incompatible types for comparison")
	errNoComparison      = errors.New("missing argument for comparison")
)

type kind int

const (
	invalidKind kind = iota
	boolKind
	complexKind
	intKind
	floatKind
	integerKind
	stringKind
	uintKind
)

func basicKind(v reflect.Value) (kind, error) {
	switch v.Kind() {
	case reflect.Bool:
		return boolKind, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind, nil
	case reflect.Float32, reflect.Float64:
		return floatKind, nil
	case reflect.Complex64, reflect.Complex128:
		return complexKind, nil
	case reflect.String:
		return stringKind, nil
	}
	return invalidKind, errBadComparisonType
}

func getCol(cols map[string]*schemas.Column, name string) *schemas.Column {
	return cols[strings.ToLower(name)]
}

func genNameSpace(targetDir string) string {
	// 先重试提取已有代码文件（排除测试代码）的包名
	files, err := filesystem.FindFiles(targetDir, ".go")
	if err == nil && len(files) > 0 {
		for fileName := range files {
			if strings.HasSuffix(fileName, "_test.go") {
				continue
			}
			cp, err := rewrite.NewFileParser(fileName)
			if err != nil {
				continue
			}
			if nameSpace := cp.GetPackage(); nameSpace != "" {
				return nameSpace
			}
		}
	}
	// 否则直接使用目录名，需要排除Golang关键词
	nameSpace := strings.ToLower(filepath.Base(targetDir))
	if nameSpace == "default" { // it is golang keyword
		nameSpace = "db"
	} else if token.IsKeyword(nameSpace) {
		nameSpace = "db" + nameSpace
	}
	return nameSpace
}

func genGoImports(tables map[string]*schemas.Table) map[string]string {
	imports := make(map[string]string)
	for _, table := range tables {
		for _, col := range table.Columns() {
			s := type2string(col)
			if s == "time.Time" {
				imports["time"] = ""
			} else if s == "sql.NullString" {
				// imports["database/sql"] = ""
			}
		}
	}
	return imports
}

func type2string(col *schemas.Column) string {
	_, s := SQLType2Type(col.SQLType)
	return s
}

func tag2string(table *schemas.Table, col *schemas.Column, genJson bool) string {
	tj, tx := "", tagXorm(table, col)
	if genJson {
		tj = tagJson(col)
	} else {
		return tx
	}
	if tx == "" {
		if tj == "" {
			return ""
		} else {
			return tj
		}
	}
	return tj + " " + tx
}

func tagJson(col *schemas.Column) string {
	if col.Name == "" {
		return ""
	}
	return fmt.Sprintf(`json:"%s"`, col.Name)
}

func tagXorm(table *schemas.Table, col *schemas.Column) string {
	isNameId := col.FieldName == "Id"
	isIdPk := isNameId && type2string(col) == "int64"

	var res []string
	if !col.Nullable {
		if !isIdPk {
			res = append(res, config.XORM_TAG_NOT_NULL)
		}
	}
	if col.IsPrimaryKey {
		res = append(res, config.XORM_TAG_PRIMARY_KEY)
	}
	if col.Default != "" {
		res = append(res, "default "+col.Default)
	}
	if col.IsAutoIncrement {
		res = append(res, config.XORM_TAG_AUTO_INCR)
	}

	if col.SQLType.IsTime() {
		lowerName := strings.ToLower(col.Name)
		if strings.HasPrefix(lowerName, "created") {
			res = append(res, "created")
		} else if strings.HasPrefix(lowerName, "updated") {
			res = append(res, "updated")
		} else if strings.HasPrefix(lowerName, "deleted") {
			res = append(res, "deleted")
		}
	}

	if col.Comment != "" {
		res = append(res, fmt.Sprintf("comment('%s')", col.Comment))
	}

	names := make([]string, 0, len(col.Indexes))
	for name := range col.Indexes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		index := table.Indexes[name]
		var uistr string
		if index.Type == schemas.UniqueType {
			uistr = config.XORM_TAG_UNIQUE
		} else if index.Type == schemas.IndexType {
			uistr = config.XORM_TAG_INDEX
		}
		if len(index.Cols) > 1 {
			uistr += "(" + index.Name + ")"
		}
		res = append(res, uistr)
	}

	res = append(res, GetColTypeString(col))
	if len(res) > 0 {
		return fmt.Sprintf(`%s:"%s"`, config.XORM_TAG_NAME, strings.Join(res, " "))
	}
	return ""
}

// default sql type change to go types
func SQLType2Type(st schemas.SQLType) (rtype reflect.Type, rtstr string) {
	name := strings.ToUpper(st.Name)
	rtype = reflect.TypeOf("")
	switch name {
	case Bool:
		rtype = reflect.TypeOf(false)
	case Bit, TinyInt:
		if st.DefaultLength == 1 {
			rtype = reflect.TypeOf(false)
		} else {
			rtype = reflect.TypeOf(1)
		}
	case SmallInt, MediumInt, Int, Integer, Serial:
		rtype = reflect.TypeOf(1)
	case BigInt, BigSerial:
		rtype = reflect.TypeOf(int64(1))
	case Float, Real:
		rtype = reflect.TypeOf(float32(1))
	case Double:
		rtype = reflect.TypeOf(float64(1))
	case DateTime, Date, Time, TimeStamp, TimeStampz, SmallDateTime, Year:
		rtype = TypeOfTime
	case TinyBlob, Blob, MediumBlob, LongBlob, Bytea, Binary, VarBinary, UniqueIdentifier:
		rtype, rtstr = reflect.TypeOf([]byte{}), "[]byte"
	case Varchar, NVarchar, TinyText, Text, NText, MediumText, LongText:
		if st.DefaultLength == 0 || st.DefaultLength > FIXED_STR_MAX_SIZE {
			rtstr = "sql.NullString"
		}
		// case Char, NChar, Enum, Set, Uuid, Clob, SysName:
		//	rtstr = rtype.String()
		// case Decimal, Numeric, Money, SmallMoney:
		//	rtstr = rtype.String()
	}
	if rtstr == "" {
		rtstr = rtype.String()
	}
	return
}

// get the col type include length, for example: VARCHAR(255)
func GetColTypeString(col *schemas.Column) string {
	ctstr := col.SQLType.Name
	if col.Length != 0 {
		if col.Length2 != 0 { // float, decimal
			ctstr += fmt.Sprintf("(%v,%v)", col.Length, col.Length2)
		} else { // int, char, varchar
			ctstr += fmt.Sprintf("(%v)", col.Length)
		}
	} else if len(col.EnumOptions) > 0 { // enum
		ctstr += "("
		opts := ""

		enumOptions := make([]string, 0, len(col.EnumOptions))
		for enumOption := range col.EnumOptions {
			enumOptions = append(enumOptions, enumOption)
		}
		sort.Strings(enumOptions)

		for _, v := range enumOptions {
			opts += fmt.Sprintf(",'%v'", v)
		}
		ctstr += strings.TrimLeft(opts, ",")
		ctstr += ")"
	} else if len(col.SetOptions) > 0 { // set
		ctstr += "("
		opts := ""

		setOptions := make([]string, 0, len(col.SetOptions))
		for setOption := range col.SetOptions {
			setOptions = append(setOptions, setOption)
		}
		sort.Strings(setOptions)

		for _, v := range setOptions {
			opts += fmt.Sprintf(",'%v'", v)
		}
		ctstr += strings.TrimLeft(opts, ",")
		ctstr += ")"
	}
	return ctstr
}
