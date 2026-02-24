package attribute

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	keyStyle = lipgloss.NewStyle().Bold(true).Faint(true)
	idStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF61F2")).Bold(true)
)

type Value struct {
	Text  string
	Style lipgloss.Style
}

func (v Value) String() string {
	return v.Style.Render(v.Text)
}

func Styled[T any](v T, style ...lipgloss.Style) Value {
	s := lipgloss.NewStyle()
	if len(style) > 0 {
		s = style[0]
	}
	return Value{
		Text:  fmt.Sprintf("%v", v),
		Style: s,
	}
}

func ID[T any](v T) KeyValue {
	return Attribute("ID", Styled(v, idStyle))
}

type KeyValue struct {
	Key   Value
	Value Value
}

func Attribute(key string, value Value) KeyValue {
	return KeyValue{
		Key:   Styled(key, keyStyle),
		Value: value,
	}
}

func ValueOf[T any](v T, style ...lipgloss.Style) Value {
	return Styled(v, style...)
}

func OptionalValue[T any](value *T) Value {
	if value == nil {
		return ValueOf("-")
	}
	return ValueOf(*value)
}

func OptionalStringValue(value *string) Value {
	if value == nil || *value == "" {
		return ValueOf("-")
	}
	return ValueOf(*value)
}

// Optional renders the provided pointer. Missing values are displayed as "-".
func Optional[T any](key string, value *T) KeyValue {
	return KeyValue{
		Key:   Styled(key, keyStyle),
		Value: OptionalValue(value),
	}
}

// OptionalString renders string pointers, treating nil or empty values as "-".
func OptionalString(key string, value *string) KeyValue {
	return KeyValue{
		Key:   Styled(key, keyStyle),
		Value: OptionalStringValue(value),
	}
}
