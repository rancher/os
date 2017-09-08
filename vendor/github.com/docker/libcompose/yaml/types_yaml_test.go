package yaml

import (
	"fmt"
	"strings"
	"testing"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/stretchr/testify/assert"
)

type StructStringorslice struct {
	Foo Stringorslice
}

func TestStringorsliceYaml(t *testing.T) {
	str := `{foo: [bar, baz]}`

	s := StructStringorslice{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, Stringorslice{"bar", "baz"}, s.Foo)

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructStringorslice{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, Stringorslice{"bar", "baz"}, s2.Foo)
}

type StructSliceorMap struct {
	Foos SliceorMap `yaml:"foos,omitempty"`
	Bars []string   `yaml:"bars"`
}

type StructCommand struct {
	Entrypoint Command `yaml:"entrypoint,flow,omitempty"`
	Command    Command `yaml:"command,flow,omitempty"`
}

func TestSliceOrMapYaml(t *testing.T) {
	str := `{foos: [bar=baz, far=faz]}`

	s := StructSliceorMap{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, SliceorMap{"bar": "baz", "far": "faz"}, s.Foos)

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructSliceorMap{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, SliceorMap{"bar": "baz", "far": "faz"}, s2.Foos)
}

var sampleStructSliceorMap = `
foos:
  io.rancher.os.bar: baz
  io.rancher.os.far: true
bars: []
`

func TestUnmarshalSliceOrMap(t *testing.T) {
	s := StructSliceorMap{}
	err := yaml.Unmarshal([]byte(sampleStructSliceorMap), &s)
	assert.Equal(t, fmt.Errorf("Cannot unmarshal 'true' of type bool into a string value"), err)
}

func TestStr2SliceOrMapPtrMap(t *testing.T) {
	s := map[string]*StructSliceorMap{"udav": {
		Foos: SliceorMap{"io.rancher.os.bar": "baz", "io.rancher.os.far": "true"},
		Bars: []string{},
	}}
	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := map[string]*StructSliceorMap{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, s, s2)
}

type StructMaporslice struct {
	Foo MaporEqualSlice
}

func contains(list []string, item string) bool {
	for _, test := range list {
		if test == item {
			return true
		}
	}
	return false
}

func TestMaporsliceYaml(t *testing.T) {
	str := `{foo: {bar: baz, far: 1}}`

	s := StructMaporslice{}
	yaml.Unmarshal([]byte(str), &s)

	assert.Equal(t, 2, len(s.Foo))
	assert.True(t, contains(s.Foo, "bar=baz"))
	assert.True(t, contains(s.Foo, "far=1"))

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructMaporslice{}
	yaml.Unmarshal(d, &s2)

	assert.Equal(t, 2, len(s2.Foo))
	assert.True(t, contains(s2.Foo, "bar=baz"))
	assert.True(t, contains(s2.Foo, "far=1"))
}

var sampleStructCommand = `command: bash`

func TestUnmarshalCommand(t *testing.T) {
	s := &StructCommand{}
	err := yaml.Unmarshal([]byte(sampleStructCommand), s)

	assert.Nil(t, err)
	assert.Equal(t, Command{"bash"}, s.Command)
	assert.Nil(t, s.Entrypoint)
	bytes, err := yaml.Marshal(s)
	assert.Nil(t, err)

	s2 := &StructCommand{}
	err = yaml.Unmarshal(bytes, s2)

	assert.Nil(t, err)
	assert.Equal(t, Command{"bash"}, s2.Command)
	assert.Nil(t, s2.Entrypoint)
}

var sampleEmptyCommand = `{}`

func TestUnmarshalEmptyCommand(t *testing.T) {
	s := &StructCommand{}
	err := yaml.Unmarshal([]byte(sampleEmptyCommand), s)

	assert.Nil(t, err)
	assert.Nil(t, s.Command)

	bytes, err := yaml.Marshal(s)
	assert.Nil(t, err)
	assert.Equal(t, "{}", strings.TrimSpace(string(bytes)))

	s2 := &StructCommand{}
	err = yaml.Unmarshal(bytes, s2)

	assert.Nil(t, err)
	assert.Nil(t, s2.Command)
}

func TestMarshalUlimit(t *testing.T) {
	ulimits := []struct {
		ulimits  *Ulimits
		expected string
	}{
		{
			ulimits: &Ulimits{
				Elements: []Ulimit{
					{
						ulimitValues: ulimitValues{
							Soft: 65535,
							Hard: 65535,
						},
						Name: "nproc",
					},
				},
			},
			expected: `nproc: 65535
`,
		},
		{
			ulimits: &Ulimits{
				Elements: []Ulimit{
					{
						Name: "nofile",
						ulimitValues: ulimitValues{
							Soft: 20000,
							Hard: 40000,
						},
					},
				},
			},
			expected: `nofile:
  soft: 20000
  hard: 40000
`,
		},
	}

	for _, ulimit := range ulimits {

		bytes, err := yaml.Marshal(ulimit.ulimits)

		assert.Nil(t, err)
		assert.Equal(t, ulimit.expected, string(bytes), "should be equal")
	}
}

func TestUnmarshalUlimits(t *testing.T) {
	ulimits := []struct {
		yaml     string
		expected *Ulimits
	}{
		{
			yaml: "nproc: 65535",
			expected: &Ulimits{
				Elements: []Ulimit{
					{
						Name: "nproc",
						ulimitValues: ulimitValues{
							Soft: 65535,
							Hard: 65535,
						},
					},
				},
			},
		},
		{
			yaml: `nofile:
  soft: 20000
  hard: 40000`,
			expected: &Ulimits{
				Elements: []Ulimit{
					{
						Name: "nofile",
						ulimitValues: ulimitValues{
							Soft: 20000,
							Hard: 40000,
						},
					},
				},
			},
		},
		{
			yaml: `nproc: 65535
nofile:
  soft: 20000
  hard: 40000`,
			expected: &Ulimits{
				Elements: []Ulimit{
					{
						Name: "nofile",
						ulimitValues: ulimitValues{
							Soft: 20000,
							Hard: 40000,
						},
					},
					{
						Name: "nproc",
						ulimitValues: ulimitValues{
							Soft: 65535,
							Hard: 65535,
						},
					},
				},
			},
		},
	}

	for _, ulimit := range ulimits {
		actual := &Ulimits{}
		err := yaml.Unmarshal([]byte(ulimit.yaml), actual)

		assert.Nil(t, err)
		assert.Equal(t, ulimit.expected, actual, "should be equal")
	}
}
