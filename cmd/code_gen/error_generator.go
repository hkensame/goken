package main // import "golang.org/x/tools/cmd/stringer"

// import (
// 	"bytes"
// 	"errors"
// 	"flag"
// 	"fmt"
// 	"go/ast"
// 	"go/constant"
// 	"go/parser"
// 	"go/token"
// 	"go/types"
// 	"log"
// 	"os"
// 	"regexp"
// 	"sort"
// 	"strings"

// 	"golang.org/x/tools/go/packages"
// )

// var (
// 	outPath   = flag.String("error_out", "./", "生成的文件位置")
// 	fileName  = flag.String("filename", "error.go", "生成的文件名称")
// 	format    = flag.String("format", "{msg}:{http}:{rpc}", "注释中格式化http_code和rpc_code的形式,语句内http代表http_code,rpc代表rpc_code")
// 	goPackage = flag.String("go_package", "", "生成文件的package名称,默认使用传入文件的package")
// 	inPath    = flag.String("path", "", "需要传入的error.go模板文件")
// )

// func Usage() {
// 	fmt.Println("ErrorGen发生错误,请遵循以下守则")
// 	s := `
// 	ErrorGen具有以下几个flag:
// 		1. -error_out  可选的指定生成的文件位置 默认为当前路径./
// 		2. -filename   可选的生成的文件名称 默认名称为error.go
// 		3. -format     可选的用于解析注释中格式化http_code和rpc_code的形式,语句内msg代表错误信息,http代表http_code,rpc代表rpc_code 默认为{msg}:{http}:{rpc}
// 		4. -go_package 可选的生成文件的package名称 默认使用传入文件的package
// 		5. -path	   必须的传入的error.go模板文件的地址
// 	`
// 	fmt.Println(s)
// }

// func parseFormat(str string) []string {
// 	//前三个位置分别存储 三个固定属性(msg,http_code,rpc_code)
// 	//后面的属性存储定义的format
// 	placeholders := make([]string, 3)
// 	paragraphs := make([]string, 0)

// 	//用正则表达式找到{}内的所有值
// 	re := regexp.MustCompile(`\{([^}]+)\}`)
// 	matches := re.FindAllStringSubmatchIndex(str, -1)
// 	lastIndex := 0

// 	for _, match := range matches {
// 		// 提取占位符内的内容,以此得到三个属性的出现顺序
// 		placeholders = append(placeholders, str[match[2]:match[3]])

// 		// 提取占位符之前的段落格式,如果前面不存在前缀则不会添加
// 		if match[0] > lastIndex {
// 			//这里可能要对字符串再修理一下,比如去除多余的空格
// 			tmstr := str[lastIndex:match[0]]
// 			paragraphs = append(paragraphs, tmstr)
// 		}
// 		// 更新最后索引位置
// 		lastIndex = match[1]
// 	}

// 	// 添加最后一个有值的段落(边界条件)
// 	if lastIndex < len(str) {
// 		paragraphs = append(paragraphs, str[lastIndex:])
// 	}
// 	return append(placeholders, paragraphs...)
// }

// func main() {
// 	flag.Usage = Usage
// 	flag.Parse()
// 	if *inPath == "" {
// 		flag.Usage()
// 		os.Exit(2)
// 	}
// 	mtdata := parseFormat(*format)
// 	file, err := parser.ParseFile(token.NewFileSet(), *inPath, nil, parser.ParseComments)
// 	if err != nil {
// 		flag.Usage()
// 		panic(err)
// 	}
// 	*goPackage = file.Name.Name
// 	genDecl(file.Decls, mtdata)

// }

// func genDecl(decls []ast.Decl, mtdata []string) {
// 	for _, decl := range decls {
// 		genDecl, ok := decl.(*ast.GenDecl)
// 		if !ok {
// 			flag.Usage()
// 			panic(errors.New("序列化代码失败"))
// 		}

// 		for _, spec := range genDecl.Specs {
// 			switch tv := spec.(type) {
// 			case *ast.TypeSpec:

