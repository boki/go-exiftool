package exiftool

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

const (
	defaultString = ""
	defaultFloat  = float64(0)
	defaultInt    = int64(0)
)

// ErrKeyNotFound is a sentinel error used when a queried key does not exist
var ErrKeyNotFound = errors.New("key not found")

// FileMetadataValue ...
type FileMetadataValue struct {
	Label string
	Value interface{}
}

// FileMetadataValues ...
type FileMetadataValues []FileMetadataValue

// FileMetadata is a structure that represents an exiftool extraction. File contains the
// filename that had to be extracted. If anything went wrong, Err will not be nil. Fields
// stores extracted fields.
type FileMetadata struct {
	File   string
	Groups map[string]FileMetadataValues
	Err    error
}

// UnmarshalJSON decodes the JSON encoding of FileMetadataValues.
func (g *FileMetadataValues) UnmarshalJSON(data []byte) error {
	l := len(data)
	if l == 0 || l <= 2 {
		return nil
	}
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	if t, err := dec.Token(); err != nil {
		return err
	} else if t != json.Delim('{') {
		return errors.New("expected {")
	}
	for {
		var l string
		if t, err := dec.Token(); err != nil {
			return fmt.Errorf("read label: %w", err)
		} else if t == json.Delim('}') {
			break
		} else if s, ok := t.(string); ok {
			l = s
		} else {
			return errors.New("expected string")
		}
		var v interface{}
		if t, err := dec.Token(); err != nil {
			return fmt.Errorf("read value: %w", err)
		} else if t == json.Delim('[') {
			a := []interface{}{}
			for {
				// TODO(bg): Support all types
				if t, err := dec.Token(); err != nil {
					return fmt.Errorf("read array value: %w", err)
				} else if t == json.Delim(']') {
					break
				} else if s, ok := t.(string); ok {
					a = append(a, s)
				}
			}
			v = a
		} else if s, ok := t.(bool); ok {
			v = s
		} else if s, ok := t.(float64); ok {
			v = s
		} else if s, ok := t.(json.Number); ok {
			if f, err := s.Float64(); err == nil {
				v = f
			} else if i, err := s.Int64(); err == nil {
				v = i
			} else {
				v = s.String()
			}
		} else if s, ok := t.(string); ok {
			v = s
		} else {
			return fmt.Errorf("unexpected token %v", t)
		}
		*g = append(*g, FileMetadataValue{l, v})
	}
	return nil
}

func (g FileMetadataValues) field(k string) (interface{}, bool) {
	for _, f := range g {
		if f.Label == k {
			return f.Value, true
		}
	}
	return nil, false
}

// GetString returns a field value as string and an error if one occurred.
// KeyNotFoundError will be returned if the key can't be found
func (g FileMetadataValues) GetString(k string) (string, error) {
	v, found := g.field(k)
	if !found {
		return defaultString, ErrKeyNotFound
	}

	return toString(v), nil
}

func toString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetFloat returns a field value as float64 and an error if one occurred.
// KeyNotFoundError will be returned if the key can't be found.
func (g FileMetadataValues) GetFloat(k string) (float64, error) {
	v, found := g.field(k)
	if !found {
		return defaultFloat, ErrKeyNotFound
	}

	switch v := v.(type) {
	case string:
		return toFloatFallback(v)
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	default:
		str := fmt.Sprintf("%v", v)
		return toFloatFallback(str)
	}
}

func toFloatFallback(str string) (float64, error) {
	f, err := strconv.ParseFloat(str, -1)
	if err != nil {
		return defaultFloat, fmt.Errorf("float64 parsing error (%v): %w", str, err)
	}

	return f, nil
}

// GetInt returns a field value as int64 and an error if one occurred.
// KeyNotFoundError will be returned if the key can't be found, ParseError if
// a parsing error occurs.
func (g FileMetadataValues) GetInt(k string) (int64, error) {
	v, found := g.field(k)
	if !found {
		return defaultInt, ErrKeyNotFound
	}

	switch v := v.(type) {
	case string:
		return toIntFallback(v)
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		str := fmt.Sprintf("%v", v)
		return toIntFallback(str)
	}
}

func toIntFallback(str string) (int64, error) {
	f, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultInt, fmt.Errorf("int64 parsing error (%v): %w", str, err)
	}

	return f, nil
}

// GetStrings returns a field value as []string and an error if one occurred.
// KeyNotFoundError will be returned if the key can't be found.
func (g FileMetadataValues) GetStrings(k string) ([]string, error) {
	v, found := g.field(k)
	if !found {
		return []string{}, ErrKeyNotFound
	}

	switch v := v.(type) {
	case []interface{}:
		is := v
		res := make([]string, len(is))

		for i, v2 := range is {
			res[i] = toString(v2)
		}

		return res, nil
	default:
		return []string{toString(v)}, nil
	}
}
