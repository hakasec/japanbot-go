package set

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hakasec/japanbot-go/bot/database"
	"github.com/hakasec/japanbot-go/bot/helpers"
)

// DBSet lets you add, get and delete entities of a given model
type DBSet struct {
	t reflect.Type

	// Maps the StructField names to their tagged names
	fieldMap  map[string]string
	tableName string

	db *database.DBConnection
}

// New creates a new DBSet
func New(tableName string, T reflect.Type, db *database.DBConnection) *DBSet {
	set := &DBSet{tableName: tableName, t: T, db: db}
	set.createFieldMap()

	return set
}

// Get will fetch an entity from the database according to the valueMap.
// valueMap should contain the struct field as a string for the key and the value being queried;
// this is done as AND statements
func (set *DBSet) Get(valueMap map[string]interface{}, entity interface{}) error {
	var (
		builder strings.Builder
		values  []interface{}
		first   = true
	)

	if len(valueMap) == 0 {
		return errors.New("valueMap is empty")
	}

	builder.WriteString(fmt.Sprintf("SELECT * FROM `%s` WHERE", set.tableName))
	for k, v := range valueMap {
		fieldName, ok := set.fieldMap[k]
		if ok {
			if first {
				builder.WriteString(fmt.Sprintf(" `%s` = ?", fieldName))
				first = false
			} else {
				builder.WriteString(fmt.Sprintf(" AND `%s` = ?", fieldName))
			}
		}
		values = append(values, v)
	}
	builder.WriteString(";")

	row := set.db.QueryRow(builder.String(), values...)
	if err := row.Scan(entity); err == nil {
		return err
	}

	return nil
}

// Add will add a new entity to the database
func (set *DBSet) Add(entity interface{}) error {
	modelType := reflect.TypeOf(entity).Elem()
	modelVal := reflect.ValueOf(entity).Elem()
	if modelType != set.t {
		return errors.New("The type of the given entity doesn't match DBSet type")
	}

	fields := helpers.GetModelFields(modelType)
	if len(fields) == 0 {
		return errors.New("This model doesn't have any model tags")
	}

	var (
		builder    strings.Builder
		usedFields int
		values     []interface{}
	)
	builder.WriteString(fmt.Sprintf("INSERT INTO `%s` (", set.tableName))
	for i, field := range fields {
		fieldVal := modelVal.FieldByName(field.Name).Interface()
		tagArgs := strings.Split(field.Tag.Get("model"), ",")

		isAutoIncrement := helpers.StringSliceContains(tagArgs, "auto")
		isZeroValue := reflect.Zero(field.Type).Interface() == fieldVal
		if isAutoIncrement && isZeroValue {
			continue
		}
		if i+1 == len(fields) {
			builder.WriteString(fmt.Sprintf("`%s`)", tagArgs[0]))
		} else {
			builder.WriteString(fmt.Sprintf("`%s`,", tagArgs[0]))
		}
		usedFields++

		values = append(values, fieldVal)
	}

	valueArgs := strings.Repeat("?,", usedFields)
	valueArgs = valueArgs[:len(valueArgs)-1]
	builder.WriteString(fmt.Sprintf(" VALUES (%s);", valueArgs))

	_, err := set.db.Exec(builder.String(), values...)
	if err != nil {
		return err
	}

	return nil
}

// Delete will delete the given entity from the database
func (set *DBSet) Delete(entity interface{}) error {
	return errors.New("This function is not implemented yet")
}

// TableName returns the table name of this set
func (set *DBSet) TableName() string {
	return set.tableName
}

// Type returns the type of this set
func (set *DBSet) Type() reflect.Type {
	return set.t
}

// CreateTable generates a database table from a given model
func (set *DBSet) CreateTable() error {
	modelType := set.t
	if modelType.NumField() == 0 {
		return errors.New("This model contains no fields")
	}

	var builder strings.Builder
	builder.WriteString(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n", set.tableName),
	)

	modelFields := helpers.GetModelFields(modelType)
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
	_, err := set.db.Exec(builder.String())
	if err != nil {
		return err
	}

	return nil
}

func (set *DBSet) createFieldMap() {
	set.fieldMap = make(map[string]string)

	fields := helpers.GetModelFields(set.t)
	for _, f := range fields {
		fieldName := strings.Split(f.Tag.Get("model"), ",")[0]
		set.fieldMap[f.Name] = fieldName
	}
}