// 			case *ast.ValueSpec:
// 				if tv.Doc != nil && tv.Doc.Text() != "" {
// 					//按format格式解析注解
// 				} else if tv.Comment != nil && tv.Comment.Text() != "" {

// 				}
// 			}
// 		}
// 	}
// }

// // baseName that will put the generated code together with pkg.
// func baseName(pkg *Package, typename string) string {
// 	suffix := "string.go"
// 	if pkg.hasTestFiles {
// 		suffix = "string_test.go"
// 	}
// 	return fmt.Sprintf("%s_%s", strings.ToLower(typename), suffix)
// }

// // isDirectory reports whether the named file is a directory.
// func isDirectory(name string) bool {
// 	info, err := os.Stat(name)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return info.IsDir()
// }

// // Generator holds the state of the analysis. Primarily used to buffer
// // the output for format.Source.
// type Generator struct {
// 	buf bytes.Buffer // Accumulated output.
// 	pkg *Package     // Package we are scanning.

// 	logf func(format string, args ...any) // test logging hook; nil when not testing
// }

// func (g *Generator) Printf(format string, args ...any) {
// 	fmt.Fprintf(&g.buf, format, args...)
// }

// // File holds a single parsed file and associated data.
// type File struct {
// 	pkg  *Package  // Package to which this file belongs.
// 	file *ast.File // Parsed AST.
// 	// These fields are reset for each type being generated.
// 	typeName string  // Name of the constant type.
// 	values   []Value // Accumulator for constant values of that type.

// 	trimPrefix  string
// 	lineComment bool
// }

// type Package struct {
// 	name         string
// 	defs         map[*ast.Ident]types.Object
// 	files        []*File
// 	hasTestFiles bool
// }

// // loadPackages analyzes the single package constructed from the patterns and tags.
// // loadPackages exits if there is an error.
// //
// // Returns all variants (such as tests) of the package.
// //
// // logf is a test logging hook. It can be nil when not testing.
// func loadPackages(
// 	patterns, tags []string,
// 	trimPrefix string, lineComment bool,
// 	logf func(format string, args ...any),
// ) []*Package {
// 	cfg := &packages.Config{
// 		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedFiles,
// 		// Tests are included, let the caller decide how to fold them in.
// 		Tests:      true,
// 		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
// 		Logf:       logf,
// 	}
// 	pkgs, err := packages.Load(cfg, patterns...)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if len(pkgs) == 0 {
// 		log.Fatalf("error: no packages matching %v", strings.Join(patterns, " "))
// 	}

// 	out := make([]*Package, len(pkgs))
// 	for i, pkg := range pkgs {
// 		p := &Package{
// 			name:  pkg.Name,
// 			defs:  pkg.TypesInfo.Defs,
// 			files: make([]*File, len(pkg.Syntax)),
// 		}

// 		for j, file := range pkg.Syntax {
// 			p.files[j] = &File{
// 				file: file,
// 				pkg:  p,

// 				trimPrefix:  trimPrefix,
// 				lineComment: lineComment,
// 			}
// 		}

// 		// Keep track of test files, since we might want to generated
// 		// code that ends up in that kind of package.
// 		// Can be replaced once https://go.dev/issue/38445 lands.
// 		for _, f := range pkg.GoFiles {
// 			if strings.HasSuffix(f, "_test.go") {
// 				p.hasTestFiles = true
// 				break
// 			}
// 		}

// 		out[i] = p
// 	}
// 	return out
// }

// func findValues(typeName string, pkg *Package) []Value {
// 	values := make([]Value, 0, 100)
// 	for _, file := range pkg.files {
// 		// Set the state for this run of the walker.
// 		file.typeName = typeName
// 		file.values = nil
// 		if file.file != nil {
// 			ast.Inspect(file.file, file.genDecl)
// 			values = append(values, file.values...)
// 		}
// 	}
// 	return values
// }

