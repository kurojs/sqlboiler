package cmds

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/jinzhu/inflection"
	"github.com/pobri19/sqlboiler/dbdrivers"
)

// generateTemplate generates the template associated to the passed in command name.
func generateTemplate(template *template.Template, data *tplData) ([]byte, error) {
	output, err := processTemplate(template, data)
	if err != nil {
		return nil, fmt.Errorf("Unable to process the template %s for table %s: %s", template.Name(), data.Table.Name, err)
	}

	return output, nil
}

// processTemplate takes a template and returns the output of the template execution.
func processTemplate(t *template.Template, data *tplData) ([]byte, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	output, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return output, nil
}

// plural converts singular words to plural words (eg: person to people)
func plural(name string) string {
	splits := strings.Split(name, "_")
	splits[len(splits)-1] = inflection.Plural(splits[len(splits)-1])
	return strings.Join(splits, "_")
}

// singular converts plural words to singular words (eg: people to person)
func singular(name string) string {
	splits := strings.Split(name, "_")
	splits[len(splits)-1] = inflection.Singular(splits[len(splits)-1])
	return strings.Join(splits, "_")
}

// titleCase changes a snake-case variable name
// into a go styled object variable name of "ColumnName".
// titleCase also fully uppercases "ID" components of names, for example
// "column_name_id" to "ColumnNameID".
func titleCase(name string) string {
	splits := strings.Split(name, "_")

	for i, split := range splits {
		if split == "id" {
			splits[i] = "ID"
			continue
		}

		splits[i] = strings.Title(split)
	}

	return strings.Join(splits, "")
}

// titleCaseSingular changes a snake-case variable name
// to a go styled object variable name of "ColumnName".
// titleCaseSingular also converts the last word in the
// variable name to a singularized version of itself.
func titleCaseSingular(name string) string {
	return titleCase(singular(name))
}

// titleCasePlural changes a snake-case variable name
// to a go styled object variable name of "ColumnName".
// titleCasePlural also converts the last word in the
// variable name to a pluralized version of itself.
func titleCasePlural(name string) string {
	return titleCase(plural(name))
}

// camelCase takes a variable name in the format of "var_name" and converts
// it into a go styled variable name of "varName".
// camelCase also fully uppercases "ID" components of names, for example
// "var_name_id" to "varNameID".
func camelCase(name string) string {
	splits := strings.Split(name, "_")

	for i, split := range splits {
		if split == "id" && i > 0 {
			splits[i] = "ID"
			continue
		}

		if i == 0 {
			continue
		}

		splits[i] = strings.Title(split)
	}

	return strings.Join(splits, "")
}

// camelCaseSingular takes a variable name in the format of "var_name" and converts
// it into a go styled variable name of "varName".
// camelCaseSingular also converts the last word in the
// variable name to a singularized version of itself.
func camelCaseSingular(name string) string {
	return camelCase(singular(name))
}

// camelCasePlural takes a variable name in the format of "var_name" and converts
// it into a go styled variable name of "varName".
// camelCasePlural also converts the last word in the
// variable name to a pluralized version of itself.
func camelCasePlural(name string) string {
	return camelCase(plural(name))
}

// makeDBName takes a table name in the format of "table_name" and a
// column name in the format of "column_name" and returns a name used in the
// `db:""` component of an object in the format of "table_name_column_name"
func makeDBName(tableName, colName string) string {
	return tableName + "_" + colName
}

// updateParamNames takes a []Column and returns a comma seperated
// list of parameter names for the update statement template SET clause.
// eg: col1=$1,col2=$2,col3=$3
// Note: updateParamNames will exclude the PRIMARY KEY column.
func updateParamNames(columns []dbdrivers.Column) string {
	names := make([]string, 0, len(columns))
	counter := 0
	for _, c := range columns {
		if c.IsPrimaryKey {
			continue
		}
		counter++
		names = append(names, fmt.Sprintf("%s=$%d", c.Name, counter))
	}
	return strings.Join(names, ",")
}

// updateParamVariables takes a prefix and a []Columns and returns a
// comma seperated list of parameter variable names for the update statement.
// eg: prefix("o."), column("name_id") -> "o.NameID, ..."
// Note: updateParamVariables will exclude the PRIMARY KEY column.
func updateParamVariables(prefix string, columns []dbdrivers.Column) string {
	names := make([]string, 0, len(columns))

	for _, c := range columns {
		if c.IsPrimaryKey {
			continue
		}
		n := prefix + titleCase(c.Name)
		names = append(names, n)
	}

	return strings.Join(names, ", ")
}

// insertParamNames takes a []Column and returns a comma seperated
// list of parameter names for the insert statement template.
func insertParamNames(columns []dbdrivers.Column) string {
	names := make([]string, 0, len(columns))
	for _, c := range columns {
		names = append(names, c.Name)
	}
	return strings.Join(names, ", ")
}

// insertParamFlags takes a []Column and returns a comma seperated
// list of parameter flags for the insert statement template.
func insertParamFlags(columns []dbdrivers.Column) string {
	params := make([]string, 0, len(columns))
	for i := range columns {
		params = append(params, fmt.Sprintf("$%d", i+1))
	}
	return strings.Join(params, ", ")
}

// insertParamVariables takes a prefix and a []Columns and returns a
// comma seperated list of parameter variable names for the insert statement.
// For example: prefix("o."), column("name_id") -> "o.NameID, ..."
func insertParamVariables(prefix string, columns []dbdrivers.Column) string {
	names := make([]string, 0, len(columns))

	for _, c := range columns {
		n := prefix + titleCase(c.Name)
		names = append(names, n)
	}

	return strings.Join(names, ", ")
}

// selectParamNames takes a []Column and returns a comma seperated
// list of parameter names with for the select statement template.
// It also uses the table name to generate the "AS" part of the statement, for
// example: var_name AS table_name_var_name, ...
func selectParamNames(tableName string, columns []dbdrivers.Column) string {
	selects := make([]string, 0, len(columns))
	for _, c := range columns {
		statement := fmt.Sprintf("%s AS %s", c.Name, makeDBName(tableName, c.Name))
		selects = append(selects, statement)
	}

	return strings.Join(selects, ", ")
}

// scanParamNames takes a []Column and returns a comma seperated
// list of parameter names for use in a db.Scan() call.
func scanParamNames(object string, columns []dbdrivers.Column) string {
	scans := make([]string, 0, len(columns))
	for _, c := range columns {
		statement := fmt.Sprintf("&%s.%s", object, titleCase(c.Name))
		scans = append(scans, statement)
	}

	return strings.Join(scans, ", ")
}

// hasPrimaryKey returns true if one of the columns passed in is a primary key
func hasPrimaryKey(columns []dbdrivers.Column) bool {
	for _, c := range columns {
		if c.IsPrimaryKey {
			return true
		}
	}

	return false
}

// getPrimaryKey returns the primary key column name if one is present
func getPrimaryKey(columns []dbdrivers.Column) string {
	for _, c := range columns {
		if c.IsPrimaryKey {
			return c.Name
		}
	}

	return ""
}
