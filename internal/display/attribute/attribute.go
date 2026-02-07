package attribute

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
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

type styled[T any] struct {
	V     T
	Style lipgloss.Style
}

func (s styled[T]) String() string {
	return s.Style.Render(fmt.Sprintf("%v", s.V))
}

func (s styled[T]) toValue() Value {
	return Value{
		Text:  fmt.Sprintf("%v", s.V),
		Style: s.Style,
	}
}

func Styled[T any](v T, style ...lipgloss.Style) styled[T] {
	styl := lipgloss.NewStyle()
	if len(style) > 0 {
		styl = style[0]
	}
	return styled[T]{
		V:     v,
		Style: styl,
	}
}

func ID[T any](v T) KeyValue {
	return Attribute("ID", Styled(v, idStyle))
}

type KeyValue struct {
	Key   Value
	Value Value
}

func Attribute[T any](key string, value styled[T]) KeyValue {
	return KeyValue{
		Key:   Styled(key, keyStyle).toValue(),
		Value: value.toValue(),
	}
}

func Int(key string, value styled[int]) KeyValue {
	return Attribute(key, value)
}

func Bool(key string, value styled[bool]) KeyValue {
	return Attribute(key, value)
}

func Stringer(key string, value styled[fmt.Stringer]) KeyValue {
	return Attribute(key, value)
}

func ValueOf[T any](v T, style ...lipgloss.Style) Value {
	return Styled(v, style...).toValue()
}

func OptionalValue[T any](value *T, formatter func(T) string) Value {
	if value == nil {
		return ValueOf("-")
	}
	return ValueOf(formatter(*value))
}

func OptionalStringValue(value *string) Value {
	if value == nil || *value == "" {
		return ValueOf("-")
	}
	return ValueOf(*value)
}

// Optional renders the provided pointer using formatter. Missing values are displayed as "-".
func Optional[T any](key string, value *T, formatter func(T) string) KeyValue {
	return KeyValue{
		Key:   Styled(key, keyStyle).toValue(),
		Value: OptionalValue(value, formatter),
	}
}

// OptionalString renders string pointers, treating nil or empty values as "-".
func OptionalString(key string, value *string) KeyValue {
	return KeyValue{
		Key:   Styled(key, keyStyle).toValue(),
		Value: OptionalStringValue(value),
	}
}
