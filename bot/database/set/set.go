package set

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

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
	return set.get(valueMap, "", false, entity)
}

// GetAsc gets the first entity in ascending order
func (set *DBSet) GetAsc(valueMap map[string]interface{}, orderBy string, entity interface{}) error {
	return set.get(valueMap, orderBy, false, entity)
}

// GetDesc gets the first entity in descending order
func (set *DBSet) GetDesc(valueMap map[string]interface{}, orderBy string, entity interface{}) error {
	return set.get(valueMap, orderBy, true, entity)
}

func (set *DBSet) get(valueMap map[string]interface{}, orderBy string, desc bool, entity interface{}) error {
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
	if orderBy != "" {
		var orderStatement string
		if desc {
			orderStatement = fmt.Sprintf(
				" ORDER BY `%s` DESC",
				set.fieldMap[orderBy],
			)
		} else {
			orderStatement = fmt.Sprintf(
				" ORDER BY `%s` ASC",
				set.fieldMap[orderBy],
			)
		}
		builder.WriteString(orderStatement)
	}
	builder.WriteString(";")

	rawData := make([]interface{}, len(set.fieldMap))
	dest := make([]interface{}, len(set.fieldMap))
	for i := range rawData {
		dest[i] = &rawData[i]
	}

	row := set.db.QueryRow(builder.String(), values...)
	if err := row.Scan(dest...); err != nil {
		return err
	}

	for i, field := range helpers.GetModelFields(set.t) {
		entityVal := reflect.ValueOf(entity).Elem()
		fieldVal := reflect.ValueOf(rawData[i])

		entityField := entityVal.FieldByName(field.Name)
		switch kind := fieldVal.Kind(); kind {
		case reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64:
			entityField.SetInt(fieldVal.Int())
		case reflect.Array, reflect.Slice:
			if entityField.Kind() == reflect.String {
				entityField.SetString(string(fieldVal.Bytes()))
			} else if entityField.Type().Name() == "Time" {
				parsedTime, err := time.Parse(
					helpers.SqliteDateFormat,
					string(fieldVal.Bytes()),
				)
				if err != nil {
					return err
				}
				entityField.Set(reflect.ValueOf(parsedTime))
			}
		default:
			entityField.Set(fieldVal)
		}
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
	return err
}

// Update will update a given entity
func (set *DBSet) Update(entity interface{}) error {
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
		builder         strings.Builder
		primaryKeyValue interface{}
		primaryKeyName  string
		values          []interface{}
	)
	builder.WriteString(fmt.Sprintf("UPDATE `%s` SET", set.tableName))
	for i, field := range fields {
		fieldVal := modelVal.FieldByName(field.Name).Interface()
		tagArgs := strings.Split(field.Tag.Get("model"), ",")
		if helpers.StringSliceContains(tagArgs, "primarykey") {
			primaryKeyValue = fieldVal
			primaryKeyName = set.fieldMap[field.Name]
		}

		if i+1 == len(fields) {
			builder.WriteString(fmt.Sprintf(" `%s` = ?", tagArgs[0]))
		} else {
			builder.WriteString(fmt.Sprintf(" `%s` = ?,", tagArgs[0]))
		}

		values = append(values, fieldVal)
	}
	builder.WriteString(fmt.Sprintf(" WHERE `%s` = ?;", primaryKeyName))
	values = append(values, primaryKeyValue)

	_, err := set.db.Exec(builder.String(), values...)
	return err
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
			isUnique        bool
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
			} else if val == "unique" {
				isUnique = true
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
		if isUnique {
			columnBuilder.WriteString(" UNIQUE")
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
