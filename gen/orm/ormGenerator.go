package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

type replacement struct {
	StructName string
	TableName  string
	Fields     []Field
	Index
}

type Field struct {
	Name string
	Type string
}

type Index struct {
	Field
	Unique bool
}

type structInfo struct {
	rep     replacement
	Indexes []Index
}

type generatingInfo struct {
	needCodegen bool
	needSwagger bool
	packageName string
	structInfos []structInfo
}

var (
	getByTpl = template.Must(template.New("getByTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (p *{{.StructName}}) GetBy{{.Index.Field.Name}}(app application.App, {{.Index.Field.Name}} {{.Index.Field.Type}}) error {
	sqlStatement := "SELECT * FROM {{withQuotes .TableName}} WHERE {{toSnakeCase .Index.Field.Name}} = ?"
	err := app.DB.QueryRow(sqlStatement, {{.Index.Field.Name}}).Scan({{range $Index, $element := .Fields}}{{if ne $Index 0}}, {{end}}&p.{{$element.Name}}{{end}})
	return err
}
`))
	deleteTpl = template.Must(template.New("deleteTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (p *{{.StructName}}) Delete(app application.App) error {
	_, err := app.DB.Exec("DELETE FROM {{withQuotes .TableName}} where ID = ?", p.ID)
	return err
}
`))
	createTpl = template.Must(template.New("createTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (p *{{.StructName}}) Create(app application.App) error {
	result, err := app.DB.Exec("INSERT INTO {{withQuotes .TableName}} ({{range $Index, $element := .Fields}}{{if ne $Index 0}}{{if ne $Index 1}}, {{end}}{{toSnakeCase $element.Name}}{{end}}{{end}}) VALUES ({{range $Index, $element := .Fields}}{{if ne $Index 0}}{{if ne $Index 1}}, {{end}}?{{end}}{{end}})"{{range $Index, $element := .Fields}}{{if ne $Index 0}}, p.{{$element.Name}}{{end}}{{end}})
	id, _ := result.LastInsertId()
	p.ID = uint(id)	
	return err
}
`))
	rowsTpl = template.Must(template.New("rowsTpl").Parse(
		`type {{.StructName}}s []{{.StructName}}
`))
	getAllTpl = template.Must(template.New("getAllTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (ps *{{.StructName}}s) GetAll(app application.App) error{
	*ps = (*ps)[:0]
	sqlStatement := "SELECT * FROM {{withQuotes .TableName}}"
	rows, err := app.DB.Query(sqlStatement)
	if err != nil{
		return err
	}
	defer rows.Close()

	for rows.Next() {
		p := {{.StructName}}{}
		err := rows.Scan({{range $Index, $element := .Fields}}{{if ne $Index 0}}, {{end}}&p.{{$element.Name}}{{end}})
		if err != nil {
			return err
		}
		*ps = append(*ps, p)
	}
	return nil
}
`))
	getAllByTpl = template.Must(template.New("getAllByTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (ps *{{.StructName}}s) GetAllBy{{.Index.Field.Name}}(app application.App, {{.Index.Field.Name}} {{.Index.Field.Type}}) error{
	*ps = (*ps)[:0]
	sqlStatement := "SELECT * FROM {{withQuotes .TableName}} WHERE {{toSnakeCase .Index.Field.Name}} = ?"
	rows, err := app.DB.Query(sqlStatement, {{.Index.Field.Name}})
	if err != nil{
		return err
	}
	defer rows.Close()

	for rows.Next() {
		p := {{.StructName}}{}
		err := rows.Scan({{range $Index, $element := .Fields}}{{if ne $Index 0}}, {{end}}&p.{{$element.Name}}{{end}})
		if err != nil {
			return err
		}
		*ps = append(*ps, p)
	}
	return nil
}
`))
	uniqueMapByTpl = template.Must(template.New("uniqueMapByTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (ps *{{.StructName}}s) MapBy{{.Index.Field.Name}}() map[{{.Index.Field.Type}}]{{.StructName}} {
	{{.StructName}}sMap := map[{{.Index.Field.Type}}]{{.StructName}}{}
	for i, {{.StructName}} := range *ps{
		{{.StructName}}sMap[{{.StructName}}.{{.Index.Field.Name}}] = (*ps)[i]
	}
	return {{.StructName}}sMap
}
`))
	mapByTpl = template.Must(template.New("mapByTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (ps *{{.StructName}}s) MapBy{{.Index.Field.Name}}() map[{{.Index.Field.Type}}]{{.StructName}}s {
	{{.StructName}}sMap := map[{{.Index.Field.Type}}]{{.StructName}}s{}
	for i, {{.StructName}} := range *ps{
		{{.StructName}}sMap[{{.StructName}}.{{.Index.Field.Name}}] = append({{.StructName}}sMap[{{.StructName}}.{{.Index.Field.Name}}], (*ps)[i])
	}
	return {{.StructName}}sMap
}
`))
	deleteAllTpl = template.Must(template.New("deleteAllTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (ps *{{.StructName}}s) Delete(app application.App) error {
	sqlStatement := "DELETE FROM {{withQuotes .TableName}} where "
	var idsToDelete []string
	for _, {{.StructName}} := range *ps{
		idsToDelete = append(idsToDelete, fmt.Sprint("id = ", {{.StructName}}.ID))
	}
	sqlStatement += strings.Join(idsToDelete, " OR ")
	_, err := app.DB.Exec(sqlStatement)
	return err
}
`))
	updateTpl = template.Must(template.New("updateTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (p *{{.StructName}}) Update(app application.App) error {
	sqlStatement := "UPDATE {{withQuotes .TableName}} SET {{range $Index, $element := .Fields}}{{if ne $Index 0}}{{if ne $Index 1}}, {{end}}{{toSnakeCase $element.Name}} = ?{{end}}{{end}} WHERE id = ?"
	_, err := app.DB.Exec(sqlStatement{{range $Index, $element := .Fields}}{{if ne $Index 0}}, p.{{$element.Name}}{{end}}{{end}}, p.ID)
	return err
}
`))
	createAllTpl = template.Must(template.New("createAllTpl").Funcs(template.FuncMap{
		"toSnakeCase": toSnakeCase,
		"withQuotes":  withQuotes,
	}).Parse(
		`func (ps *{{.StructName}}s) Create(app application.App) error {
	if len(*ps) == 0 {
		return nil
	}
	sqlStatement := "INSERT INTO {{withQuotes .TableName}} ({{range $Index, $element := .Fields}}{{if ne $Index 0}}{{if ne $Index 1}}, {{end}}{{toSnakeCase $element.Name}}{{end}}{{end}}) values "
	var valuesBlocks []string
	for _, p:= range *ps{
		values := []string{
			{{range $Index, $element := .Fields}}{{if ne $Index 0}}{{if ne $element.Type "string"}}fmt.Sprint({{else}}"\"" + {{end}}p.{{$element.Name}}{{if ne $element.Type "string"}}){{else}} + "\""{{end}}, {{end}}{{end}}
		}

		valuesBlocks = append(valuesBlocks, "(" + strings.Join(values, ",") + ")")
	}
	sqlStatement += strings.Join(valuesBlocks, ",")
	_, err := app.DB.Exec(sqlStatement)
	return err
}
`))
)

func main() {
	files, err := ioutil.ReadDir("src/models/dbaseModels")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		genInfo := getGeneratingInfoFrom(file)
		if genInfo.needCodegen {
			generateForFile(file, genInfo)
		}
	}
}

func getGeneratingInfoFrom(file os.FileInfo) generatingInfo {
	var genInfo generatingInfo
	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, "src/models/dbaseModels/"+file.Name(), nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	genInfo.packageName = node.Name.Name
	genInfo.needCodegen = false

	for _, f := range node.Decls {
		rep := replacement{}
		var Indexes []Index
		g, ok := f.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %#T is not *ast.GenDecl\n", f)
			continue
		}
	SPECS_LOOP:
		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				fmt.Printf("SKIP %#T is not ast.FieldTypeSpec\n", spec)
				continue
			}
			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				fmt.Printf("SKIP %#T is not ast.StructType\n", currStruct)
				continue
			}
			if g.Doc == nil {
				fmt.Printf("SKIP struct %#v doesn't have comments\n",
					currType.Name.Name)
				continue
			}
			for _, comment := range g.Doc.List {
				genInfo.needCodegen = genInfo.needCodegen || strings.HasPrefix(comment.Text,
					"// gen")
				genInfo.needSwagger = strings.Contains(comment.Text, "swag")
			}
			if !genInfo.needCodegen {
				fmt.Printf("SKIP struct %#v doesn't have gen mark\n",
					currType.Name.Name)
				continue SPECS_LOOP
			}
			rep.StructName = currType.Name.Name
			rep.TableName = toSnakeCase(currType.Name.Name) + "s"
			for _, field := range currStruct.Fields.List {
				Field := Field{
					field.Names[0].Name,
					field.Type.(*ast.Ident).Name,
				}
				rep.Fields = append(rep.Fields, Field)
				if field.Tag != nil {
					tag := reflect.StructTag(
						field.Tag.Value[1 : len(field.Tag.Value)-1])
					switch tag.Get("gen") {
					case "-":
						continue
					case "uindex":
						Indexes = append(Indexes, Index{Field, true})
					case "index":
						Indexes = append(Indexes, Index{Field, false})
					}
				}
			}
			genInfo.structInfos = append(genInfo.structInfos, structInfo{
				rep:     rep,
				Indexes: Indexes,
			})
		}
	}

	return genInfo
}

func generateForFile(file os.FileInfo, genInfo generatingInfo) {
	out, _ := os.Create("src/models/dbaseModels/" + strings.Replace(file.Name(), ".", "_standard_implementation.", -1))

	fmt.Fprintln(out, `//Файл сгенерирован автоматически по `+file.Name())
	fmt.Fprintln(out, `//ищи генератор в ./gen/orm/ormGenerator.go`)
	fmt.Fprintln(out, `package `+genInfo.packageName)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `import "fmt"`)
	fmt.Fprintln(out, `import "strings"`)
	fmt.Fprintln(out, `import "urms/application"`)
	fmt.Fprintln(out)

	for _, structInfo := range genInfo.structInfos {
		rowsTpl.Execute(out, structInfo.rep)
		fmt.Fprintln(out)
		for _, ind := range structInfo.Indexes {
			structInfo.rep.Index.Field.Name = ind.Field.Name
			structInfo.rep.Index.Field.Type = ind.Field.Type
			if ind.Unique {
				getByTpl.Execute(out, structInfo.rep)
				fmt.Fprintln(out)
				uniqueMapByTpl.Execute(out, structInfo.rep)
			} else {
				getAllByTpl.Execute(out, structInfo.rep)
				fmt.Fprintln(out)
				mapByTpl.Execute(out, structInfo.rep)
			}
			fmt.Fprintln(out)
		}
		deleteTpl.Execute(out, structInfo.rep)
		fmt.Fprintln(out)
		deleteAllTpl.Execute(out, structInfo.rep)
		fmt.Fprintln(out)
		createAllTpl.Execute(out, structInfo.rep)
		fmt.Fprintln(out)
		updateTpl.Execute(out, structInfo.rep)
		fmt.Fprintln(out)
		createTpl.Execute(out, structInfo.rep)
		fmt.Fprintln(out)
		if genInfo.needSwagger {
			fmt.Fprintln(out, "//swagger:response", structInfo.rep.StructName+"s")
		}
		getAllTpl.Execute(out, structInfo.rep)
	}
}

func toSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func withQuotes(str string) string {
	return "`" + str + "`"
}
