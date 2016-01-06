/* {{{ Copyright (c) Paul R. Tagliamonte <paultag@debian.org>, 2015
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE. }}} */

package control

import (
    "fmt"
    "strings"
    "strconv"
    "reflect"
)

type Marshalable interface {
    MarshalControl() (string, error)
}

func Marshal(incoming interface{}) (string, error) {
    /**
     * Given an incoming value, we want to try to turn it into a string. (marshal)
     *
     * If the data is a pointer, deref it to the Elem. (recurse)
     * If the data is a slice, create a series of Paragraphs, decoding each (marshalSlice)
     * If the data is a struct, decode the Struct (marshalStruct)
     * Otherwise, blow up.
     *
     * Struct Decoding (marshalStruct)
     *  Go over all public elements of the Control
     *  Get the key name by Name (or `control`)
     *  Get the value, and try to Marshal the Paragraph Value (marshalStructValue)
     *
     * Marshal the Paragraph Value (marshalStructValue)
     *  If the type is builtin, do the normal Marshal bits
     *  Otherwise, try to use the Marshalable interface
     *  Otherwise, blow up
     *
     */
     return marshal(reflect.ValueOf(incoming))
}

func dotEncodeValue(out string) string {
    el := strings.Replace(out, "\n", "\n ", -1)
    return strings.Replace(el, "\n \n", "\n .\n", -1)
}

func marshal(incoming reflect.Value) (string, error) {
	if incoming.Type().Kind() == reflect.Ptr {
		/* If we have a pointer, let's follow it */
		return marshal(incoming.Elem())
	}
    return marshalStruct(incoming)
}

func marshalStruct(incoming reflect.Value) (string, error) {
    ret := ""
	for i := 0; i < incoming.NumField(); i++ {
		field := incoming.Field(i)
		fieldType := incoming.Type().Field(i)

		paragraphKey := fieldType.Name
		if it := fieldType.Tag.Get("control"); it != "" {
			paragraphKey = it
		}

		if paragraphKey == "-" || fieldType.Anonymous {
			continue
		}

        val, err := marshalStructValue(field, fieldType)
        if err != nil {
            return "", err
        }
        val = dotEncodeValue(val)
        if val != "" {
            ret = ret + fmt.Sprintf("%s: %s\n", paragraphKey, val)
        }
    }
    return ret, nil
}

func marshalStructValueSlice(incoming reflect.Value, fieldType reflect.StructField) (string, error) {
	var delim = " "
	if it := fieldType.Tag.Get("delim"); it != "" {
		delim = it
	}

    ret := []string{}
    for i := 0; i < incoming.Len(); i++ {
        data, err := marshalStructValue(incoming.Index(i), fieldType)
        if err != nil {
            return "", err
        }
        ret = append(ret, data)
    }
    return strings.Join(ret, delim), nil
}

func marshalStructValueStruct(incoming reflect.Value) (string, error) {
	elem := incoming.Addr()

	if marshal, ok := elem.Interface().(Marshalable); ok {
		return marshal.MarshalControl()
	}
    return "", fmt.Errorf("%s doesn't implement the Marshalable interface",
                          elem.Type().Name())
}

func marshalStructValue(incoming reflect.Value, fieldType reflect.StructField) (string, error) {
	switch incoming.Type().Kind() {
	case reflect.String:
        return incoming.String(), nil
	case reflect.Int:
        return strconv.Itoa(int(incoming.Int())), nil
	case reflect.Slice:
		return marshalStructValueSlice(incoming, fieldType)
	case reflect.Struct:
		return marshalStructValueStruct(incoming)
	}
	return "", fmt.Errorf("Unknown type of field: %s", incoming.Type())
}

// vim: foldmethod=marker
