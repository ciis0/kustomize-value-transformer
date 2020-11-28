package main

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	"io/ioutil"
	"log"
	"os"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/walk"
	"text/template"
)

// created based kio example
// https://pkg.go.dev/sigs.k8s.io/kustomize/kyaml@v0.10.1/kio#example-package
func main() {

	yamlFile, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	v := Visitor{}

	err = yaml.Unmarshal(yamlFile, &v)
	if err != nil {
		panic(err)
	}

	fn := kio.FilterFunc(func(operand []*yaml.RNode) ([]*yaml.RNode, error) {

		for i := range operand {
			resource := operand[i]
			_, err := walk.Walker{
				Visitor: v,
				Sources: []*yaml.RNode{resource}}.Walk()
			if err != nil {
				return nil, err
			}
		}
		return operand, nil
	})

	err = kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: os.Stdin}},
		Filters: []kio.Filter{fn},
		Outputs: []kio.Writer{kio.ByteWriter{Writer: os.Stdout}},
	}.Execute()
	if err != nil {
		log.Fatal(err)
	}

}

type Visitor struct {
	Values map[string]string `yaml:"values"`
}

func (m Visitor) VisitScalar(nodes walk.Sources, _ *openapi.ResourceSchema) (*yaml.RNode, error) {

	str := nodes.Dest().YNode().Value
	buf := new(bytes.Buffer)

	tmpl, err := template.New("test").Option("missingkey=error").Delims("§((", "))").Funcs(sprig.TxtFuncMap()).Parse(str)

	if err != nil {
		return nil, err
	}

	err = tmpl.Execute(buf, m.Values)

	if err != nil {
		return nil, err
	}

	return yaml.NewScalarRNode(buf.String()), nil
}

func (m Visitor) VisitMap(nodes walk.Sources, _ *openapi.ResourceSchema) (*yaml.RNode, error) {
	return nodes.Dest(), nil
}

func (m Visitor) VisitList(nodes walk.Sources, _ *openapi.ResourceSchema, _ walk.ListKind) (*yaml.RNode, error) {
	return nodes.Dest(), nil
}
