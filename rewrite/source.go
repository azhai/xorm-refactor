package rewrite

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"sort"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// 替换位置
type PosAlt struct {
	Pos, End  token.Position
	Alternate []byte
}

type CodeSource struct {
	Fileast    *ast.File
	Fileset    *token.FileSet
	Source     []byte
	Alternates []PosAlt // Source 只能替换一次，然后必须重新解析 Fileast
	*printer.Config
}

func NewCodeSource() *CodeSource {
	return &CodeSource{
		Fileset: token.NewFileSet(),
		Config: &printer.Config{
			Mode:     printer.TabIndent,
			Tabwidth: 4,
		},
	}
}

func (cs *CodeSource) SetSource(source []byte) (err error) {
	cs.Source = source
	cs.Fileast, err = parser.ParseFile(cs.Fileset, "", source, parser.ParseComments)
	return
}

func (cs *CodeSource) GetContent() ([]byte, error) {
	var buf bytes.Buffer
	err := cs.Config.Fprint(&buf, cs.Fileset, cs.Fileast)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (cs *CodeSource) AddCode(code []byte) error {
	content, err := cs.GetContent()
	if err != nil {
		return err
	}
	return cs.SetSource(append(content, code...))
}

func (cs *CodeSource) AddStringCode(code string) error {
	return cs.AddCode([]byte(code))
}

func (cs *CodeSource) GetPackage() string {
	if cs.Fileast != nil {
		return cs.Fileast.Name.Name
	}
	return ""
}

func (cs *CodeSource) GetPackageOffset() int {
	if cs.Fileast != nil {
		pos := cs.Fileast.Name.End()
		return cs.Fileset.PositionFor(pos, false).Offset
	}
	return 0
}

func (cs *CodeSource) SetPackage(name string) (err error) {
	if cs.Fileast == nil {
		code := fmt.Sprintf("package %s", name)
		err = cs.SetSource([]byte(code))
	} else {
		cs.Fileast.Name.Name = name
	}
	return
}

func (cs *CodeSource) AddImport(path, alias string) bool {
	return astutil.AddNamedImport(cs.Fileset, cs.Fileast, alias, path)
}

func (cs *CodeSource) DelImport(path, alias string) bool {
	if astutil.UsesImport(cs.Fileast, path) {
		return false
	}
	return astutil.DeleteNamedImport(cs.Fileset, cs.Fileast, alias, path)
}

func (cs *CodeSource) CleanImports() (removes int) {
	for _, groups := range astutil.Imports(cs.Fileset, cs.Fileast) {
		for _, imp := range groups {
			var path, alias string
			if imp.Name != nil {
				alias = imp.Name.Name
			}
			path = strings.Trim(imp.Path.Value, "\"")
			if cs.DelImport(path, alias) {
				removes++
			}
		}
	}
	return
}

func (cs *CodeSource) GetNodeCode(node ast.Node) string {
	// 请先保证 node 不是 nil
	pos := cs.Fileset.PositionFor(node.Pos(), false)
	end := cs.Fileset.PositionFor(node.End(), false)
	return string(cs.Source[pos.Offset:end.Offset])
}

func (cs *CodeSource) GetFieldCode(node *DeclNode, i int) string {
	if i < 0 {
		i += len(node.Fields)
	}
	ffcode := cs.GetNodeCode(node.Fields[i])
	return strings.TrimSpace(ffcode)
}

func (cs *CodeSource) GetComment(c *ast.CommentGroup, trim bool) string {
	if c == nil {
		return ""
	}
	comment := cs.GetNodeCode(c)
	if trim {
		comment = TrimComment(comment)
	}
	return comment
}

func (cs *CodeSource) AddReplace(first, last ast.Node, code string) {
	// 请先保证 first, last 不是 nil
	pos := cs.Fileset.PositionFor(first.Pos(), false)
	end := cs.Fileset.PositionFor(last.End(), false)
	alt := PosAlt{Pos: pos, End: end, Alternate: []byte(code)}
	cs.Alternates = append(cs.Alternates, alt)
}

func (cs *CodeSource) AltSource() ([]byte, bool) {
	if len(cs.Alternates) == 0 {
		return cs.Source, false
	}
	sort.Slice(cs.Alternates, func(i, j int) bool {
		return cs.Alternates[i].Pos.Offset < cs.Alternates[j].Pos.Offset
	})
	var chunks [][]byte
	start, stop := 0, 0
	for _, alt := range cs.Alternates {
		start = alt.Pos.Offset
		chunks = append(chunks, cs.Source[stop:start])
		chunks = append(chunks, alt.Alternate)
		stop = alt.End.Offset
	}
	if stop < len(cs.Source) {
		chunks = append(chunks, cs.Source[stop:])
	}
	cs.Alternates = make([]PosAlt, 0)
	return bytes.Join(chunks, nil), true
}

func (cs *CodeSource) WriteTo(filename string) error {
	code, err := cs.GetContent()
	if err != nil {
		return err
	}
	_, err = WriteGolangFile(filename, code)
	return err
}

func WithImports(pkg string, source []byte, imports map[string]string) (*CodeSource, error) {
	cs := NewCodeSource()
	if err := cs.SetPackage(pkg); err != nil {
		return cs, err
	}
	// 添加可能引用的包，后面再尝试删除不一定会用的包
	for imp, alias := range imports {
		cs.AddImport(imp, alias)
	}
	if err := cs.AddCode(source); err != nil {
		return cs, err
	}
	for imp, alias := range imports {
		cs.DelImport(imp, alias)
	}
	return cs, nil
}

func ResetImports(cs *CodeSource, imports map[string]string) (*CodeSource, error) {
	var err error
	pkg, offset := cs.GetPackage(), cs.GetPackageOffset()
	source, _err := FormatGolangCode(cs.Source[offset:])
	cs, err = WithImports(pkg, source, imports)
	if err == nil {
		err = _err
	}
	return cs, err
}
