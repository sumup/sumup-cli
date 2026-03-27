package display

import (
	"io"

	"github.com/sumup/sumup-cli/internal/display/attribute"
)

// DetailsBuilder helps commands build detail views without manual slice plumbing.
type DetailsBuilder struct {
	pairs []attribute.KeyValue
}

func NewDetailsBuilder() *DetailsBuilder {
	return &DetailsBuilder{}
}

func (b *DetailsBuilder) Add(key string, value attribute.Value) *DetailsBuilder {
	b.pairs = append(b.pairs, attribute.Attribute(key, value))
	return b
}

func (b *DetailsBuilder) AddPair(pair attribute.KeyValue) *DetailsBuilder {
	b.pairs = append(b.pairs, pair)
	return b
}

func (b *DetailsBuilder) AddID(value any) *DetailsBuilder {
	b.pairs = append(b.pairs, attribute.ID(value))
	return b
}

func (b *DetailsBuilder) AddOptionalString(key string, value *string) *DetailsBuilder {
	b.pairs = append(b.pairs, attribute.OptionalString(key, value))
	return b
}

func (b *DetailsBuilder) AddWhen(ok bool, pair attribute.KeyValue) *DetailsBuilder {
	if ok {
		b.pairs = append(b.pairs, pair)
	}
	return b
}

func (b *DetailsBuilder) Pairs() []attribute.KeyValue {
	return append([]attribute.KeyValue(nil), b.pairs...)
}

func (b *DetailsBuilder) Render(w io.Writer) error {
	return DataList(w, b.pairs)
}
