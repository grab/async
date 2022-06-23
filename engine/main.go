package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/grab/async/engine/sample/config"
	"github.com/grab/async/engine/sample/server"
	"github.com/grab/async/engine/sample/service/miscellaneous"
	"github.com/grab/async/engine/sample/service/scaffolding/parallel"
	"github.com/grab/async/engine/sample/service/scaffolding/sequential"
	"golang.org/x/tools/go/packages"
)

type customPostHook struct{}

func (customPostHook) PostExecute(p any) error {
	config.Print("After sequential plan custom hook")

	return nil
}

func main() {
	// method, ok := reflect.ValueOf(dummy{}).Type().MethodByName("Do")
	// if ok {
	// 	fmt.Println(method)
	// }
	//
	// method.Func.Call([]reflect.Value{reflect.ValueOf(dummy{})})

	testParsePackage2()
}

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo

func testParsePackage2() {
	loadConfig := new(packages.Config)
	loadConfig.Mode = loadMode
	loadConfig.Fset = token.NewFileSet()
	pkgs, err := packages.Load(loadConfig, "github.com/grab/async/engine/sample/...")
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		for _, syn := range pkg.Syntax {
			for _, dec := range syn.Decls {
				if gen, ok := dec.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
					// print doc comment of the type
					// fmt.Println(gen.Doc.List[0])
					for _, spec := range gen.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							obj, ok := pkg.TypesInfo.Defs[ts.Name]
							if !ok {
								continue
							}

							typeName, ok := obj.(*types.TypeName)
							if !ok {
								continue
							}

							named, ok := typeName.Type().(*types.Named)
							if !ok {
								continue
							}

							// print the full name of the type
							fmt.Println(named)
							fmt.Println(pkg.TypesInfo.Types[ts.Type].Type)

							s, ok := named.Underlying().(*types.Struct)
							if !ok {
								continue
							}

							// print the struct's fields and tags
							for i := 0; i < s.NumFields(); i++ {
								idx := fmt.Sprint(i)
								fmt.Println("s.Field(", idx, ").Name(): ", s.Field(i).Name())
								fmt.Println("s.Tag(", idx, "): ", s.Tag(i))
							}
						}
					}
				}
			}
		}
	}

	// pkg, err := importer.Default().Import("github.com/grab/async/engine/core")
	// if err != nil {
	// 	fmt.Printf("error: %s\n", err.Error())
	// 	return
	// }
	// for _, declName := range pkg.Scope().Names() {
	// 	fmt.Println(declName)
	// }
}

func testEngine() {
	server.Serve()

	config.Engine.ConnectPostHook(&sequential.SequentialPlan{}, customPostHook{})

	p := parallel.NewPlan(
		miscellaneous.CostRequest{
			PointA: "Clementi",
			PointB: "Changi Airport",
		},
	)

	if err := p.Execute(context.Background()); err != nil {
		config.Print(err)
	}

	config.Print(p.GetTravelCost())
	config.Print(p.GetTotalCost())
}
