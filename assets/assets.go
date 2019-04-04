package assets

//go:generate go run gen.go

import (
	"fmt"
	"html/template"
	"io/ioutil"
)

// GetTemplate returns html template.
func GetTemplate() (*template.Template, error) {
	f, err := Assets.Open("index.tmpl")
	if err != nil {
		return nil, fmt.Errorf("Cannot open asset file: %v", err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Get template read: %v", err)
	}

	tpl, err := template.New("index").Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("Get template read: %v", err)
	}
	return tpl, err
}
