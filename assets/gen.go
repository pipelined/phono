// +build ignore

package main

import (
	"log"

	"github.com/pipelined/convert/assets"
	"github.com/shurcooL/vfsgen"
)

func main() {
	err := vfsgen.Generate(assets.Assets, vfsgen.Options{
		PackageName:  "assets",
		VariableName: "Assets",
		BuildTags:    "!dev",
		Filename:     "assets_data.go",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
