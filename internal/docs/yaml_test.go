package docs_test

import (
	"fmt"
	"testing"

	"github.com/Jeffail/benthos/v3/internal/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestFieldsFromNode(t *testing.T) {
	tests := []struct {
		name   string
		yaml   string
		fields docs.FieldSpecs
	}{
		{
			name: "flat object",
			yaml: `a: foo
b: bar
c: 21`,
			fields: docs.FieldSpecs{
				docs.FieldString("a", "").HasDefault("foo"),
				docs.FieldString("b", "").HasDefault("bar"),
				docs.FieldInt("c", "").HasDefault(int64(21)),
			},
		},
		{
			name: "nested object",
			yaml: `a: foo
b:
  d: bar
  e: 22
c: true`,
			fields: docs.FieldSpecs{
				docs.FieldString("a", "").HasDefault("foo"),
				docs.FieldCommon("b", "").WithChildren(
					docs.FieldString("d", "").HasDefault("bar"),
					docs.FieldInt("e", "").HasDefault(int64(22)),
				),
				docs.FieldBool("c", "").HasDefault(true),
			},
		},
		{
			name: "array of strings",
			yaml: `a:
- foo`,
			fields: docs.FieldSpecs{
				docs.FieldString("a", "").Array().HasDefault([]string{"foo"}),
			},
		},
		{
			name: "array of ints",
			yaml: `a:
- 5
- 8`,
			fields: docs.FieldSpecs{
				docs.FieldInt("a", "").Array().HasDefault([]int64{5, 8}),
			},
		},
		{
			name: "nested array of strings",
			yaml: `a:
  b:
    - foo
    - bar`,
			fields: docs.FieldSpecs{
				docs.FieldCommon("a", "").WithChildren(
					docs.FieldString("b", "").Array().HasDefault([]string{"foo", "bar"}),
				),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			confBytes := []byte(test.yaml)

			var node yaml.Node
			require.NoError(t, yaml.Unmarshal(confBytes, &node))

			assert.Equal(t, test.fields, docs.FieldsFromYAML(&node))
		})
	}
}

func TestFieldsNodeToMap(t *testing.T) {
	spec := docs.FieldSpecs{
		docs.FieldCommon("a", ""),
		docs.FieldCommon("b", "").HasDefault(11),
		docs.FieldCommon("c", "").WithChildren(
			docs.FieldCommon("d", "").HasDefault(true),
			docs.FieldCommon("e", "").HasDefault("evalue"),
			docs.FieldCommon("f", "").WithChildren(
				docs.FieldCommon("g", "").HasDefault(12),
				docs.FieldCommon("h", ""),
				docs.FieldCommon("i", "").HasDefault(13),
			),
		),
	}

	var node yaml.Node
	err := yaml.Unmarshal([]byte(`
a: setavalue
c:
  f:
    g: 22
    h: sethvalue
    i: 23.1
`), &node)
	require.NoError(t, err)

	generic, err := spec.YAMLToMap(false, &node)
	require.NoError(t, err)

	assert.Equal(t, map[string]interface{}{
		"a": "setavalue",
		"b": 11,
		"c": map[string]interface{}{
			"d": true,
			"e": "evalue",
			"f": map[string]interface{}{
				"g": 22,
				"h": "sethvalue",
				"i": 23.1,
			},
		},
	}, generic)
}

