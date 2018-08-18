package database

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	// Driver for sqlite3
	_ "github.com/mattn/go-sqlite3"

	"github.com/hakasec/japanbot-go/bot/config"
	"github.com/hakasec/japanbot-go/bot/helpers"
)

// DBConnection is an extension of sql.DB
type DBConnection struct {
	*sql.DB

	config *config.DBConfiguration
}

// OpenFromConfig creates a DBConnection from a given DBConfiguration
func OpenFromConfig(config *config.DBConfiguration) (*DBConnection, error) {
	db, err := sql.Open(config.DriverName, config.ConnString)
	if err != nil {
		return nil, err
	}

	return &DBConnection{DB: db, config: config}, nil
}

// CreateTable generates a database table from a given model
func (db *DBConnection) CreateTable(model interface{}) error {
	modelType := reflect.TypeOf(model)
	if modelType.NumField() == 0 {
		return errors.New("This model contains no fields")
	}
	tableName := modelType.Name()

	var builder strings.Builder
	builder.WriteString(
		fmt.Sprintf("CREATE TABLE `%s` (\n", tableName),
	)

	modelFields := getModelFields(modelType)
	for i, field := range modelFields {
		tag := field.Tag
		tagVal := tag.Get("model")
		var (
			isPrimaryKey    bool
			isAutoIncrement bool
			isNotNull       bool
			hasDefaultVal   bool
			defaultVal      string
			sqlType         string
			columnBuilder   strings.Builder
		)
		tagArgs := strings.Split(tagVal, ",")
		if len(tagArgs) <= 0 {
			return errors.New("Model tag has no value")
		}

		columnName := tagArgs[0]
		if columnName == "" {
			return errors.New("Field name in model tag is blank")
		}

		for i, val := range tagArgs[1:] {
			if val == "primarykey" {
				isPrimaryKey = true
			} else if val == "auto" {
				isAutoIncrement = true
			} else if val == "notnull" {
				isNotNull = true
			} else if i == len(tagArgs)-2 {
				defaultVal = val
			}
		}

		sqlType = helpers.GetSQLType(field.Type)
		columnBuilder.WriteString(
			fmt.Sprintf(
				"`%s` %s",
				columnName,
				sqlType,
			),
		)
		if isPrimaryKey {
			columnBuilder.WriteString(" PRIMARY KEY")
		}
		if isAutoIncrement {
			columnBuilder.WriteString(" AUTOINCREMENT")
		}
		if isNotNull {
			columnBuilder.WriteString(" NOT NULL")
		}
		if hasDefaultVal {
			if sqlType == "TEXT" {
				columnBuilder.WriteString(fmt.Sprintf(" DEFAULT '%s'", defaultVal))
			} else {
				columnBuilder.WriteString(fmt.Sprintf(" DEFAULT %s", defaultVal))
			}
		}

		// If last column
		if i+1 == len(modelFields) {
			builder.WriteString(fmt.Sprintf("\t%s\n", columnBuilder.String()))
		} else {
			builder.WriteString(fmt.Sprintf("\t%s,\n", columnBuilder.String()))
		}
	}
	builder.WriteString(");")

	// execute build table query
	db.Exec(builder.String())

	return nil
}

func getModelFields(T reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	for i := 0; i < T.NumField(); i++ {
		field := T.Field(i)
		_, ok := field.Tag.Lookup("model")
		if ok {
			fields = append(fields, field)
		}
	}
	return fields
}
