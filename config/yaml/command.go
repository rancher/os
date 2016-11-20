package yaml

import "fmt"

// StringandSlice stores either a string or slice depending on original type
// Differs from libcompose Stringorslice by being able to determine original type
type StringandSlice struct {
	StringValue string
	SliceValue  []string
}

// UnmarshalYAML implements the Unmarshaller interface.
// TODO: this needs to be ported to go-yaml
func (s *StringandSlice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		s.SliceValue = parts
	case string:
		s.StringValue = value
	default:
		return fmt.Errorf("Failed to unmarshal StringandSlice: %#v", value)
	}
	return nil
}

// TODO: use this function from libcompose
func toStrings(s []interface{}) ([]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	r := make([]string, len(s))
	for k, v := range s {
		if sv, ok := v.(string); ok {
			r[k] = sv
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
		}
	}
	return r, nil
}