func TestFieldsNodeToMapTypeCoercion(t *testing.T) {
	tests := []struct {
		name   string
		spec   docs.FieldSpecs
		yaml   string
		result interface{}
	}{
		{
			name: "string fields",
			spec: docs.FieldSpecs{
				docs.FieldCommon("a", "").HasType("string"),
				docs.FieldCommon("b", "").HasType("string"),
				docs.FieldCommon("c", "").HasType("string"),
				docs.FieldCommon("d", "").HasType("string"),
				docs.FieldCommon("e", "").HasType("string").Array(),
				docs.FieldCommon("f", "").HasType("string").Map(),
			},
			yaml: `
a: no
b: false
c: 10
d: 30.4
e:
 - no
 - false
 - 10
f:
 "1": no
 "2": false
 "3": 10
`,
			result: map[string]interface{}{
				"a": "no",
				"b": "false",
				"c": "10",
				"d": "30.4",
				"e": []interface{}{
					"no", "false", "10",
				},
				"f": map[string]interface{}{
					"1": "no", "2": "false", "3": "10",
				},
			},
		},
		{
			name: "bool fields",
			spec: docs.FieldSpecs{
				docs.FieldCommon("a", "").HasType("bool"),
				docs.FieldCommon("b", "").HasType("bool"),
				docs.FieldCommon("c", "").HasType("bool"),
				docs.FieldCommon("d", "").HasType("bool").Array(),
				docs.FieldCommon("e", "").HasType("bool").Map(),
			},
			yaml: `
a: no
b: false
c: true
d:
 - no
 - false
 - true
e:
 "1": no
 "2": false
 "3": true
`,
			result: map[string]interface{}{
				"a": false,
				"b": false,
				"c": true,
				"d": []interface{}{
					false, false, true,
				},
				"e": map[string]interface{}{
					"1": false, "2": false, "3": true,
				},
			},
		},
		{
			name: "int fields",
			spec: docs.FieldSpecs{
				docs.FieldCommon("a", "").HasType("int"),
				docs.FieldCommon("b", "").HasType("int"),
				docs.FieldCommon("c", "").HasType("int"),
				docs.FieldCommon("d", "").HasType("int").Array(),
				docs.FieldCommon("e", "").HasType("int").Map(),
			},
			yaml: `
a: 11
b: -12
c: 13.4
d:
 - 11
 - -12
 - 13.4
e:
 "1": 11
 "2": -12
 "3": 13.4
`,
			result: map[string]interface{}{
				"a": 11,
				"b": -12,
				"c": 13,
				"d": []interface{}{
					11, -12, 13,
				},
				"e": map[string]interface{}{
					"1": 11, "2": -12, "3": 13,
				},
			},
		},
		{
			name: "float fields",
			spec: docs.FieldSpecs{
				docs.FieldCommon("a", "").HasType("float"),
				docs.FieldCommon("b", "").HasType("float"),
				docs.FieldCommon("c", "").HasType("float"),
				docs.FieldCommon("d", "").HasType("float").Array(),
				docs.FieldCommon("e", "").HasType("float").Map(),
			},
			yaml: `
a: 11
b: -12
c: 13.4
d:
 - 11
 - -12
 - 13.4
e:
 "1": 11
 "2": -12
 "3": 13.4
`,
			result: map[string]interface{}{
				"a": 11.0,
				"b": -12.0,
				"c": 13.4,
				"d": []interface{}{
					11.0, -12.0, 13.4,
				},
				"e": map[string]interface{}{
					"1": 11.0, "2": -12.0, "3": 13.4,
				},
			},
		},
		{
			name: "recurse array of objects",
			spec: docs.FieldSpecs{
				docs.FieldCommon("foo", "").WithChildren(
					docs.FieldCommon("eles", "").Array().WithChildren(
						docs.FieldCommon("bar", "").HasType(docs.FieldTypeString).HasDefault("default"),
					),
				),
			},
			yaml: `
foo:
  eles:
    - bar: bar1
    - bar: bar2
`,
			result: map[string]interface{}{
				"foo": map[string]interface{}{
					"eles": []interface{}{
						map[string]interface{}{
							"bar": "bar1",
						},
						map[string]interface{}{
							"bar": "bar2",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(test.yaml), &node)
			require.NoError(t, err)

			generic, err := test.spec.YAMLToMap(false, &node)
			require.NoError(t, err)

			assert.Equal(t, test.result, generic)
		})
	}
}

func TestFieldToNode(t *testing.T) {
	tests := []struct {
		name     string
		spec     docs.FieldSpec
		recurse  bool
		expected string
	}{
		{
			name: "no recurse single node null",
			spec: docs.FieldCommon("foo", ""),
			expected: `null
`,
		},
		{
			name: "no recurse with children",
			spec: docs.FieldCommon("foo", "").WithChildren(
				docs.FieldCommon("bar", ""),
				docs.FieldCommon("baz", ""),
			),
			expected: `{}
`,
		},
		{
			name: "no recurse map",
			spec: docs.FieldCommon("foo", "").Map(),
			expected: `{}
`,
		},
		{
			name: "recurse with children",
			spec: docs.FieldCommon("foo", "").WithChildren(
				docs.FieldCommon("bar", "").HasType(docs.FieldTypeString),
				docs.FieldCommon("baz", "").HasType(docs.FieldTypeString).HasDefault("baz default"),
				docs.FieldCommon("buz", "").HasType(docs.FieldTypeInt),
				docs.FieldCommon("bev", "").HasType(docs.FieldTypeFloat),
				docs.FieldCommon("bun", "").HasType(docs.FieldTypeBool),
				docs.FieldCommon("bud", "").Array(),
			),
			recurse: true,
			expected: `bar: ""
baz: baz default
buz: 0
bev: 0
bun: false
bud: []
`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			n, err := test.spec.ToYAML(test.recurse)
			require.NoError(t, err)

			b, err := yaml.Marshal(n)
			require.NoError(t, err)

			assert.Equal(t, test.expected, string(b))
		})
	}
}

func TestYAMLComponentLinting(t *testing.T) {
	for _, t := range docs.Types() {
		docs.RegisterDocs(docs.ComponentSpec{
			Name: fmt.Sprintf("testlintfoo%v", string(t)),
			Type: t,
			Config: docs.FieldComponent().WithChildren(
				docs.FieldString("foo1", "").Linter(func(ctx docs.LintContext, line, col int, v interface{}) []docs.Lint {
					if v == "lint me please" {
						return []docs.Lint{
							docs.NewLintError(line, "this is a custom lint"),
						}
					}
					return nil
				}).Optional(),
				docs.FieldString("foo2", "").Advanced().OmitWhen(func(field, parent interface{}) (string, bool) {
					if field == "drop me" {
						return "because foo", true
					}
					return "", false
				}).Optional(),
				docs.FieldCommon("foo3", "").HasType(docs.FieldTypeProcessor).Optional(),
				docs.FieldAdvanced("foo4", "").Array().HasType(docs.FieldTypeProcessor).Optional(),
				docs.FieldCommon("foo5", "").Map().HasType(docs.FieldTypeProcessor).Optional(),
				docs.FieldDeprecated("foo6").Optional(),
				docs.FieldAdvanced("foo7", "").Array().WithChildren(
					docs.FieldString("foochild1", "").Optional(),
				).Optional(),
				docs.FieldAdvanced("foo8", "").Map().WithChildren(
					docs.FieldInt("foochild1", "").Optional(),
				).Optional(),
			),
		})
	}

	type testCase struct {
		name      string
		inputType docs.Type
		inputConf string

		res []docs.Lint
	}

	tests := []testCase{
		{
			name:      "ignores comments",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  # comment here
  foo1: hello world # And what's this?`,
		},
		{
			name:      "allows anchors",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput: &test-anchor
  foo1: hello world
processors:
  - testlintfooprocessor: *test-anchor`,
		},
		{
			name:      "lints through anchors",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput: &test-anchor
  foo1: hello world
  nope: bad field
processors:
  - testlintfooprocessor: *test-anchor`,
			res: []docs.Lint{
				docs.NewLintError(4, "field nope not recognised"),
			},
		},
		{
			name:      "unknown fields",
			inputType: docs.TypeInput,
			inputConf: `
type: testlintfooinput
testlintfooinput:
  not_recognised: yuh
  foo1: hello world
  also_not_recognised: nah
definitely_not_recognised: huh`,
			res: []docs.Lint{
				docs.NewLintError(4, "field not_recognised not recognised"),
				docs.NewLintError(6, "field also_not_recognised not recognised"),
				docs.NewLintError(7, "field definitely_not_recognised is invalid when the component type is testlintfooinput (input)"),
			},
		},
		{
			name:      "reserved field unknown fields",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  not_recognised: yuh
  foo1: hello world
processors:
  - testlintfooprocessor:
      also_not_recognised: nah`,
			res: []docs.Lint{
				docs.NewLintError(3, "field not_recognised not recognised"),
				docs.NewLintError(7, "field also_not_recognised not recognised"),
			},
		},
		{
			name:      "collision of labels",
			inputType: docs.TypeInput,
			inputConf: `
label: foo
testlintfooinput:
  foo1: hello world
processors:
  - label: bar
    testlintfooprocessor: {}
  - label: foo
    testlintfooprocessor: {}`,
			res: []docs.Lint{
				docs.NewLintError(8, "Label 'foo' collides with a previously defined label at line 2"),
			},
		},
		{
			name:      "empty processors",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo1: hello world
processors: []`,
			res: []docs.Lint{
				docs.NewLintError(4, "field processors is empty and can be removed"),
			},
		},
		{
			name:      "custom omit func",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo1: hello world
  foo2: drop me`,
			res: []docs.Lint{
				docs.NewLintError(4, "because foo"),
			},
		},
		{
			name:      "nested array not an array",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo4:
    key1:
      testlintfooprocessor:
        foo1: somevalue
        not_recognised: nah`,
			res: []docs.Lint{
				docs.NewLintError(4, "expected array value"),
			},
		},
		{
			name:      "nested fields",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo3:
    testlintfooprocessor:
      foo1: somevalue
      not_recognised: nah`,
			res: []docs.Lint{
				docs.NewLintError(6, "field not_recognised not recognised"),
			},
		},
		{
			name:      "array for string",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo3:
    testlintfooprocessor:
      foo1: [ somevalue ]
`,
			res: []docs.Lint{
				docs.NewLintError(5, "expected string value"),
			},
		},
		{
			name:      "nested map fields",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo5:
    key1:
      testlintfooprocessor:
        foo1: somevalue
        not_recognised: nah`,
			res: []docs.Lint{
				docs.NewLintError(7, "field not_recognised not recognised"),
			},
		},
		{
			name:      "nested map not a map",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo5:
    - testlintfooprocessor:
        foo1: somevalue
        not_recognised: nah`,
			res: []docs.Lint{
				docs.NewLintError(4, "expected object value"),
			},
		},
		{
			name:      "array field",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo7:
   - foochild1: yep`,
		},
		{
			name:      "array field bad",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo7:
   - wat: no`,
			res: []docs.Lint{
				docs.NewLintError(4, "field wat not recognised"),
			},
		},
		{
			name:      "array field not array",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo7:
    key1:
      wat: no`,
			res: []docs.Lint{
				docs.NewLintError(4, "expected array value"),
			},
		},
		{
			name:      "map field",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo8:
    key1:
      foochild1: 10`,
		},
		{
			name:      "map field bad",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo8:
    key1:
      wat: nope`,
			res: []docs.Lint{
				docs.NewLintError(5, "field wat not recognised"),
			},
		},
		{
			name:      "map field not map",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo8:
    - wat: nope`,
			res: []docs.Lint{
				docs.NewLintError(4, "expected object value"),
			},
		},
		{
			name:      "custom lint",
			inputType: docs.TypeInput,
			inputConf: `
testlintfooinput:
  foo1: lint me please`,
			res: []docs.Lint{
				docs.NewLintError(3, "this is a custom lint"),
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var node yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(test.inputConf), &node))
			lints := docs.LintYAML(docs.NewLintContext(), test.inputType, &node)
			assert.Equal(t, test.res, lints)
		})
	}
}

