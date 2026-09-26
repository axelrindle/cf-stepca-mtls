package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"iter"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"text/template"

	"github.com/axelrindle/cf-stepca-mtls/internal/config"
)

const (
	configDir      = "internal/config"
	readmeTemplate = "README.md.tmpl"
	readmeOutput   = "README.md"
)

var tableTpl = template.Must(template.New("table").Parse(`| Variable | Required | Default | Description |
| --- | --- | --- | --- |
{{- range .Fields }}
| {{ if.Required }}**{{ .Name }}**{{ else }}{{ .Name }}{{ end }} | {{ if .Required }}**yes**{{ else }}no{{ end }} | {{ .Default }} | {{ .Description }} |
{{- end }}`))

func main() {
	docs, err := parseFieldDocs(configDir)
	if err != nil {
		panic(err)
	}

	t := reflect.TypeOf(config.Config{})

	fields := slices.Collect(collect(t, docs))
	slices.SortFunc(fields, func(a *envVar, b *envVar) int {
		return strings.Compare(a.Name, b.Name)
	})

	var table bytes.Buffer
	if err := tableTpl.Execute(&table, map[string]any{"Fields": fields}); err != nil {
		panic(err)
	}

	readmeTpl, err := template.ParseFiles(readmeTemplate)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(readmeOutput)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	if err := readmeTpl.Execute(out, map[string]any{
		"ConfigTable": table.String(),
	}); err != nil {
		panic(err)
	}
}

type fieldDocs map[string]map[string]string

// parseFieldDocs parses all Go files in dir and returns the doc comment of
// every struct field, keyed by struct type name and then field name.
func parseFieldDocs(dir string) (fieldDocs, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	docs := make(fieldDocs)

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}

				fieldDocs := make(map[string]string)
				for _, field := range st.Fields.List {
					if field.Doc == nil || len(field.Names) == 0 {
						continue
					}
					text := strings.TrimSpace(strings.ReplaceAll(field.Doc.Text(), "\n", "; "))
					fieldDocs[field.Names[0].Name] = text
				}

				docs[ts.Name.Name] = fieldDocs
			}
		}
	}

	return docs, nil
}

var kindSkip = []reflect.Kind{
	reflect.Chan,
	reflect.Func,
	reflect.Interface,
	reflect.UnsafePointer,
}

type envVar struct {
	Name        string
	Kind        string
	Required    bool
	Default     string
	Description string
}

func collect(t reflect.Type, docs fieldDocs) iter.Seq[*envVar] {
	return func(yield func(*envVar) bool) {
		fieldDocs := docs[t.Name()]

		for f := range t.Fields() {
			kind := f.Type.Kind()

			// resolve pointers before continuing
			if kind == reflect.Pointer {
				kind = f.Type.Elem().Kind()
			}

			if slices.Contains(kindSkip, kind) {
				continue
			}

			if kind == reflect.Struct {
				for c := range collect(f.Type, docs) {
					yield(c)
				}
			} else {
				env := f.Tag.Get("env")
				def, hasDef := f.Tag.Lookup("default")
				if !hasDef {
					def = "-"
				}

				yield(&envVar{
					Name:        env,
					Kind:        kind.String(),
					Required:    !hasDef,
					Default:     def,
					Description: fieldDocs[f.Name],
				})
			}
		}
	}
}
