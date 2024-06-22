package jsonmask

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	pathKey          = "/"
	randomIntRange   = 1000
	randomFloatRange = "1000.3"
)

// list of func type that must be satisfied to add a custom mask
type (
	MaskStringFunc  func(path, value string) (string, error)
	MaskIntFunc     func(path string, value int) (int, error)
	MaskFloat64Func func(path string, value float64) (float64, error)
)

// JsonMask is a struct that defines the masking process
type JsonMask struct {
	maskStringFunc  MaskStringFunc
	maskIntFunc     MaskIntFunc
	maskFloat64Func MaskFloat64Func
	pathFields      map[string]struct{}
	globalFields    map[string]struct{}
}

// NewJSONMask initializes a JsonMask
// Mask fields:
// 1. Global (a,b,c) - will mask all encountered json fields (nested fields will be masked entirely)
// 2. XPath (/a/b/c) - will mask only specified json fields by xpath
func NewJSONMask(fields ...string) *JsonMask {
	m := &JsonMask{
		pathFields:   make(map[string]struct{}),
		globalFields: make(map[string]struct{}),
	}

	for _, field := range fields {
		if strings.Contains(field, pathKey) {
			m.pathFields[field] = struct{}{}
		} else {
			m.globalFields[field] = struct{}{}
		}
	}

	return m
}

// RegisterMaskStringFunc method for adding MaskStringFunc to JsonMask
func (j *JsonMask) RegisterMaskStringFunc(fn MaskStringFunc) {
	j.maskStringFunc = fn
}

// RegisterMaskIntFunc method for adding MaskIntFunc to JsonMask
func (j *JsonMask) RegisterMaskIntFunc(fn MaskIntFunc) {
	j.maskIntFunc = fn
}

// RegisterMaskFloat64Func method for adding MaskFloat64Func to JsonMask
func (j *JsonMask) RegisterMaskFloat64Func(fn MaskFloat64Func) {
	j.maskFloat64Func = fn
}

// Mask method for masking JSON fields globally or by xpath
func (j *JsonMask) Mask(value string) (string, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(value), &m); err != nil {
		return "", fmt.Errorf("json unmarshal: %w", err)
	}

	if err := j.mask("", m); err != nil {
		return "", fmt.Errorf("mask: %w", err)
	}

	b, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}

	return string(b), nil
}

// mask method for masking parsed map with global and xpath fields
func (j *JsonMask) mask(pk string, m map[string]any) (err error) {
	for k, val := range m {
		fk := pk + pathKey + k
		switch v := val.(type) {
		case map[string]any:
			if _, ok := j.globalFields[k]; ok {
				if err = j.maskAll(fk, v); err != nil {
					return err
				}
				break
			}

			if err = j.mask(fk, v); err != nil {
				return err
			}
		case string:
			if j.maskStringFunc != nil {
				if _, ok := j.globalFields[k]; ok {
					m[k], err = j.maskStringFunc(fk, v)
					if err != nil {
						return err
					}
				}

				if _, ok := j.pathFields[fk]; ok {
					m[k], err = j.maskStringFunc(fk, v)
					if err != nil {
						return err
					}
				}
			}
		case float64:
			if isInteger(v) {
				if j.maskIntFunc == nil {
					break
				}

				if _, ok := j.globalFields[k]; ok {
					m[k], err = j.maskIntFunc(fk, int(v))
					if err != nil {
						return err
					}
				}

				if _, ok := j.pathFields[fk]; ok {
					m[k], err = j.maskIntFunc(fk, int(v))
					if err != nil {
						return err
					}
				}
			}

			if j.maskFloat64Func == nil {
				break
			}

			if _, ok := j.globalFields[k]; ok {
				m[k], err = j.maskFloat64Func(fk, v)
				if err != nil {
					return err
				}
			}

			if _, ok := j.pathFields[fk]; ok {
				m[k], err = j.maskFloat64Func(fk, v)
				if err != nil {
					return err
				}
			}
		case []any:
			if err = j.maskSlice(k, fk, v, true); err != nil {
				return err
			}
		case bool, nil: // skip nil or boolean types
		default:
			return fmt.Errorf("unknow type: %T", v)
		}
	}

	return nil
}

