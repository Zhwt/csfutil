package csfutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// A CSFUtil provides some basic methods for accessing a CSF file.
// The zero value of CSFUtil cannot be used and may cause nil pointer panic,
// always use Open() to open a file.
type CSFUtil struct {
	dword []byte // buffer for read uint values

	// file holds a temporary reference to the underlying CSF file while
	// reading contents.
	file *os.File

	// name of currently opening file
	filename string

	Version    uint // CSF file Version section
	NumLabels  uint // CSF file NumLabels section
	NumStrings uint // CSF file NumStrings section
	Unused     uint // CSF file (unused) section
	Language   uint // CSF file Language section

	// Values holds all CSF LabelValue pair, uses capitalized CSF Label name
	// as map key
	Values map[string]LabelValue
	// Categories holds CSF Label categories -> CSF Label Name map, uses
	// capitalized category name as map key
	Categories map[string][]string
	// Order stores original CSF Label order according to the CSF file, in
	// capitalized form.
	Order []string
}

// pos returns current reading position.
func (r *CSFUtil) pos() uint {
	offset, err := r.file.Seek(0, io.SeekCurrent)
	if err != nil {
		panic(err)
	}

	return uint(offset)
}

// read reads next size bytes and return them.
func (r *CSFUtil) read(size uint) ([]byte, error) {
	b := make([]byte, size)
	n, err := r.file.Read(b)
	if err != nil {
		return nil, err
	}

	if uint(n) != size {
		return b, fmt.Errorf("not enough data, want [%d]byte, got [%d]byte", size, n)
	}

	return b, nil
}

// readDWORD reads the next 4 bytes and return them.
func (r *CSFUtil) readDWORD() ([]byte, error) {
	n, err := r.file.Read(r.dword)
	if err != nil {
		return nil, err
	}

	if n != 4 {
		return nil, fmt.Errorf("not a dword value")
	}

	return r.dword, nil
}

// readUInt reads the next DWORD and return them.
func (r *CSFUtil) readUInt() (uint, error) {
	_, err := r.readDWORD()
	if err != nil {
		return 0, err
	}
	return uint(binary.LittleEndian.Uint32(r.dword)), nil
}

// checkIdentifier checks if the following 4 bytes is presented in the given
// identifier array.
func (r *CSFUtil) checkIdentifier(ids ...string) error {
	_, err := r.readDWORD()
	if err != nil {
		return err
	}

	for _, id := range ids {
		if bytes.Equal(r.dword, []byte(id)) {
			return nil
		}
	}

	return fmt.Errorf("not a valid identifier, want %v, got `%s`", ids, string(r.dword))
}

// readHeader reads and parse CSF file header for later use.
func (r *CSFUtil) readHeader() error {
	var i uint
	var err error
	// file header
	if err := r.checkIdentifier(CSFFileIdentifier); err != nil {
		return fmt.Errorf("not a valid CSF file, reason: %w at %x", err, r.pos())
	}

	// file version
	i, err = r.readUInt()
	if err != nil {
		return err
	}
	r.Version = i

	// num labels
	i, err = r.readUInt()
	if err != nil {
		return err
	}
	r.NumLabels = i

	// num strings
	i, err = r.readUInt()
	if err != nil {
		return err
	}
	r.NumStrings = i

	// unused
	i, err = r.readUInt()
	if err != nil {
		return err
	}
	r.Unused = i

	// language
	i, err = r.readUInt()
	if err != nil {
		return err
	}
	r.Language = i

	return nil
}

// readLabel reads the following CSF Label and returns.
func (r *CSFUtil) readLabel() (Label, error) {
	offset := r.pos()

	if err := r.checkIdentifier(LabelIdentifier); err != nil {
		return Label{}, fmt.Errorf("not a valid label start, reason: %w at %x", err, offset)
	}

	pairs, err := r.readUInt()
	if err != nil {
		return Label{}, err
	}

	length, err := r.readUInt()
	if err != nil {
		return Label{}, err
	}

	value, err := r.read(length)
	if err != nil {
		return Label{}, err
	}

	lbl := Label{
		Offset:      offset,
		StringPairs: pairs,
		Value:       value,
	}

	return lbl, nil
}

