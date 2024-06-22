package jsonmask

import (
	"fmt"
	"testing"
)

func TestNewJSONMask(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		mask    *JsonMask
		rFuncs  []interface{}
		expect  string
		wantErr bool
	}{
		{
			name:    "should hash all keys for key metadata",
			mask:    NewJSONMask("metadata", "key1"),
			rFuncs:  []interface{}{MaskHashString()},
			value:   `{"name": "testname", "metadata": {"labels": {"key1": "value1", "key2": 123456}, "annotations": {"key1": "value1"}}}`,
			expect:  `{"metadata":{"annotations":{"key1":"8107759ababcbfa34bcb02bc4309caf6354982ab"},"labels":{"key1":"8107759ababcbfa34bcb02bc4309caf6354982ab","key2":123456}},"name":"testname"}`,
			wantErr: false,
		},
		{
			name:    "should hash all keys for nested json with different types",
			mask:    NewJSONMask("annotations", "key2", "/metadata/labels/key3[1]", "/metadata/labels/key3[4]/key5"),
			rFuncs:  []interface{}{MaskHashString()},
			value:   `{"name": "testname", "key4": {}, "key5": null, "key3": ["value3", 111, 2.111, true, {"key2": "value1"}], "metadata": {"labels": {"key1": "value1", "key2": 123456, "key3": ["111", "value3", 2.111, true, {"key8":"value8", "key5": "value5"}, 111]}, "annotations": {"key1": "value1"}}}`,
			expect:  `{"key3":["value3",111,2.111,true,{"key2":"8107759ababcbfa34bcb02bc4309caf6354982ab"}],"key4":{},"key5":null,"metadata":{"annotations":{"key1":"8107759ababcbfa34bcb02bc4309caf6354982ab"},"labels":{"key1":"value1","key2":123456,"key3":["111","94ca33031e37aa3f3b67e5b921c729f08a6bba75",2.111,true,{"key5":"fedff2a2d3db7e4fb7c050e89496a33aac9f4a79","key8":"value8"},111]}},"name":"testname"}`,
			wantErr: false,
		},
		{
			name:    "should hash key by xpath",
			mask:    NewJSONMask("/metadata/labels/name"),
			rFuncs:  []interface{}{MaskHashString()},
			value:   `{"name": "testname", "metadata": {"labels": {"name": "testname", "key1": "value1"}, "annotations": {"key1": "value1"}}}`,
			expect:  `{"metadata":{"annotations":{"key1":"value1"},"labels":{"key1":"value1","name":"adc8de6b036aed3455b44abc62639e708d3ffef5"}},"name":"testname"}`,
			wantErr: false,
		},
		{
			name:    "should hash all fields with key fieldA",
			mask:    NewJSONMask("fieldA"),
			rFuncs:  []interface{}{MaskHashString()},
			value:   `{"fieldA": "valueA", "metadata": {"fieldA": "valueA"}}`,
			expect:  `{"fieldA":"fbae193291110932610c75eced91174b72406c95","metadata":{"fieldA":"fbae193291110932610c75eced91174b72406c95"}}`,
			wantErr: false,
		},
		{
			name:    "should hash all fields with key fieldA and fieldB",
			mask:    NewJSONMask("fieldA", "fieldB"),
			rFuncs:  []interface{}{MaskHashString()},
			value:   `{"fieldA": "valueA", "metadata": {"fieldA": "valueA", "fieldB": "valueB", "fieldC": "valueC"}}`,
			expect:  `{"fieldA":"fbae193291110932610c75eced91174b72406c95","metadata":{"fieldA":"fbae193291110932610c75eced91174b72406c95","fieldB":"9c9a5aaec27293677711598fdc277212c331c884","fieldC":"valueC"}}`,
			wantErr: false,
		},
		{
			name:    "should hash all fields with key with filled type",
			mask:    NewJSONMask("fieldA"),
			rFuncs:  []interface{}{MaskFilledString("*")},
			value:   `{"fieldA": "valueA", "metadata": {"fieldA": "valueAA", "fieldB": "valueB", "fieldC": "valueC"}}`,
			expect:  `{"fieldA":"******","metadata":{"fieldA":"*******","fieldB":"valueB","fieldC":"valueC"}}`,
			wantErr: false,
		},
		{
			name:    "should hash all fields with key with filled(10) type",
			mask:    NewJSONMask("fieldA"),
			rFuncs:  []interface{}{MaskFilledString("*", 10)},
			value:   `{"fieldA": "valueA", "metadata": {"fieldA": "valueAA", "fieldB": "valueB", "fieldC": "valueC"}}`,
			expect:  `{"fieldA":"**********","metadata":{"fieldA":"**********","fieldB":"valueB","fieldC":"valueC"}}`,
			wantErr: false,
		},
		{
			name:    "should hash all fields with key fieldA with int type",
			mask:    NewJSONMask("fieldA"),
			rFuncs:  []interface{}{testMaskRandomInt(9998)},
			value:   `{"fieldA": 12345, "metadata": {"fieldA": 1.234, "fieldB": "valueB", "fieldC": "valueC"}}`,
			expect:  `{"fieldA":9998,"metadata":{"fieldA":1.234,"fieldB":"valueB","fieldC":"valueC"}}`,
			wantErr: false,
		},
		{
			name:    "should hash all fields with key fieldA with float64 type",
			mask:    NewJSONMask("fieldA"),
			rFuncs:  []interface{}{testMaskRandomFloat64(998.998)},
			value:   `{"fieldA": 12345, "metadata": {"fieldA": 1.234, "fieldB": "valueB", "fieldC": "valueC"}}`,
			expect:  `{"fieldA":12345,"metadata":{"fieldA":998.998,"fieldB":"valueB","fieldC":"valueC"}}`,
			wantErr: false,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("#%d:%s", i, tt.name), func(t *testing.T) {
			if len(tt.rFuncs) > 0 {
				for _, rFn := range tt.rFuncs {
					switch fn := rFn.(type) {
					case MaskStringFunc:
						tt.mask.RegisterMaskStringFunc(fn)
					case MaskIntFunc:
						tt.mask.RegisterMaskIntFunc(fn)
					case MaskFloat64Func:
						tt.mask.RegisterMaskFloat64Func(fn)
					}
				}
			}

			got, err := tt.mask.Mask(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expect {
				t.Errorf("Process() got = %v, want %v", got, tt.expect)
			}
		})
	}
}

