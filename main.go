package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"
)

//go:embed enum.go.tmpl
var tmplBytes []byte

//go:embed enum_test.go.tmpl
var tmplTestBytes []byte

type enum struct {
	Name    string
	Package string
	Type    string
	Values  []value
}

type value struct {
	Name string
}

var (
	packageName string
	enumName    string
	enumValues  enumValuesList
	showHelp    bool
	withTests   bool
)

type enumValuesList []string

func (evl *enumValuesList) Set(value string) error {
	*evl = append(*evl, strings.Split(value, ",")...)
	return nil
}

func (evl *enumValuesList) String() string {
	return fmt.Sprint(*evl)
}

func init() {
	flag.BoolVar(&showHelp, "help", false, "Show usage information.")
	flag.StringVar(&enumName, "name", "", "the name of the enum to generate")
	flag.StringVar(&packageName, "pkg", "", "the package name to generate the file for")
	flag.BoolVar(&withTests, "tests", true, "Include an auto-generated test for the enum type.")
}

func main() {
	flag.Var(&enumValues, "values", "A comma-separated list of the enum values to generate")
	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if len(enumName) == 0 {
		fmt.Println("You must set a name to generate the enum with.")
		flag.Usage()
		os.Exit(1)
	}

	if len(packageName) == 0 {
		fmt.Println("You must set a package name to generate the enum with.")
		flag.Usage()
		os.Exit(1)
	}

	if len(enumValues) == 0 {
		fmt.Println("You must set values to generate your enum with.")
		flag.Usage()
		os.Exit(1)
	}

	enumT, err := enumType(int64(len(enumValues)))
	if err != nil {
		fmt.Println("The number of values must be a valid int64.")
		flag.Usage()
		os.Exit(1)
	}

	templateFuncs := template.FuncMap{
		"Title": strings.Title,
		"Upper": strings.ToUpper,
	}

	enumTmpl, err := template.New("enum").Funcs(templateFuncs).Parse(string(tmplBytes))
	if err != nil {
		log.Fatalf("Unable to parse template file enum.go.tmpl with error: %v", err)
	}

	enumTestTmpl, err := template.New("enum").Funcs(templateFuncs).Parse(string(tmplTestBytes))
	if err != nil {
		log.Fatalf("Unable to parse template file enum_test.go.tmpl with error: %v", err)
	}

	data := enum{
		Name:    enumName,
		Package: packageName,
		Type:    enumT,
		Values:  make([]value, 0, len(enumValues)),
	}

	for _, ev := range enumValues {
		data.Values = append(data.Values, value{Name: ev})
	}

	bufTmpl := new(bytes.Buffer)
	err = enumTmpl.Execute(bufTmpl, data)
	if err != nil {
		log.Fatalf("Unable to execute template for generated file %s_gen.go with error: %v", data.Name, err)
	}

	enumFile, err := os.Create(fmt.Sprintf("./%s_gen.go", strings.ToLower(data.Name)))
	if err != nil {
		log.Fatalf("Unable to create generated file %s_gen.go with error: %v", data.Name, err)
	}

	formatted, err := format.Source(bufTmpl.Bytes())
	if err != nil {
		log.Fatalf("Unable to format data for generated file %s_gen.go with error: %v", data.Name, err)
	}

	_, err = enumFile.Write(formatted)
	if err != nil {
		log.Fatalf("Unable to write data to generated file %s_gen.go with error: %v", data.Name, err)
	}

	if !withTests {
		return
	}

	bufTestTmpl := new(bytes.Buffer)
	err = enumTestTmpl.Execute(bufTestTmpl, data)
	if err != nil {
		log.Fatalf("Unable to execute template for generated file %s_gen.go with error: %v", data.Name, err)
	}

	enumTestFile, err := os.Create(fmt.Sprintf("./%s_gen_test.go", strings.ToLower(data.Name)))
	if err != nil {
		log.Fatalf("Unable to create generated file %s_gen.go with error: %v", data.Name, err)
	}

	formattedTest, err := format.Source(bufTestTmpl.Bytes())
	if err != nil {
		log.Fatalf("Unable to format data for generated file %s_gen.go with error: %v", data.Name, err)
	}

	_, err = enumTestFile.Write(formattedTest)
	if err != nil {
		log.Fatalf("Unable to write data to generated file %s_gen.go with error: %v", data.Name, err)
	}

	log.Println("Finished!")
}

func enumType(valuesSize int64) (string, error) {
	if valuesSize < 255 {
		return "uint8", nil
	}

	if valuesSize < 65535 {
		return "uint16", nil
	}

	if valuesSize < 4294967295 {
		return "uint32", nil
	}

	if valuesSize < 9223372036854775807 {
		return "uint64", nil
	}

	return "", fmt.Errorf("%d is too many values to create a constant for", valuesSize)
}