// readValue reads the following CSF Value and returns.
func (r *CSFUtil) readValue() (Value, error) {
	offset := r.pos()

	if err := r.checkIdentifier(ValueIdentifier, ValueWithExtraIdentifier); err != nil {
		return Value{}, fmt.Errorf("not a valid value start, reason: %w, at %x", err, offset)
	}

	hasExtra := bytes.Equal(r.dword, []byte("WRTS"))

	length, err := r.readUInt()
	if err != nil {
		return Value{}, err
	}

	value, err := r.read(length * 2)
	if err != nil {
		return Value{}, err
	}

	var extraVal []byte
	if hasExtra {
		extraLength, err := r.readUInt()

		extraVal, err = r.read(extraLength)
		if err != nil {
			return Value{}, err
		}
	}

	val := Value{
		Offset:     offset,
		HaveExtra:  hasExtra,
		Value:      value,
		ExtraValue: extraVal,
	}

	return val, nil
}

// readContent reads the CSF file till end and store all parsed
// CSF LabelValue pairs into Values map.
func (r *CSFUtil) readContent() error {
	mapLabelValue := map[string]LabelValue{}
	mapCate := map[string][]string{}
	listOrder := []string{}

	counter := 0
	for {
		label, err := r.readLabel()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return err
			}
		}

		value, err := r.readValue()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return err
			}
		}

		labelName := label.ValueString()
		upperLabelName := strings.ToUpper(labelName)
		// make sure entries in Categories and Order are unique
		if _, ok := mapLabelValue[upperLabelName]; !ok {
			// ignore Labels that don't have a category
			if strings.Contains(upperLabelName, ":") {
				categoryName := upperLabelName[0:strings.Index(upperLabelName, ":")]
				if list, ok := mapCate[categoryName]; ok {
					mapCate[categoryName] = append(list, labelName)
				} else {
					mapCate[categoryName] = []string{}
				}
			} else {
				if list, ok := mapCate[""]; ok {
					mapCate[""] = append(list, labelName)
				} else {
					mapCate[""] = []string{}
				}
			}

			listOrder = append(listOrder, upperLabelName)
		}

		lv := LabelValue{Label: label, Value: value}
		mapLabelValue[upperLabelName] = lv

		counter++
		if counter > 20000 {
			return fmt.Errorf("too many strings")
		}
	}

	r.Values = mapLabelValue
	r.Categories = mapCate
	r.Order = listOrder

	return nil
}

// openAndParse open and parses the named file.
// Must be called before using CSFUtil.
func (r *CSFUtil) openAndParse(name string) error {
	f, err := os.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	// initialization
	r.file = f
	r.dword = make([]byte, 4)

	// read file as ModEnc suggests
	if err := r.readHeader(); err != nil {
		return err
	}

	if err := r.readContent(); err != nil {
		return err
	}

	return nil
}

// LanguageName returns language name of current CSF file, if not in the list,
// "Unknown" is returned.
func (r *CSFUtil) LanguageName() string {
	if r.Language > 9 {
		return Languages[10]
	} else {
		return Languages[r.Language]
	}
}

// PrintAll prints all LabelValue pairs stored in Values in
// "LABELNAME (upper case) => LabelName -> Value , ExtraValue" format.
// Mainly for debug purpose.
func (r *CSFUtil) PrintAll() {
	for label, labelValue := range r.Values {
		fmt.Println(label, "=>", labelValue)
	}
}

// WriteLabelValue do the same as WriteLabelValueAfter() does, but will
// always write at file end if no old CSF Label presents.
func (r *CSFUtil) WriteLabelValue(lv LabelValue, overwriteLabel bool) {
	r.WriteLabelValueBefore(lv, overwriteLabel, "")
}

