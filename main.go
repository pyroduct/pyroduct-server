package main

import gy "github.com/graniticio/granitic-yaml/v2"
import "github.com/pyroduct/pyroduct-server/bindings"

func main() {
	gy.StartGraniticWithYaml(bindings.Components())
}
