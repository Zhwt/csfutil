package csfutil

import (
	"bytes"
	"encoding/binary"
	"strings"
)

// checkIdentifier() values.
const (
	CSFFileIdentifier        = " FSC" // CSF file header magic value
	LabelIdentifier          = " LBL" // Label identifier
	ValueIdentifier          = " RTS" // Values identifier
	ValueWithExtraIdentifier = "WRTS" // Values with extra value
)

// Languages holds all language name according to ModEnc.
var Languages = []string{"US (English)", "UK (English)", "German", "French", "Spanish", "Italian", "Japanese", "Jabberwockie", "Korean", "Chinese", "Unknown"}

// Label holds info for one CSF Label section.
type Label struct {
	Offset      uint   // offset from file beginning
	StringPairs uint   // number of string pairs associated with this label
	Value       []byte // label name
}

// ValueString returns LabelName section of a CSF Label.
func (l Label) ValueString() string {
	return string(l.Value)
}

// Bytes returns the binary form of a CSF Label data.
func (l Label) Bytes() []byte {
	b := []byte{}
	buf := bytes.NewBuffer(b)
	dword := make([]byte, 4)

	buf.WriteString(LabelIdentifier)
	binary.LittleEndian.PutUint32(dword, uint32(1))
	buf.Write(dword)
	binary.LittleEndian.PutUint32(dword, uint32(len(l.Value)))
	buf.Write(dword)
	buf.Write(l.Value)

	return buf.Bytes()
}

// Write writes s to CSF Label LabelName section.
func (l *Label) Write(s string) {
	l.Value = []byte(s)
}

// Value holds info for one CSF Values section.
type Value struct {
	Offset     uint
	HaveExtra  bool
	Value      []byte
	ExtraValue []byte
}

// ValueString returns Value section of a CSF Value.
func (v Value) ValueString() string {
	// decode by NOT every byte of the value data
	b := make([]byte, len(v.Value))
	copy(b, v.Value)
	for i := range b {
		b[i] = ^b[i]
	}

	s, _ := DecodeUTF16(b)
	return s
}

// ExtraValueString returns ExtraValue section of a CSF Value.
func (v Value) ExtraValueString() string {
	if !v.HaveExtra {
		return ""
	}

	return string(v.ExtraValue)
}

// Bytes returns the binary form of CSF Value data.
func (v Value) Bytes() []byte {
	b := []byte{}
	buf := bytes.NewBuffer(b)
	dword := make([]byte, 4)

	if v.HaveExtra {
		buf.WriteString(ValueWithExtraIdentifier)
	} else {
		buf.WriteString(ValueIdentifier)
	}
	binary.LittleEndian.PutUint32(dword, uint32(len(v.Value)/2))
	buf.Write(dword)
	buf.Write(v.Value)
	if v.HaveExtra {
		binary.LittleEndian.PutUint32(dword, uint32(len(v.ExtraValue)))
		buf.Write(dword)
		buf.Write(v.ExtraValue)
	}

	return buf.Bytes()
}

// Write writes s to CSF Value Values section.
func (v *Value) Write(s string) {
	v.Value = EncodeUTF16(s)

	b := v.Value
	for i := range b {
		b[i] = ^b[i]
	}
}

// WriteExtra writes s to CSF Value ExtraValue section.
func (v *Value) WriteExtra(s string) {
	v.HaveExtra = true
	v.ExtraValue = []byte(s)
}

// ClearExtra clears CSF Value ExtraValue section.
func (v *Value) ClearExtra() {
	v.HaveExtra = false
	v.ExtraValue = nil
}

// LabelValue is used to store Label-Value pair.
type LabelValue struct {
	Label Label
	Value Value
}

// String prints a Label-Value pair in "LabelName -> Value , ExtraValue" format.
// Mainly for debug purpose.
func (lv LabelValue) String() string {
	sb := strings.Builder{}
	sb.WriteString(lv.Label.ValueString())
	sb.WriteString(" -> ")
	sb.WriteString(lv.Value.ValueString())
	if lv.Value.HaveExtra {
		sb.WriteString(" , ")
		sb.WriteString(lv.Value.ExtraValueString())
	}
	return sb.String()
}

// Bytes returns the binary form of a CSF Value data.
func (lv LabelValue) Bytes() []byte {
	lb := lv.Label.Bytes()
	vb := lv.Value.Bytes()
	return append(lb[:], vb...)
}

// NewLabelValue does the same as NewLabelValueWithOffset, but doesn't touch
// the Offset field.
func NewLabelValue(label, value string, extra ...string) LabelValue {
	return NewLabelValueWithOffset(label, 0, value, 0, extra...)
}

// NewLabelValueWithOffset combines a label and a value and sets their offset, finally
// returns its LabelValue form.
func NewLabelValueWithOffset(label string, labelOffset uint, value string, valueOffset uint, extra ...string) LabelValue {
	l := Label{
		Offset:      labelOffset,
		StringPairs: 1,
	}
	l.Write(label)
	v := Value{Offset: valueOffset}
	v.Write(value)
	if len(extra) > 0 {
		v.WriteExtra(extra[0])
	}
	return LabelValue{
		Label: l,
		Value: v,
	}
}