// BenchmarkNewJSONMaskHashString-16    	  343420	      3341 ns/op	    1929 B/op	      47 allocs/op
func BenchmarkNewJSONMaskHashString(b *testing.B) {
	var (
		mask = NewJSONMask("fieldA")
		json = `{"fieldA": "valueA", "metadata": {"fieldA": 1.234, "fieldB": "valueB", "fieldC": "valueC"}}`
	)
	mask.RegisterMaskStringFunc(MaskHashString())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = mask.Mask(json)
	}
}

// BenchmarkNewJSONMaskFilledString-16    	  335380	      3125 ns/op	    1771 B/op	      46 allocs/op
func BenchmarkNewJSONMaskFilledString(b *testing.B) {
	var (
		mask = NewJSONMask("fieldA")
		json = `{"fieldA": "valueA", "metadata": {"fieldA": 1.234, "fieldB": "valueB", "fieldC": "valueC"}}`
	)
	mask.RegisterMaskStringFunc(MaskFilledString("*"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = mask.Mask(json)
	}
}

// BenchmarkNewJSONMaskInt-16    	  355957	      3167 ns/op	    1717 B/op	      44 allocs/op
func BenchmarkNewJSONMaskInt(b *testing.B) {
	var (
		mask = NewJSONMask("fieldA")
		json = `{"fieldA": 123456, "metadata": {"fieldA": 1.234, "fieldB": "valueB", "fieldC": "valueC"}}`
	)
	mask.RegisterMaskIntFunc(MaskRandomInt())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = mask.Mask(json)
	}
}

// BenchmarkNewJSONMaskFloat64-16    	  353574	      3215 ns/op	    1785 B/op	      46 allocs/op
func BenchmarkNewJSONMaskFloat64(b *testing.B) {
	var (
		mask = NewJSONMask("fieldA")
		json = `{"fieldA": 123456, "metadata": {"fieldA": 1.234, "fieldB": "valueB", "fieldC": "valueC"}}`
	)
	mask.RegisterMaskFloat64Func(MaskRandomFloat64())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = mask.Mask(json)
	}
}

func testMaskRandomInt(val int) MaskIntFunc {
	return func(path string, value int) (int, error) {
		return val, nil
	}
}

func testMaskRandomFloat64(val float64) MaskFloat64Func {
	return func(path string, value float64) (float64, error) {
		return val, nil
	}
}
