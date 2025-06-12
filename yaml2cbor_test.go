// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
)

func Test_encodedCBOR(t *testing.T) {
	text := `
0: "foo"
1: 42
2:
  tag: 6
  value: "bar"
3:
  encodedCBOR:
    - tag: 6
      value: "buzz"
    - 32
    - test
`
	expected := []byte{
		0xa4,  // map size 4
			0x0, // key - 0
			0x63,  // value - tstr 3
				0x66, 0x6f, 0x6f, // "foo"
			0x1,  // key -  1
			0x18, 0x2a,  // value - uint 42 
			0x2, // key - 2
			0xc6,  // value - tag 6
				0x63, // tstr 3
					0x62, 0x61, 0x72, // "bar"
			0x3, // key - 3
			0x58, 0x0e, // value - bstr 14
				0x83, // array 3
					0xc6, // elt  0 - tag 6
						0x64, // tstr 4
							0x62, 0x75, 0x7a, 0x7a, // "buzz"
					0x18, 0x20, // elt 1 - uint 32
					0x64,  // elt 2 - tstr 4
						0x74, 0x65, 0x73, 0x74,  // "test"
	}

	out, err := yaml2cbor([]byte(text))
	assert.NoError(t, err)
	assertCBOREq(t, expected, out)

	data  := struct {
		Value []byte `cbor:"3,keyasint"`
	}{}

	err = dm.Unmarshal(out, &data)
	assert.NoError(t, err)
}

func Test_encodedCBOR_inside_tag(t *testing.T) {
    text :=`
tag: 25
value:
  encodedCBOR:
    0: "foo"
`
    expected := []byte{
        0xd8, 0x19, // tag(25)
          0x46, // bstr(6)
            0xa1, // map(1)
              0x00, // key: 0
              0x63, // value: tstr(3)
                0x66, 0x6f, 0x6f, // "foo"
    }

    out, err := yaml2cbor([]byte(text))
    assert.NoError(t, err)
    assertCBOREq(t, expected, out)
}

func Test_YAML_binary(t *testing.T) {
	text := `
0:
- !!binary |-
  dGVzdA==
`
	out, err := yaml2cbor([]byte(text))
	assert.NoError(t, err)

	data  := struct {
		Values [][]byte `cbor:"0,keyasint"`
	}{}
	err = dm.Unmarshal(out, &data)
	assert.NoError(t, err)

	assert.Equal(t, "test", string(data.Values[0]))
}

var dm, dmErr = cbor.DecOptions{
		IndefLength: cbor.IndefLengthForbidden,
		TimeTag:     cbor.DecTagRequired,
	}.DecMode()


func assertCBOREq(t *testing.T, expected []byte, actual []byte, msgAndArgs ...interface{}) bool {
	var expectedCBOR, actualCBOR interface{}

	if err := dm.Unmarshal([]byte(expected), &expectedCBOR); err != nil {
		return assert.Fail(t, fmt.Sprintf("Expected value ('%s') is not valid cbor.\nCBOR parsing error: '%s'", expected, err.Error()), msgAndArgs...)
	}

	if err := dm.Unmarshal([]byte(actual), &actualCBOR); err != nil {
		return assert.Fail(t, fmt.Sprintf("Input ('%s') needs to be valid cbor.\nCBOR parsing error: '%s'", actual, err.Error()), msgAndArgs...)
	}

	return assert.Equal(t, expectedCBOR, actualCBOR, msgAndArgs...)
}

func init() {
	if dmErr != nil {
		panic(dmErr)
	}
}