// maskAll method for masking all what inside under key
func (j *JsonMask) maskAll(pk string, m map[string]any) (err error) {
	for k, val := range m {
		fk := pk + pathKey + k
		switch v := val.(type) {
		case map[string]any:
			if err = j.maskAll(fk, v); err != nil {
				return err
			}
		case string:
			if j.maskStringFunc != nil {
				m[k], err = j.maskStringFunc(fk, v)
				if err != nil {
					return err
				}
			}
		case float64:
			if isInteger(v) {
				if j.maskIntFunc == nil {
					break
				}

				if m[k], err = j.maskIntFunc(fk, int(v)); err != nil {
					return err
				}
			}

			if j.maskFloat64Func != nil {
				if m[k], err = j.maskFloat64Func(fk, v); err != nil {
					return err
				}
			}
		case []any:
			if err = j.maskSlice(k, fk, v, false); err != nil {
				return err
			}
		case bool, nil: // skip nil or boolean types
		default:
			return fmt.Errorf("unknow type: %T", v)
		}
	}

	return nil
}

// maskSlice method for masking values what inside array
func (j *JsonMask) maskSlice(k, pk string, sl []any, ignoreGlobal bool) (err error) {
	for i, val := range sl {
		fk := fmt.Sprintf("%s[%d]", pk, i)
		switch v := val.(type) {
		case map[string]any:
			if _, ok := j.globalFields[k]; !ignoreGlobal || ok {
				if err = j.maskAll(fk, v); err != nil {
					return err
				}

				break
			}

			if err = j.mask(fk, v); err != nil {
				return err
			}
		case string:
			if j.maskStringFunc != nil {
				if _, ok := j.globalFields[k]; !ignoreGlobal || ok {
					sl[i], err = j.maskStringFunc(fk, v)
					if err != nil {
						return err
					}
				}

				if _, ok := j.pathFields[fk]; ok {
					sl[i], err = j.maskStringFunc(pk, v)
					if err != nil {
						return err
					}
				}
			}
		case float64:
			if isInteger(v) {
				if j.maskIntFunc == nil {
					break
				}

				if _, ok := j.globalFields[k]; !ignoreGlobal || ok {
					sl[i], err = j.maskIntFunc(fk, int(v))
					if err != nil {
						return err
					}
				}

				if _, ok := j.pathFields[fk]; ok {
					sl[i], err = j.maskIntFunc(pk, int(v))
					if err != nil {
						return err
					}
				}
			}

			if j.maskFloat64Func == nil {
				break
			}

			if _, ok := j.globalFields[k]; !ignoreGlobal || ok {
				if sl[i], err = j.maskFloat64Func(fk, v); err != nil {
					return err
				}
			}

			if _, ok := j.pathFields[fk]; ok {
				sl[i], err = j.maskFloat64Func(fk, v)
				if err != nil {
					return err
				}
			}
		case []any:
			if err = j.maskSlice(k, fk, v, false); err != nil {
				return err
			}
		case bool, nil: // skip nil or boolean types
		default:
			return fmt.Errorf("unknow type: %T", v)
		}
	}

	return nil
}

// MaskFilledString masks the string length of the value with the same length or by passed length
func MaskFilledString(maskChar string, length ...int) MaskStringFunc {
	hasLen := len(length) > 0
	return func(_, val string) (string, error) {
		if hasLen {
			return strings.Repeat(maskChar, length[0]), nil
		}

		return strings.Repeat(maskChar, utf8.RuneCountInString(val)), nil
	}
}

// MaskHashString masks and hashes (sha1) a string
func MaskHashString() MaskStringFunc {
	return func(_, val string) (string, error) {
		hash := sha1.Sum(([]byte)(val))
		return hex.EncodeToString(hash[:]), nil
	}
}

// MaskRandomInt masks converts an integer (int) into a random number in range (default 1000)
func MaskRandomInt(arg ...int) MaskIntFunc {
	hasArg := len(arg) > 0
	return func(_ string, val int) (int, error) {
		rn := randomIntRange
		if hasArg {
			rn = arg[0]
		}

		return rand.Intn(rn), nil
	}
}

// MaskRandomFloat64 converts a float64 to a random number in range (default 1000.3)
// if you pass "1000.3" to arg, it sets a random number in the range of 0.000 to 999.999
func MaskRandomFloat64(arg ...string) MaskFloat64Func {
	hasArg := len(arg) > 0
	return func(_ string, val float64) (float64, error) {
		var (
			i, d int
			err  error
			rn   = randomFloatRange
		)

		if hasArg {
			rn = arg[0]
		}

		digits := strings.Split(rn, ".")
		if len(digits) > 0 {
			if i, err = strconv.Atoi(digits[0]); err != nil {
				return 0, err
			}
		}
		if len(digits) == 2 {
			if d, err = strconv.Atoi(digits[1]); err != nil {
				return 0, err
			}
		}

		dd := math.Pow10(d)
		x := float64(int(rand.Float64() * float64(i) * dd))

		return x / dd, nil
	}
}

// isInteger method for check float value on integer
func isInteger(val float64) bool {
	return val == float64(int(val))
}
