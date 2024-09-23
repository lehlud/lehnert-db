package ldb

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

type Forwardable interface {
	Forward()
}

type Clonable[T any] interface {
	Clone() T
}

// ensure interface implementation
var _ Forwardable = (*Collection)(nil)
var _ Forwardable = (*Field)(nil)
var _ Clonable[*Collection] = Collection{}
var _ Clonable[*CollectionSchema] = CollectionSchema{}
var _ Clonable[*Field] = Field{}
var _ Clonable[*FieldSchema] = FieldSchema{}
var _ Clonable[FieldType] = (FieldType)(nil)
var _ FieldType = FieldTypeId{}
var _ FieldType = FieldTypeText{}
var _ FieldType = FieldTypeInt{}
var _ FieldType = FieldTypeFloat{}
var _ FieldType = FieldTypeBool{}
var _ FieldType = FieldTypeDateTime{}
var _ FieldType = FieldTypeEnum{}
var _ FieldType = FieldTypeSingleRelation{}

type Collection struct {
	// collection data on last migration; useful for detecting schema changes
	original *Collection

	Name   string
	Schema *CollectionSchema
}

func (c *Collection) Forward() {
	c.original = c.Clone()

	for _, field := range c.Schema.Fields {
		field.Forward()
	}
}

func (c Collection) Clone() *Collection {
	cloned := Collection{}
	cloned.Name = c.Name
	cloned.Schema = c.Schema.Clone()
	return &cloned
}

type CollectionSchema struct {
	Fields      []*Field
	ViewFilter  func() bool
	AllowCreate func() bool
	AllowUpdate func() bool
	AllowDelete func() bool
}

func (s CollectionSchema) Clone() *CollectionSchema {
	cloned := s

	clonedFields := make([]*Field, len(s.Fields))
	for i, field := range s.Fields {
		clonedFields[i] = field.Clone()
	}

	return &cloned
}

type Field struct {
	// field data on last migration; useful for detecting schema changes
	original *Field

	Name   string
	Schema *FieldSchema
}

func (f *Field) Forward() {
	f.original = f.Clone()
}

func (f Field) Clone() *Field {
	cloned := Field{}
	cloned.Name = f.Name
	cloned.Schema = f.Schema.Clone()
	return &cloned
}

type FieldSchema struct {
	Type FieldType
}

func (s FieldSchema) Clone() *FieldSchema {
	cloned := FieldSchema{}
	cloned.Type = s.Type.Clone()
	return &cloned
}

type FieldType interface {
	Clone() FieldType

	// validates if the specified value suits the field type;
	// returns the value either in original or in encoded/decoded/recoded form;
	// returns a comprehensive error message if the value is not suitable
	ValidateValue(value any) (any, error)
}

func validateNullable(nullable bool, value any) error {
	if value == nil && !nullable {
		return fmt.Errorf("invalid value, expected non-null")
	}

	return nil
}

type FieldTypeId struct {
	Nullable           bool
	PrimaryKey         bool
	CreateDefaultValue func() string
}

func (ft FieldTypeId) Clone() FieldType {
	return FieldType(ft)
}

func (fieldType FieldTypeId) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable || fieldType.PrimaryKey, value); err != nil {
		return nil, err
	}

	if value == nil {
		return nil, nil
	}

	if err := ValidateId(value); err != nil {
		return nil, err
	}

	return value, nil
}

type FieldTypeText struct {
	Nullable           bool
	CreateDefaultValue func() string
	CreateMaxLength    func() int
	CreateMinLength    func() int
	CreatePattern      func() string
}

func (ft FieldTypeText) Clone() FieldType {
	return FieldType(ft)
}

func (fieldType FieldTypeText) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		if fieldType.CreateDefaultValue != nil {
			return fieldType.CreateDefaultValue(), nil
		}

		return nil, nil
	}

	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value, expected string")
	}

	if fieldType.CreateMinLength != nil {
		if minLength := fieldType.CreateMinLength(); len(str) < minLength {
			return nil, fmt.Errorf("value too short, min length is %v", minLength)
		}
	}

	if fieldType.CreateMaxLength != nil {
		if maxLength := fieldType.CreateMaxLength(); len(str) > maxLength {
			return nil, fmt.Errorf("value too long, max length is %v", maxLength)
		}
	}

	if fieldType.CreatePattern != nil {
		pattern := fieldType.CreatePattern()
		if _, err := regexp.MatchString(pattern, str); err != nil {
			return nil, fmt.Errorf("value does not match pattern, pattern is %v", pattern)
		}
	}

	return str, nil
}

type FieldTypeInt struct {
	Nullable           bool
	CreateDefaultValue func() int64
	CreateMinValue     func() int64
	CreateMaxValue     func() int64
}

func (ft FieldTypeInt) Clone() FieldType {
	return FieldType(ft)
}