// // generate produces the String method for the named type.
// func (g *Generator) generate(typeName string, values []Value) {
// 	// Generate code that will fail if the constants change value.
// 	g.Printf("func _() {\n")
// 	g.Printf("\t// An \"invalid array index\" compiler error signifies that the constant values have changed.\n")
// 	g.Printf("\t// Re-run the stringer command to generate them again.\n")
// 	g.Printf("\tvar x [1]struct{}\n")
// 	for _, v := range values {
// 		g.Printf("\t_ = x[%s - %s]\n", v.originalName, v.str)
// 	}
// 	g.Printf("}\n")
// 	runs := splitIntoRuns(values)
// 	// The decision of which pattern to use depends on the number of
// 	// runs in the numbers. If there's only one, it's easy. For more than
// 	// one, there's a tradeoff between complexity and size of the data
// 	// and code vs. the simplicity of a map. A map takes more space,
// 	// but so does the code. The decision here (crossover at 10) is
// 	// arbitrary, but considers that for large numbers of runs the cost
// 	// of the linear scan in the switch might become important, and
// 	// rather than use yet another algorithm such as binary search,
// 	// we punt and use a map. In any case, the likelihood of a map
// 	// being necessary for any realistic example other than bitmasks
// 	// is very low. And bitmasks probably deserve their own analysis,
// 	// to be done some other day.
// 	switch {
// 	case len(runs) == 1:
// 		g.buildOneRun(runs, typeName)
// 	case len(runs) <= 10:
// 		g.buildMultipleRuns(runs, typeName)
// 	default:
// 		g.buildMap(runs, typeName)
// 	}
// }

// // splitIntoRuns breaks the values into runs of contiguous sequences.
// // For example, given 1,2,3,5,6,7 it returns {1,2,3},{5,6,7}.
// // The input slice is known to be non-empty.
// func splitIntoRuns(values []Value) [][]Value {
// 	// We use stable sort so the lexically first name is chosen for equal elements.
// 	sort.Stable(byValue(values))
// 	// Remove duplicates. Stable sort has put the one we want to print first,
// 	// so use that one. The String method won't care about which named constant
// 	// was the argument, so the first name for the given value is the only one to keep.
// 	// We need to do this because identical values would cause the switch or map
// 	// to fail to compile.
// 	j := 1
// 	for i := 1; i < len(values); i++ {
// 		if values[i].value != values[i-1].value {
// 			values[j] = values[i]
// 			j++
// 		}
// 	}
// 	values = values[:j]
// 	runs := make([][]Value, 0, 10)
// 	for len(values) > 0 {
// 		// One contiguous sequence per outer loop.
// 		i := 1
// 		for i < len(values) && values[i].value == values[i-1].value+1 {
// 			i++
// 		}
// 		runs = append(runs, values[:i])
// 		values = values[i:]
// 	}
// 	return runs
// }

// // format returns the gofmt-ed contents of the Generator's buffer.
// func (g *Generator) format() []byte {
// 	src, err := format.Source(g.buf.Bytes())
// 	if err != nil {
// 		// Should never happen, but can arise when developing this code.
// 		// The user can compile the output to see the error.
// 		log.Printf("warning: internal error: invalid Go generated: %s", err)
// 		log.Printf("warning: compile the package to analyze the error")
// 		return g.buf.Bytes()
// 	}
// 	return src
// }

// // Value represents a declared constant.
// type Value struct {
// 	originalName string // The name of the constant.
// 	name         string // The name with trimmed prefix.
// 	// The value is stored as a bit pattern alone. The boolean tells us
// 	// whether to interpret it as an int64 or a uint64; the only place
// 	// this matters is when sorting.
// 	// Much of the time the str field is all we need; it is printed
// 	// by Value.String.
// 	value  uint64 // Will be converted to int64 when needed.
// 	signed bool   // Whether the constant is a signed type.
// 	str    string // The string representation given by the "go/constant" package.
// }

