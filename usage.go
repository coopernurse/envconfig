// Copyright (c) 2016 Kelsey Hightower and others. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"encoding"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
)

const (
	// DefaultListFormat constant to use to display usage in a list format
	DefaultListFormat = `This application is configured via the environment. The following environment
variables can be used:
{{range .}}
{{usage_key .}}
  [description] {{usage_description .}}
  [type]        {{usage_type .}}
  [default]     {{usage_default .}}
  [required]    {{usage_required .}}{{end}}
`
	// DefaultTableFormat constant to use to display usage in a tabular format
	DefaultTableFormat = `This application is configured via the environment. The following environment
variables can be used:

KEY	TYPE	DEFAULT	REQUIRED	DESCRIPTION
{{range .}}{{usage_key .}}	{{usage_type .}}	{{usage_default .}}	{{usage_required .}}	{{usage_description .}}
{{end}}`
)

//nolint:gochecknoglobals
var (
	decoderType           = reflect.TypeOf((*Decoder)(nil)).Elem()
	setterType            = reflect.TypeOf((*Setter)(nil)).Elem()
	textUnmarshalerType   = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	binaryUnmarshalerType = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
)

type UsageOptions struct {
	Prefix     string
	SplitWords bool
	Out        io.Writer
	Format     string
	Template   *template.Template
}

func implementsInterface(t reflect.Type) bool {
	return t.Implements(decoderType) ||
		reflect.PtrTo(t).Implements(decoderType) ||
		t.Implements(setterType) ||
		reflect.PtrTo(t).Implements(setterType) ||
		t.Implements(textUnmarshalerType) ||
		reflect.PtrTo(t).Implements(textUnmarshalerType) ||
		t.Implements(binaryUnmarshalerType) ||
		reflect.PtrTo(t).Implements(binaryUnmarshalerType)
}

// toTypeDescription converts Go types into a human readable description
func toTypeDescription(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		return fmt.Sprintf("Comma-separated list of %s", toTypeDescription(t.Elem()))
	case reflect.Map:
		return fmt.Sprintf(
			"Comma-separated list of %s:%s pairs",
			toTypeDescription(t.Key()),
			toTypeDescription(t.Elem()),
		)
	case reflect.Ptr:
		return toTypeDescription(t.Elem())
	case reflect.Struct:
		if implementsInterface(t) && t.Name() != "" {
			return t.Name()
		}
		return ""
	case reflect.String:
		name := t.Name()
		if name != "" && name != "string" {
			return name
		}
		return "String"
	case reflect.Bool:
		name := t.Name()
		if name != "" && name != "bool" {
			return name
		}
		return "True or False"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		name := t.Name()
		if name != "" && !strings.HasPrefix(name, "int") {
			return name
		}
		return "Integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		name := t.Name()
		if name != "" && !strings.HasPrefix(name, "uint") {
			return name
		}
		return "Unsigned Integer"
	case reflect.Float32, reflect.Float64:
		name := t.Name()
		if name != "" && !strings.HasPrefix(name, "float") {
			return name
		}
		return "Float"
	}
	return fmt.Sprintf("%+v", t)
}

// Usage writes usage information to stderr using the default header and table format
func Usage(prefix string, spec interface{}) error {
	return UsageX(spec, Options{Prefix: prefix})
}

// Usagef writes usage information to the specified io.Writer using the specified template specification
func Usagef(prefix string, spec interface{}, out io.Writer, format string) error {
	usageOptions := UsageOptions{
		Prefix: prefix,
		Out:    out,
		Format: format,
	}

	return UsagefX(spec, usageOptions)
}

// Usaget writes usage information to the specified io.Writer using the specified template
func Usaget(prefix string, spec interface{}, out io.Writer, tmpl *template.Template) error {
	usageOptions := UsageOptions{
		Prefix:   prefix,
		Out:      out,
		Template: tmpl,
	}

	return UsagetX(spec, usageOptions)
}

func UsageX(spec interface{}, options Options) error {
	// The default is to output the usage information as a table
	// Create tabwriter instance to support table output
	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, 4, ' ', 0)

	usageOptions := UsageOptions{
		Prefix:     options.Prefix,
		SplitWords: options.SplitWords,
		Out:        tabs,
		Format:     DefaultTableFormat,
	}

	err := UsagefX(spec, usageOptions)
	tabs.Flush()
	return err
}

func UsagefX(spec interface{}, usageOptions UsageOptions) error {
	// Specify the default usage template functions
	functions := template.FuncMap{
		"usage_key":         func(v varInfo) string { return v.Key },
		"usage_description": func(v varInfo) string { return v.Tags.Get("desc") },
		"usage_type":        func(v varInfo) string { return toTypeDescription(v.Field.Type()) },
		"usage_default":     func(v varInfo) string { return v.Tags.Get("default") },
		"usage_required": func(v varInfo) (string, error) {
			req := v.Tags.Get("required")
			if req != "" {
				reqB, err := strconv.ParseBool(req)
				if err != nil {
					return "", err
				}
				if reqB {
					req = "true"
				}
			}
			return req, nil
		},
	}

	if usageOptions.Template == nil {
		tmpl, err := template.New("envconfig").Funcs(functions).Parse(usageOptions.Format)
		if err != nil {
			return err
		}

		usageOptions.Template = tmpl
	}

	return UsagetX(spec, usageOptions)
}

func UsagetX(spec interface{}, usageOptions UsageOptions) error {
	options := Options{
		Prefix:     usageOptions.Prefix,
		SplitWords: usageOptions.SplitWords,
	}

	infos, err := gatherInfo(spec, options)
	if err != nil {
		return err
	}

	return usageOptions.Template.Execute(usageOptions.Out, infos)
}