func TestYAMLLinting(t *testing.T) {
	type testCase struct {
		name      string
		inputSpec docs.FieldSpec
		inputConf string

		res []docs.Lint
	}

	tests := []testCase{
		{
			name:      "expected string got array",
			inputSpec: docs.FieldString("foo", ""),
			inputConf: `["foo","bar"]`,
			res: []docs.Lint{
				docs.NewLintError(1, "expected string value"),
			},
		},
		{
			name:      "expected array got string",
			inputSpec: docs.FieldString("foo", "").Array(),
			inputConf: `"foo"`,
			res: []docs.Lint{
				docs.NewLintError(1, "expected array value"),
			},
		},
		{
			name: "expected object got string",
			inputSpec: docs.FieldCommon("foo", "").WithChildren(
				docs.FieldString("bar", ""),
			),
			inputConf: `"foo"`,
			res: []docs.Lint{
				docs.NewLintError(1, "expected object value"),
			},
		},
		{
			name: "expected string got object",
			inputSpec: docs.FieldCommon("foo", "").WithChildren(
				docs.FieldString("bar", ""),
			),
			inputConf: `bar: {}`,
			res: []docs.Lint{
				docs.NewLintError(1, "expected string value"),
			},
		},
		{
			name: "expected string got object nested",
			inputSpec: docs.FieldCommon("foo", "").WithChildren(
				docs.FieldCommon("bar", "").WithChildren(
					docs.FieldString("baz", ""),
				),
			),
			inputConf: `bar:
  baz: {}`,
			res: []docs.Lint{
				docs.NewLintError(2, "expected string value"),
			},
		},
		{
			name: "missing non-optional field",
			inputSpec: docs.FieldCommon("foo", "").WithChildren(
				docs.FieldString("bar", "").HasDefault("barv"),
				docs.FieldString("baz", ""),
				docs.FieldString("buz", "").Optional(),
				docs.FieldString("bev", ""),
			),
			inputConf: `bev: hello world`,
			res: []docs.Lint{
				docs.NewLintError(1, "field baz is required"),
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var node yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(test.inputConf), &node))

			lints := test.inputSpec.LintYAML(docs.NewLintContext(), &node)
			assert.Equal(t, test.res, lints)
		})
	}
}

