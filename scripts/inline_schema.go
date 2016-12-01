package main

import (
	"io/ioutil"
	"os"
	"text/template"
)

func main() {
	t, err := template.New("schema_template").ParseFiles("./scripts/schema_template")
	if err != nil {
		panic(err)
	}

	schema, err := ioutil.ReadFile("./scripts/schema.json")
	if err != nil {
		panic(err)
	}

	inlinedFile, err := os.Create("config/schema.go")
	if err != nil {
		panic(err)
	}

	err = t.Execute(inlinedFile, map[string]string{
		"schema": string(schema),
	})

	if err != nil {
		panic(err)
	}
}
