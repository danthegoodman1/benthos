package template

import (
	"bytes"
	"text/template"

	"github.com/Jeffail/benthos/v3/internal/docs"
)

// TODO: Put this somewhere else and use the new Go 1.16 file APIs.
var templateDocsTemplate = docs.FieldsTemplate(false) + `---
title: Templating
description: Learn how Benthos templates work.
---

<!--
     THIS FILE IS AUTOGENERATED!

     To make changes please edit the contents of:
     internal/docs/template_docs.go
-->

EXPERIMENTAL: Templates are an experimental feature and therefore subject to change (or removal) outside of major version releases.

Templates are a way to define new Benthos components (similar to plugins) that are implemented by generating a Benthos config snippet from pre-defined parameter fields. This is useful when a common pattern of Benthos configuration is used but with varying parameters each time.

Templates will be available to try in version 3.47.0. The release date is yet to be determined, but in the meantime you can grab release candidates from [the releases page](https://github.com/Jeffail/benthos/releases) or pull nightly builds from docker with the ` + "`jeffail/benthos:edge`" + ` tag.

A template is defined in a YAML file that can be imported when Benthos runs using the flag ` + "`-t`" + `:

` + "```sh" + `
benthos -t "./templates/*.yaml" -c ./config.yaml
` + "```" + `

You can see examples of templates, including some that are included as part of the standard Benthos distribution, at [https://github.com/Jeffail/benthos/tree/master/template](https://github.com/Jeffail/benthos/tree/master/template).

## Fields

The schema of a template file is as follows:

{{template "field_docs" . -}}
`

type templateContext struct {
	Fields []docs.FieldSpecCtx
}

// DocsMarkdown returns a markdown document for the templates documentation.
func DocsMarkdown() ([]byte, error) {
	var buf bytes.Buffer
	err := template.Must(template.New("templates").Parse(templateDocsTemplate)).Execute(&buf, templateContext{
		Fields: docs.FieldCommon("", "").WithChildren(ConfigSpec()...).FlattenChildrenForDocs(),
	})

	return buf.Bytes(), err
}
