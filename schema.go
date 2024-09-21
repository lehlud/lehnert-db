package ldb

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

type CollectionSchema struct {
	Name   string
	Fields []*SchemaField
}

type SchemaField struct {
	Name string
	Type SchemaFieldType
}

type SchemaFieldType interface {
	GetName() string
	ValidateValue(value any) (any, error)
}

func validateNullable(nullable bool, value any) error {
	if value == nil && !nullable {
		return fmt.Errorf("invalid value, expected non-null")
	}

	return nil
}

type TextFieldType struct {
	Nullable     bool
	DefaultValue *string
	MaxLength    *int
	MinLength    *int
	Pattern      *string
}

func (fieldType TextFieldType) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		return fieldType.DefaultValue, nil
	}

	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid value, expected string")
	}

	if fieldType.MaxLength != nil && len(str) > *fieldType.MaxLength {
		return nil, fmt.Errorf("value too long, max length is %v", *fieldType.MaxLength)
	}

	if fieldType.MinLength != nil && len(str) < *fieldType.MaxLength {
		return nil, fmt.Errorf("value too short, min length is %v", *fieldType.MinLength)
	}

	if fieldType.Pattern != nil {
		if _, err := regexp.MatchString(*fieldType.Pattern, str); err != nil {
			return nil, fmt.Errorf("value does not match pattern, pattern is %v", *fieldType.Pattern)
		}
	}

	return &str, nil
}

type IntFieldType struct {
	Nullable     bool
	DefaultValue *int64
	MinValue     *int64
	MaxValue     *int64
}

func (fieldType IntFieldType) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		return fieldType.DefaultValue, nil
	}

	i, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("invalid value, expected integer")
	}

	if fieldType.MinValue != nil && i < *fieldType.MinValue {
		return nil, fmt.Errorf("value too small, min value is %v", *fieldType.MinValue)
	}

	if fieldType.MaxValue != nil && i > *fieldType.MaxValue {
		return nil, fmt.Errorf("value too big, max value is %v", *fieldType.MaxValue)
	}

	return &i, nil
}

type FloatFieldType struct {
	Nullable     bool
	DefaultValue *float64
	MinValue     *float64
	MaxValue     *float64
}

func (fieldType FloatFieldType) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		return fieldType.DefaultValue, nil
	}

	f, ok := value.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value, expected float")
	}

	if fieldType.MinValue != nil && f < *fieldType.MinValue {
		return nil, fmt.Errorf("value too small, min value is %v", *fieldType.MinValue)
	}

	if fieldType.MaxValue != nil && f > *fieldType.MaxValue {
		return nil, fmt.Errorf("value too big, max value is %v", *fieldType.MaxValue)
	}

	return &f, nil
}

type BoolFieldType struct {
	Nullable     bool
	DefaultValue *bool
}

func (fieldType BoolFieldType) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		return fieldType.DefaultValue, nil
	}

	b, ok := value.(bool)
	if !ok {
		return nil, fmt.Errorf("invalid value, expected bool")
	}

	return &b, nil
}

type DateTimeFieldType struct {
	Nullable           bool
	CreateDefaultValue func() time.Time
	CreateMinValue     func() time.Time
	CreateMaxValue     func() time.Time
}

func (fieldType DateTimeFieldType) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		if fieldType.CreateDefaultValue != nil {
			defaultValue := fieldType.CreateDefaultValue()
			return &defaultValue, nil
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

	return &d, nil
}

type EnumFieldType struct {
	Nullable     bool
	DefaultValue *string
	EnumValues   []string
}

func (fieldType EnumFieldType) ValidateValue(value any) (any, error) {
	if fieldType.DefaultValue != nil && !slices.Contains(fieldType.EnumValues, *fieldType.DefaultValue) {
		return nil, fmt.Errorf("configuration error, invalid default value")
	}

	if err := validateNullable(fieldType.Nullable, value); err != nil {
		return nil, err
	}

	if value == nil {
		return fieldType.DefaultValue, nil
	}

	str, ok := value.(string)
	if !ok || !slices.Contains(fieldType.EnumValues, str) {
		return nil, fmt.Errorf("invalid value, expected one of [%s]", strings.Join(fieldType.EnumValues, ", "))
	}

	return str, nil
}

type SingleRelationFieldType struct {
	Nullable   bool
	Collection string
}

func (fieldType SingleRelationFieldType) ValidateValue(value any) (any, error) {
	if err := validateNullable(fieldType.Nullable, value); err != nil {
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
