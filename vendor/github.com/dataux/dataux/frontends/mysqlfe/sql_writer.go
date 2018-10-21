package mysqlfe

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strings"

	u "github.com/araddon/gou"

	"github.com/araddon/qlbridge/expr"
	"github.com/araddon/qlbridge/rel"
	"github.com/araddon/qlbridge/schema"
	"github.com/araddon/qlbridge/value"
)

var _ = u.EMPTY

var fr = expr.NewFuncRegistry()

func init() {
	fr.Add("typewriter", &MysqlTypeWriter{})
}

// MysqlTypeWriter Convert a qlbridge value type to a mysql type
type MysqlTypeWriter struct{}

func (m *MysqlTypeWriter) Eval(ctx expr.EvalContext, args []value.Value) (value.Value, bool) {

	if len(args) == 0 {
		return value.NewStringValue(""), false
	}

	out := ""
	switch sv := args[0].(type) {
	case value.StringValue:
		switch vt := value.ValueFromString(sv.Val()); vt {
		case value.NilType:
			out = "niltype"
		case value.ErrorType:
			out = "error"
		case value.UnknownType:
			out = "unknown"
		case value.ValueInterfaceType:
			out = "text"
		case value.NumberType:
			out = "double"
		case value.IntType:
			out = "bigint"
		case value.BoolType:
			out = "tinyint"
		case value.TimeType:
			out = "datetime"
		case value.ByteSliceType:
			out = "text"
		case value.StringType:
			out = "varchar(255)"
		case value.StringsType:
			out = "text"
		case value.MapValueType:
			out = "text"
		case value.MapIntType:
			out = "text"
		case value.MapStringType:
			out = "text"
		case value.MapNumberType:
			out = "text"
		case value.MapBoolType:
			out = "text"
		case value.SliceValueType:
			out = "text"
		case value.StructType:
			out = "text"
		case value.JsonType:
			out = "json"
		default:
			out = "text"
		}
	}
	if out == "" {
		return value.NewStringValue(out), false
	}
	return value.NewStringValue(out), true
}
func (m *MysqlTypeWriter) Validate(n *expr.FuncNode) (expr.EvaluatorFunc, error) {
	return m.Eval, nil
}
func (m *MysqlTypeWriter) Type() value.ValueType { return value.StringType }

func typeToMysql(f *schema.Field) string {
	// char(60)
	// varchar(255)
	// text
	switch f.Type {
	case value.IntType:
		if f.Length == 64 {
			return "bigint"
		} else if f.Length == 0 {
			return "int(32)"
		}
		return fmt.Sprintf("int(%d)", f.Length)
	case value.NumberType:
		if f.Length == 64 {
			return "float"
		} else if f.Length == 0 {
			return "float"
		}
		return "float"
	case value.BoolType:
		return "boolean"
	case value.TimeType:
		return "datetime"
	case value.StringType:
		if f.Length != 0 {
			return fmt.Sprintf("varchar(%d)", f.Length)
		}
		return "varchar(255)"
	}
	return "text"
}
func fieldDescribe(proj *rel.Projection, f *schema.Field) []driver.Value {

	null := "YES"
	if f.NoNulls {
		null = "NO"
	}
	if len(proj.Columns) == 6 {
		//[]string{"Field", "Type",  "Null", "Key", "Default", "Extra"}
		if f.Name == "_id" {
			u.Debugf("nulls? %v", f.NoNulls)
		}
		return []driver.Value{
			f.Name,
			typeToMysql(f),
			null,
			f.Key,
			f.DefaultValue,
			f.Description,
		}
	}
	privileges := ""
	if len(f.Roles) > 0 {
		privileges = fmt.Sprintf("{%s}", strings.Join(f.Roles, ", "))
	}
	//[]string{"Field", "Type", "Collation", "Null", "Key", "Default", "Extra", "Privileges", "Comment"}
	return []driver.Value{
		f.Name,
		typeToMysql(f),
		"", // collation
		null,
		f.Key,
		f.DefaultValue,
		f.Extra,
		privileges,
		f.Description,
	}
}

// Implement Dialect Specific Writers
//     ie, mysql, postgres, cassandra all have different dialects
//     so the Create statements are quite different

// Take a table and make create statement
func TableCreate(tbl *schema.Table) (string, error) {

	w := &bytes.Buffer{}
	fmt.Fprintf(w, "CREATE TABLE `%s` (", tbl.Name)
	for i, fld := range tbl.Fields {
		if i != 0 {
			w.WriteByte(',')
		}
		fmt.Fprint(w, "\n    ")
		writeField(w, fld)
	}
	fmt.Fprint(w, "\n) ENGINE=InnoDB DEFAULT CHARSET=utf8;")
	return w.String(), nil
}
func writeField(w *bytes.Buffer, fld *schema.Field) {
	fmt.Fprintf(w, "`%s` ", fld.Name)
	deflen := fld.Length
	switch fld.Type {
	case value.BoolType:
		fmt.Fprint(w, "tinyint(1) DEFAULT NULL")
	case value.IntType:
		fmt.Fprint(w, "bigint DEFAULT NULL")
	case value.StringType:
		if deflen == 0 {
			deflen = 255
		}
		fmt.Fprintf(w, "varchar(%d) DEFAULT NULL", deflen)
	case value.NumberType:
		fmt.Fprint(w, "float DEFAULT NULL")
	case value.TimeType:
		fmt.Fprint(w, "datetime DEFAULT NULL")
	default:
		fmt.Fprint(w, "text DEFAULT NULL")
	}
	if len(fld.Description) > 0 {
		fmt.Fprintf(w, " COMMENT %q", fld.Description)
	}
}