// func (v *Value) String() string {
// 	return v.str
// }

// // byValue lets us sort the constants into increasing order.
// // We take care in the Less method to sort in signed or unsigned order,
// // as appropriate.
// type byValue []Value

// func (b byValue) Len() int      { return len(b) }
// func (b byValue) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
// func (b byValue) Less(i, j int) bool {
// 	if b[i].signed {
// 		return int64(b[i].value) < int64(b[j].value)
// 	}
// 	return b[i].value < b[j].value
// }

// // genDecl processes one declaration clause.
// func (f *File) genDecl(node ast.Node) bool {
// 	decl, ok := node.(*ast.GenDecl)
// 	if !ok || decl.Tok != token.CONST {
// 		// We only care about const declarations.
// 		return true
// 	}
// 	// The name of the type of the constants we are declaring.
// 	// Can change if this is a multi-element declaration.
// 	typ := ""
// 	// Loop over the elements of the declaration. Each element is a ValueSpec:
// 	// a list of names possibly followed by a type, possibly followed by values.
// 	// If the type and value are both missing, we carry down the type (and value,
// 	// but the "go/types" package takes care of that).
// 	for _, spec := range decl.Specs {
// 		vspec := spec.(*ast.ValueSpec) // Guaranteed to succeed as this is CONST.
// 		if vspec.Type == nil && len(vspec.Values) > 0 {
// 			// "X = 1". With no type but a value. If the constant is untyped,
// 			// skip this vspec and reset the remembered type.
// 			typ = ""

// 			// If this is a simple type conversion, remember the type.
// 			// We don't mind if this is actually a call; a qualified call won't
// 			// be matched (that will be SelectorExpr, not Ident), and only unusual
// 			// situations will result in a function call that appears to be
// 			// a type conversion.
// 			ce, ok := vspec.Values[0].(*ast.CallExpr)
// 			if !ok {
// 				continue
// 			}
// 			id, ok := ce.Fun.(*ast.Ident)
// 			if !ok {
// 				continue
// 			}
// 			typ = id.Name
// 		}
// 		if vspec.Type != nil {
// 			// "X T". We have a type. Remember it.
// 			ident, ok := vspec.Type.(*ast.Ident)
// 			if !ok {
// 				continue
// 			}
// 			typ = ident.Name
// 		}
// 		if typ != f.typeName {
// 			// This is not the type we're looking for.
// 			continue
// 		}
// 		// We now have a list of names (from one line of source code) all being
// 		// declared with the desired type.
// 		// Grab their names and actual values and store them in f.values.
// 		for _, name := range vspec.Names {
// 			if name.Name == "_" {
// 				continue
// 			}
// 			// This dance lets the type checker find the values for us. It's a
// 			// bit tricky: look up the object declared by the name, find its
// 			// types.Const, and extract its value.
// 			obj, ok := f.pkg.defs[name]
// 			if !ok {
// 				log.Fatalf("no value for constant %s", name)
// 			}
// 			info := obj.Type().Underlying().(*types.Basic).Info()
// 			if info&types.IsInteger == 0 {
// 				log.Fatalf("can't handle non-integer constant type %s", typ)
// 			}
// 			value := obj.(*types.Const).Val() // Guaranteed to succeed as this is CONST.
// 			if value.Kind() != constant.Int {
// 				log.Fatalf("can't happen: constant is not an integer %s", name)
// 			}
// 			i64, isInt := constant.Int64Val(value)
// 			u64, isUint := constant.Uint64Val(value)
// 			if !isInt && !isUint {
// 				log.Fatalf("internal error: value of %s is not an integer: %s", name, value.String())
// 			}
// 			if !isInt {
// 				u64 = uint64(i64)
// 			}
// 			v := Value{
// 				originalName: name.Name,
// 				value:        u64,
// 				signed:       info&types.IsUnsigned == 0,
// 				str:          value.String(),
// 			}
// 			if c := vspec.Comment; f.lineComment && c != nil && len(c.List) == 1 {
// 				v.name = strings.TrimSpace(c.Text())
// 			} else {
// 				v.name = strings.TrimPrefix(v.originalName, f.trimPrefix)
// 			}
// 			f.values = append(f.values, v)
// 		}
// 	}
// 	return false
// }

