package main

import (
	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

var em, emErr = cbor.EncOptions{
		IndefLength: cbor.IndefLengthForbidden,
		TimeTag:     cbor.EncTagRequired,
	}.EncMode()

func retag(v interface{}) (interface{}, bool) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return v, false
	}

	if len(m) !=  2 {
		return v, false
	}

	tag, ok := m["tag"]
	if !ok {
		return v, false
	}

	value, ok := m["value"]
	if !ok {
		return v, false
	}

	return cbor.Tag{
		Number: uint64(tag.(int)),
		Content: value,
	}, true
}

func retagRecursively(v interface{}) interface{} {
	out, retagged := retag(v)
	if retagged {
		tag := out.(cbor.Tag)
		tag.Content = retagRecursively(tag.Content)
		return tag
	}

	switch t := v.(type) {
	case map[interface{}]interface{}:
		updated := make(map[interface{}]interface{})
		for key, val := range t {
			updated[key] = retagRecursively(val)
		}
		return updated
	case map[string]interface{}:
		updated := make(map[string]interface{})
		for key, val := range t {
			updated[key] = retagRecursively(val)
		}
		return updated
	case []interface{}:
		var updated []interface{}
		for _, val := range t {
			updated = append(updated, retagRecursively(val))
		}
		return updated
	default:
		return v
	}
}

func encodeCBOR(v interface{}) (interface{}, bool, error) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return v, false, nil
	}

	if len(m) !=  1 {
		return v, false, nil
	}

	toEncode, ok := m["encodedCBOR"]
	if !ok {
		return v, false, nil
	}

	out, err := em.Marshal(toEncode)
	return out, true, err
}

func encodeCBORRecursively(v interface{}) (interface{}, error) {
	out, encoded, err := encodeCBOR(v)
	if err != nil {
		return nil, err
	}
	if encoded {
		return out, nil
	}

	switch t := v.(type) {
	case map[interface{}]interface{}:
		updated := make(map[interface{}]interface{})
		for key, val := range t {
			updated[key], err = encodeCBORRecursively(val)
			if err != nil {
				return nil, err
			}
		}
		return updated, nil
	case map[string]interface{}:
		updated := make(map[string]interface{})
		for key, val := range t {
			updated[key], err = encodeCBORRecursively(val)
			if err != nil {
				return nil, err
			}
		}
		return updated, nil
	case []interface{}:
		var updated []interface{}
		for _, val := range t {
			encodedVal, err := encodeCBORRecursively(val)
			if err != nil {
				return nil, err
			}
			updated = append(updated, encodedVal)
		}
		return updated, nil
	default:
		return v, nil
	}
}

func yaml2cbor(data []byte) ([]byte, error) {
        m := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(data), &m)
        if err != nil {
                return nil, err
        }

	retagged := retagRecursively(m)

	out, err := encodeCBORRecursively(retagged)
	if err != nil {
		return nil, err
	}

	return em.Marshal(out)
}

func init() {
	if emErr != nil {
		panic(emErr)
	}
}