func TestYAMLSanitation(t *testing.T) {
	for _, t := range docs.Types() {
		docs.RegisterDocs(docs.ComponentSpec{
			Name: fmt.Sprintf("testyamlsanitfoo%v", string(t)),
			Type: t,
			Config: docs.FieldComponent().WithChildren(
				docs.FieldCommon("foo1", ""),
				docs.FieldAdvanced("foo2", ""),
				docs.FieldCommon("foo3", "").HasType(docs.FieldTypeProcessor),
				docs.FieldAdvanced("foo4", "").Array().HasType(docs.FieldTypeProcessor),
				docs.FieldCommon("foo5", "").Map().HasType(docs.FieldTypeProcessor),
				docs.FieldDeprecated("foo6"),
			),
		})
		docs.RegisterDocs(docs.ComponentSpec{
			Name: fmt.Sprintf("testyamlsanitbar%v", string(t)),
			Type: t,
			Config: docs.FieldComponent().Array().WithChildren(
				docs.FieldCommon("bar1", ""),
				docs.FieldAdvanced("bar2", ""),
				docs.FieldCommon("bar3", "").HasType(docs.FieldTypeProcessor),
			),
		})
		docs.RegisterDocs(docs.ComponentSpec{
			Name: fmt.Sprintf("testyamlsanitbaz%v", string(t)),
			Type: t,
			Config: docs.FieldComponent().Map().WithChildren(
				docs.FieldCommon("baz1", ""),
				docs.FieldAdvanced("baz2", ""),
				docs.FieldCommon("baz3", "").HasType(docs.FieldTypeProcessor),
			),
		})
	}

	type testCase struct {
		name        string
		inputType   docs.Type
		inputConf   string
		inputFilter func(f docs.FieldSpec) bool

		res string
		err string
	}

	tests := []testCase{
		{
			name:      "input with processors",
			inputType: docs.TypeInput,
			inputConf: `testyamlsanitfooinput:
  foo1: simple field
  foo2: advanced field
  foo6: deprecated field
someotherinput:
  ignore: me please
processors:
  - testyamlsanitbarprocessor:
      bar1: bar value
      bar5: undocumented field
    someotherprocessor:
      ignore: me please
`,
			res: `testyamlsanitfooinput:
    foo1: simple field
    foo2: advanced field
    foo6: deprecated field
processors:
    - testyamlsanitbarprocessor:
        bar1: bar value
        bar5: undocumented field
`,
		},
		{
			name:      "output array with nested map processor",
			inputType: docs.TypeOutput,
			inputConf: `testyamlsanitbaroutput:
    - bar1: simple field
      bar3:
          testyamlsanitbazprocessor:
              customkey1:
                  baz1: simple field
          someotherprocessor:
             ignore: me please
    - bar2: advanced field
`,
			res: `testyamlsanitbaroutput:
    - bar1: simple field
      bar3:
        testyamlsanitbazprocessor:
            customkey1:
                baz1: simple field
    - bar2: advanced field
`,
		},
		{
			name:      "output with empty processors",
			inputType: docs.TypeOutput,
			inputConf: `testyamlsanitbaroutput:
    - bar1: simple field
processors: []
`,
			res: `testyamlsanitbaroutput:
    - bar1: simple field
`,
		},
		{
			name:      "metrics map with nested map processor",
			inputType: docs.TypeMetrics,
			inputConf: `testyamlsanitbazmetrics:
  customkey1:
    baz1: simple field
    baz3:
      testyamlsanitbazprocessor:
        customkey1:
          baz1: simple field
      someotherprocessor:
        ignore: me please
  customkey2:
    baz2: advanced field
`,
			res: `testyamlsanitbazmetrics:
    customkey1:
        baz1: simple field
        baz3:
            testyamlsanitbazprocessor:
                customkey1:
                    baz1: simple field
    customkey2:
        baz2: advanced field
`,
		},
		{
			name:      "ratelimit with array field processor",
			inputType: docs.TypeRateLimit,
			inputConf: `testyamlsanitfoorate_limit:
    foo1: simple field
    foo4:
      - testyamlsanitbazprocessor:
            customkey1:
                baz1: simple field
        someotherprocessor:
            ignore: me please
`,
			res: `testyamlsanitfoorate_limit:
    foo1: simple field
    foo4:
        - testyamlsanitbazprocessor:
            customkey1:
                baz1: simple field
`,
		},
		{
			name:      "ratelimit with map field processor",
			inputType: docs.TypeRateLimit,
			inputConf: `testyamlsanitfoorate_limit:
    foo1: simple field
    foo5:
        customkey1:
            testyamlsanitbazprocessor:
                customkey1:
                    baz1: simple field
            someotherprocessor:
                ignore: me please
`,
			res: `testyamlsanitfoorate_limit:
    foo1: simple field
    foo5:
        customkey1:
            testyamlsanitbazprocessor:
                customkey1:
                    baz1: simple field
`,
		},
		{
			name:        "input with processors no deprecated",
			inputType:   docs.TypeInput,
			inputFilter: docs.ShouldDropDeprecated(true),
			inputConf: `testyamlsanitfooinput:
    foo1: simple field
    foo2: advanced field
    foo6: deprecated field
someotherinput:
    ignore: me please
processors:
    - testyamlsanitfooprocessor:
        foo1: simple field
        foo2: advanced field
        foo6: deprecated field
      someotherprocessor:
        ignore: me please
`,
			res: `testyamlsanitfooinput:
    foo1: simple field
    foo2: advanced field
processors:
    - testyamlsanitfooprocessor:
        foo1: simple field
        foo2: advanced field
`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var node yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(test.inputConf), &node))
			err := docs.SanitiseYAML(test.inputType, &node, docs.SanitiseConfig{
				RemoveTypeField:  true,
				Filter:           test.inputFilter,
				RemoveDeprecated: false,
			})
			if len(test.err) > 0 {
				assert.EqualError(t, err, test.err)
			} else {
				assert.NoError(t, err)

				resBytes, err := yaml.Marshal(node.Content[0])
				require.NoError(t, err)
				assert.Equal(t, test.res, string(resBytes))
			}
		})
	}
}
