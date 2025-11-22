package attribute

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	keyStyle = lipgloss.NewStyle().Bold(true).Faint(true)
	idStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF61F2")).Bold(true)
)

type styled[T any] struct {
	V     T
	Style lipgloss.Style
}

func (s styled[T]) String() string {
	return s.Style.Render(fmt.Sprintf("%v", s.V))
}

func (s styled[T]) toStyledString() styled[string] {
	return styled[string]{
		V:     fmt.Sprintf("%v", s.V),
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
	return Attribute("ID", Styled(v, idStyle).toStyledString())
}

type KeyValue struct {
	Key   styled[string]
	Value styled[string]
}

func Attribute[T any](key string, value styled[T]) KeyValue {
	return KeyValue{
		Key:   Styled(key, keyStyle),
		Value: value.toStyledString(),
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