// WriteLabelValueBefore writes given CSF LabelValue pair into the file, will
// overwrite and preserve order if old CSF Label presents, otherwise will write
// before given label. If the given label is not found, then it will write at
// file end.
func (r *CSFUtil) WriteLabelValueBefore(lv LabelValue, overwriteLabel bool, successor string) {
	successorUpper := strings.ToUpper(successor)
	labelNameUpper := strings.ToUpper(lv.Label.ValueString())
	if _, ok := r.Values[labelNameUpper]; ok {
		// overwrite exist, preserve original order
		if !overwriteLabel {
			lv.Label.Write(r.Values[labelNameUpper].Label.ValueString())
		}
	} else if _, ok := r.Values[successorUpper]; ok {
		// new in the middle
		pos := len(r.Order)
		l := pos
		for i, s := range r.Order {
			if s == successorUpper {
				pos = i
			}
		}
		r.Order = append(append(r.Order[0:pos], labelNameUpper), r.Order[pos:l]...)
	} else {
		// new at the end
		r.Order = append(r.Order, labelNameUpper)
	}
	r.Values[labelNameUpper] = lv

	r.NumLabels = uint(len(r.Values))
	r.NumStrings = uint(len(r.Values))
}

// RemoveLabelValue is used to eliminate the given CSF LabelValue
func (r *CSFUtil) RemoveLabelValue(name string) {
	if len(r.Order) == 0 {
		return
	} else {
		upperName := strings.ToUpper(name)
		delete(r.Values, upperName)
		for i, s := range r.Order {
			if s == upperName {
				if len(r.Order) > 1 {
					r.Order = append(r.Order[:i], r.Order[i+1:]...)
				} else {
					r.Order = []string{}
				}
				break
			}
		}
		r.NumLabels = uint(len(r.Values))
		r.NumStrings = uint(len(r.Values))
	}
}

// Save saves the changes to the CSF file. Intermediate backup file will be
// created as $tmp_filename.csf.
func (r *CSFUtil) Save() error {
	f, err := os.OpenFile("$tmp_"+r.filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	b := []byte{}
	buf := bytes.NewBuffer(b)
	dword := make([]byte, 4)

	// write file header
	buf.WriteString(CSFFileIdentifier)
	binary.LittleEndian.PutUint32(dword, uint32(r.Version))
	buf.Write(dword)
	binary.LittleEndian.PutUint32(dword, uint32(r.NumLabels))
	buf.Write(dword)
	binary.LittleEndian.PutUint32(dword, uint32(r.NumStrings))
	buf.Write(dword)
	binary.LittleEndian.PutUint32(dword, uint32(r.Unused))
	buf.Write(dword)
	binary.LittleEndian.PutUint32(dword, uint32(r.Language))
	buf.Write(dword)

	for _, s := range r.Order {
		buf.Write(r.Values[s].Bytes())
	}

	_, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	f.Close()

	if err := os.Rename("$tmp_"+r.filename, r.filename); err != nil {
		return err
	}

	return nil
}

// Open opens the given file and parse it, returning pointer to CSFUtil.
// Remember to call Close() on the returning CSFUtil.
func Open(name string) (*CSFUtil, error) {
	reader := &CSFUtil{filename: name}
	return reader, reader.openAndParse(name)
}

// MustOpen do the same thing as Open, but panics if error occurs.
func MustOpen(name string) *CSFUtil {
	u, err := Open(name)
	if err != nil {
		panic(err)
	}

	return u
}

func New(name string, version, unused, language uint) *CSFUtil {
	return &CSFUtil{
		filename:   name,
		Version:    version,
		NumLabels:  0,
		NumStrings: 0,
		Unused:     unused,
		Language:   language,
		Values:     map[string]LabelValue{},
		Categories: map[string][]string{},
		Order:      []string{},
	}
}
