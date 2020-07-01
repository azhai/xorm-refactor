// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package refactor

import (
	"strings"
	"text/template"

	"gitee.com/azhai/xorm-refactor/setting"
	"xorm.io/xorm/schemas"
)

var (
	languages       = make(map[string]*Language)
	presetTemplates = make(map[string]*template.Template)
)

type (
	Formatter func(fileName string, sourceCode []byte) ([]byte, error)
	Importter func(tables map[string]*schemas.Table) map[string]string
	Packager  func(targetDir string) string
)

// Language represents a languages supported when reverse codes
type Language struct {
	Name      string
	ExtName   string
	Template  string
	Types     map[string]string
	Funcs     template.FuncMap
	Formatter Formatter
	Importter Importter
	Packager  Packager
}

// RegisterLanguage registers a language
func RegisterLanguage(l *Language) {
	languages[l.Name] = l
}

// GetLanguage returns a language if exists
func GetLanguage(name string) *Language {
	return languages[name]
}

func (l *Language) FixTarget(target *setting.ReverseTarget) {
	if target.ExtName == "" && l.ExtName != "" {
		if !strings.HasPrefix(l.ExtName, ".") {
			l.ExtName = "." + l.ExtName
		}
		target.ExtName = l.ExtName
	}
	if target.NameSpace == "" {
		if pck := l.Packager; pck != nil {
			target.NameSpace = pck(target.OutputDir)
		}
		if target.NameSpace == "" {
			target.NameSpace = "models"
		}
	}
}

func NewTemplate(name, content string, funcs template.FuncMap) *template.Template {
	t := template.New(name).Funcs(funcs)
	tmpl, err := t.Parse(content)
	if err != nil {
		panic(err)
	}
	presetTemplates[name] = tmpl
	return tmpl
}

func GetPresetTemplate(name string) *template.Template {
	if tmpl, ok := presetTemplates[name]; ok {
		return tmpl
	}
	return nil
}