// // Helpers

// // usize returns the number of bits of the smallest unsigned integer
// // type that will hold n. Used to create the smallest possible slice of
// // integers to use as indexes into the concatenated strings.
// func usize(n int) int {
// 	switch {
// 	case n < 1<<8:
// 		return 8
// 	case n < 1<<16:
// 		return 16
// 	default:
// 		// 2^32 is enough constants for anyone.
// 		return 32
// 	}
// }

// // declareIndexAndNameVars declares the index slices and concatenated names
// // strings representing the runs of values.
// func (g *Generator) declareIndexAndNameVars(runs [][]Value, typeName string) {
// 	var indexes, names []string
// 	for i, run := range runs {
// 		index, name := g.createIndexAndNameDecl(run, typeName, fmt.Sprintf("_%d", i))
// 		if len(run) != 1 {
// 			indexes = append(indexes, index)
// 		}
// 		names = append(names, name)
// 	}
// 	g.Printf("const (\n")
// 	for _, name := range names {
// 		g.Printf("\t%s\n", name)
// 	}
// 	g.Printf(")\n\n")

// 	if len(indexes) > 0 {
// 		g.Printf("var (")
// 		for _, index := range indexes {
// 			g.Printf("\t%s\n", index)
// 		}
// 		g.Printf(")\n\n")
// 	}
// }

// // declareIndexAndNameVar is the single-run version of declareIndexAndNameVars
// func (g *Generator) declareIndexAndNameVar(run []Value, typeName string) {
// 	index, name := g.createIndexAndNameDecl(run, typeName, "")
// 	g.Printf("const %s\n", name)
// 	g.Printf("var %s\n", index)
// }

// // createIndexAndNameDecl returns the pair of declarations for the run. The caller will add "const" and "var".
// func (g *Generator) createIndexAndNameDecl(run []Value, typeName string, suffix string) (string, string) {
// 	b := new(bytes.Buffer)
// 	indexes := make([]int, len(run))
// 	for i := range run {
// 		b.WriteString(run[i].name)
// 		indexes[i] = b.Len()
// 	}
// 	nameConst := fmt.Sprintf("_%s_name%s = %q", typeName, suffix, b.String())
// 	nameLen := b.Len()
// 	b.Reset()
// 	fmt.Fprintf(b, "_%s_index%s = [...]uint%d{0, ", typeName, suffix, usize(nameLen))
// 	for i, v := range indexes {
// 		if i > 0 {
// 			fmt.Fprintf(b, ", ")
// 		}
// 		fmt.Fprintf(b, "%d", v)
// 	}
// 	fmt.Fprintf(b, "}")
// 	return b.String(), nameConst
// }

// // declareNameVars declares the concatenated names string representing all the values in the runs.
// func (g *Generator) declareNameVars(runs [][]Value, typeName string, suffix string) {
// 	g.Printf("const _%s_name%s = \"", typeName, suffix)
// 	for _, run := range runs {
// 		for i := range run {
// 			g.Printf("%s", run[i].name)
// 		}
// 	}
// 	g.Printf("\"\n")
// }

// // buildOneRun generates the variables and String method for a single run of contiguous values.
// func (g *Generator) buildOneRun(runs [][]Value, typeName string) {
// 	values := runs[0]
// 	g.Printf("\n")
// 	g.declareIndexAndNameVar(values, typeName)
// 	// The generated code is simple enough to write as a Printf format.
// 	lessThanZero := ""
// 	if values[0].signed {
// 		lessThanZero = "i < 0 || "
// 	}
// 	if values[0].value == 0 { // Signed or unsigned, 0 is still 0.
// 		g.Printf(stringOneRun, typeName, usize(len(values)), lessThanZero)
// 	} else {
// 		g.Printf(stringOneRunWithOffset, typeName, values[0].String(), usize(len(values)), lessThanZero)
// 	}
// }