func (fieldType FieldTypeInt) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		if fieldType.CreateDefaultValue != nil {
			return fieldType.CreateDefaultValue(), nil
		}

		return nil, nil
	}

	i, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("invalid value, expected integer")
	}

	if fieldType.CreateMinValue != nil {
		if minValue := fieldType.CreateMinValue(); i < minValue {
			return nil, fmt.Errorf("value too small, min value is %v", minValue)
		}
	}

	if fieldType.CreateMaxValue != nil {
		if maxValue := fieldType.CreateMaxValue(); i > maxValue {
			return nil, fmt.Errorf("value too big, max value is %v", maxValue)
		}
	}

	return i, nil
}

type FieldTypeFloat struct {
	Nullable           bool
	CreateDefaultValue func() float64
	CreateMinValue     func() float64
	CreateMaxValue     func() float64
}

func (ft FieldTypeFloat) Clone() FieldType {
	return FieldType(ft)
}

func (fieldType FieldTypeFloat) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		if fieldType.CreateDefaultValue != nil {
			return fieldType.CreateDefaultValue(), nil
		}

		return nil, nil
	}

	f, ok := value.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value, expected float")
	}

	if fieldType.CreateMinValue != nil {
		if minValue := fieldType.CreateMinValue(); f < minValue {
			return nil, fmt.Errorf("value too small, min value is %v", minValue)
		}
	}

	if fieldType.CreateMaxValue != nil {
		if maxValue := fieldType.CreateMaxValue(); f > maxValue {
			return nil, fmt.Errorf("value too big, max value is %v", maxValue)
		}
	}

	return f, nil
}

type FieldTypeBool struct {
	Nullable           bool
	CreateDefaultValue func() bool
}

func (ft FieldTypeBool) Clone() FieldType {
	return FieldType(ft)
}

func (fieldType FieldTypeBool) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		if fieldType.CreateDefaultValue != nil {
			return fieldType.CreateDefaultValue(), nil
		}

		return nil, nil
	}

	b, ok := value.(bool)
	if !ok {
		return nil, fmt.Errorf("invalid value, expected bool")
	}

	return b, nil
}

type FieldTypeDateTime struct {
	Nullable           bool
	CreateDefaultValue func() time.Time
	CreateMinValue     func() time.Time
	CreateMaxValue     func() time.Time
}

func (ft FieldTypeDateTime) Clone() FieldType {
	return FieldType(ft)
}

func (fieldType FieldTypeDateTime) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		if fieldType.CreateDefaultValue != nil {
			return fieldType.CreateDefaultValue(), nil
		}

		return nil, nil
	}

	const timeFormat = time.RFC3339
	const timeFormatName = "RFC-3339"

	d, ok := value.(time.Time)
	if !ok {
		str, _ := value.(string)

		var err error
		if d, err = time.Parse(timeFormat, str); err != nil {
			return nil, fmt.Errorf("invalid value, expected datetime or %s datetime string", timeFormatName)
		}
	}

	if fieldType.CreateMinValue != nil {
		minValue := fieldType.CreateMinValue()
		if d.Before(minValue) {
			return nil, fmt.Errorf("value too early, min value is %s", d.Format(timeFormat))
		}
	}

	if fieldType.CreateMaxValue != nil {
		maxValue := fieldType.CreateMaxValue()
		if d.After(maxValue) {
			return nil, fmt.Errorf("value too late, max value is %s", d.Format(timeFormat))
		}
	}

	return d, nil
}

type FieldTypeEnum struct {
	Nullable           bool
	EnumValues         []string
	CreateDefaultValue func() string
}

func (ft FieldTypeEnum) Clone() FieldType {
	values := ft.EnumValues
	ft.EnumValues = make([]string, len(values))
	copy(ft.EnumValues, values)
	return FieldType(ft)
}

func (fieldType FieldTypeEnum) ValidateValue(value any) (any, error) {
	var defaultValue string = ""
	if fieldType.CreateDefaultValue != nil {
		defaultValue = fieldType.CreateDefaultValue()
		if !slices.Contains(fieldType.EnumValues, defaultValue) {
			return nil, fmt.Errorf("configuration error, invalid default value")
		}
	}

	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		if len(defaultValue) > 0 {
			return defaultValue, nil
		}

		return nil, nil
	}

	str, ok := value.(string)
	if !ok || !slices.Contains(fieldType.EnumValues, str) {
		return nil, fmt.Errorf("invalid value, expected one of [%s]", strings.Join(fieldType.EnumValues, ", "))
	}

	return str, nil
}

type FieldTypeSingleRelation struct {
	Nullable      bool
	Collection    string
	CascadeDelete bool
}

func (ft FieldTypeSingleRelation) Clone() FieldType {
	return FieldType(ft)
}

func (fieldType FieldTypeSingleRelation) ValidateValue(value any) (any, error) {
	idType := FieldTypeId{Nullable: fieldType.Nullable}
	return idType.ValidateValue(value)
}

type View struct {
	// collection name on last migration; empty for newly created collections;
	// useful for detecting when a collection has been renamed
	originalName string

	Name   string
	Schema ViewSchema
}

type ViewSchema struct{}
