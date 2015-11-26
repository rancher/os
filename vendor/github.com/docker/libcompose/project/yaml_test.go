package project

import (
	"testing"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/stretchr/testify/assert"
)

type structStringorslice struct {
	Foo Stringorslice `yaml:"foo,flow,omitempty"`
}

func TestMarshal(t *testing.T) {
	s := &structStringorslice{Foo: NewStringorslice("a", "b", "c")}
	b, err := yaml.Marshal(s)
	assert.Equal(t, "foo: [a, b, c]\n", string(b))
	assert.Nil(t, err)
}

func TestMarshalEmpty(t *testing.T) {
	s := &structStringorslice{}
	b, err := yaml.Marshal(s)
	assert.Equal(t, "foo: []\n", string(b))
	assert.Nil(t, err)
}

func TestUnmarshalSlice(t *testing.T) {
	expected := &structStringorslice{Foo: NewStringorslice("a", "b", "c")}
	b := []byte("foo: [a, b, c]\n")
	s := &structStringorslice{}
	err := yaml.Unmarshal(b, s)
	assert.Equal(t, expected, s)
	assert.Nil(t, err)
}

func TestUnmarshalString(t *testing.T) {
	expected := &structStringorslice{Foo: NewStringorslice("abc")}
	b := []byte("foo: abc\n")
	s := &structStringorslice{}
	err := yaml.Unmarshal(b, s)
	assert.Equal(t, expected, s)
	assert.Nil(t, err)
}

func TestUnmarshalEmpty(t *testing.T) {
	expected := &structStringorslice{Foo: NewStringorslice()}
	b := []byte("{}\n")
	s := &structStringorslice{}
	err := yaml.Unmarshal(b, s)
	assert.Equal(t, expected, s)
	assert.Nil(t, err)
}