// // Arguments to format are:
// //
// //	[1]: type name
// //	[2]: size of index element (8 for uint8 etc.)
// //	[3]: less than zero check (for signed types)
// const stringOneRun = `func (i %[1]s) String() string {
// 	if %[3]si >= %[1]s(len(_%[1]s_index)-1) {
// 		return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
// 	}
// 	return _%[1]s_name[_%[1]s_index[i]:_%[1]s_index[i+1]]
// }
// `

// // Arguments to format are:
// //	[1]: type name
// //	[2]: lowest defined value for type, as a string
// //	[3]: size of index element (8 for uint8 etc.)
// //	[4]: less than zero check (for signed types)
// /*
//  */
// const stringOneRunWithOffset = `func (i %[1]s) String() string {
// 	i -= %[2]s
// 	if %[4]si >= %[1]s(len(_%[1]s_index)-1) {
// 		return "%[1]s(" + strconv.FormatInt(int64(i + %[2]s), 10) + ")"
// 	}
// 	return _%[1]s_name[_%[1]s_index[i] : _%[1]s_index[i+1]]
// }
// `

// // buildMultipleRuns generates the variables and String method for multiple runs of contiguous values.
// // For this pattern, a single Printf format won't do.
// func (g *Generator) buildMultipleRuns(runs [][]Value, typeName string) {
// 	g.Printf("\n")
// 	g.declareIndexAndNameVars(runs, typeName)
// 	g.Printf("func (i %s) String() string {\n", typeName)
// 	g.Printf("\tswitch {\n")
// 	for i, values := range runs {
// 		if len(values) == 1 {
// 			g.Printf("\tcase i == %s:\n", &values[0])
// 			g.Printf("\t\treturn _%s_name_%d\n", typeName, i)
// 			continue
// 		}
// 		if values[0].value == 0 && !values[0].signed {
// 			// For an unsigned lower bound of 0, "0 <= i" would be redundant.
// 			g.Printf("\tcase i <= %s:\n", &values[len(values)-1])
// 		} else {
// 			g.Printf("\tcase %s <= i && i <= %s:\n", &values[0], &values[len(values)-1])
// 		}
// 		if values[0].value != 0 {
// 			g.Printf("\t\ti -= %s\n", &values[0])
// 		}
// 		g.Printf("\t\treturn _%s_name_%d[_%s_index_%d[i]:_%s_index_%d[i+1]]\n",
// 			typeName, i, typeName, i, typeName, i)
// 	}
// 	g.Printf("\tdefault:\n")
// 	g.Printf("\t\treturn \"%s(\" + strconv.FormatInt(int64(i), 10) + \")\"\n", typeName)
// 	g.Printf("\t}\n")
// 	g.Printf("}\n")
// }

// // buildMap handles the case where the space is so sparse a map is a reasonable fallback.
// // It's a rare situation but has simple code.
// func (g *Generator) buildMap(runs [][]Value, typeName string) {
// 	g.Printf("\n")
// 	g.declareNameVars(runs, typeName, "")
// 	g.Printf("\nvar _%s_map = map[%s]string{\n", typeName, typeName)
// 	n := 0
// 	for _, values := range runs {
// 		for _, value := range values {
// 			g.Printf("\t%s: _%s_name[%d:%d],\n", &value, typeName, n, n+len(value.name))
// 			n += len(value.name)
// 		}
// 	}
// 	g.Printf("}\n\n")
// 	g.Printf(stringMap, typeName)
// }

// // Argument to format is the type name.
// const stringMap = `func (i %[1]s) String() string {
// 	if str, ok := _%[1]s_map[i]; ok {
// 		return str
// 	}
// 	return "%[1]s(" + strconv.FormatInt(int64(i), 10) + ")"
// }
// `
