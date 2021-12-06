package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func main(){
	// 决定运行模式
	mode := ""
	flag.StringVar(&mode,"mode","central","run mode (central or distributed)")
	// 输入目标父package
	rootPkg := ""
	flag.StringVar(&rootPkg,"pkg","","the root package to parse")
	rootPath := ""
	flag.StringVar(&rootPath,"path","./","path for target package,ends with /")
	flag.Parse()
	if rootPkg == "" {
		flag.PrintDefaults()
		return
	}
	if mode == "central"{
		CentralMain(rootPkg,rootPath)
	} else if mode == "distributed" {
		DistributedMain(rootPkg,rootPath)
	}
	
}

// 中心化模式,适合main使用，main导入registry中心包即可注册
func CentralMain(rootPkg string,rootPath string){
	// 创建registry文件
	outPath := "registryContent.go"
	rfile,err := os.Create(outPath)
	if err != nil {
		panic("Error create registry file: "+err.Error())
	}
	// 生成registry
	rfile.Write([]byte("package registry_content\n"))
	rfile.Close()
	rfset := token.NewFileSet()
	rastRoot,err := parser.ParseFile(rfset,outPath,nil,parser.AllErrors)
	astutil.AddNamedImport(rfset,rastRoot,"__registry","github.com/wshhyh/go_type_registry/registry")
	totalImports := map[string]string{}

	registerStmts := []string{}
	
	// for每个子package
	filepath.Walk(rootPath,func(path string, info fs.FileInfo, err error) error {
		if info.IsDir(){
			return nil
		}
		if !strings.HasSuffix(path,".go"){
			return nil
		}
		if strings.HasSuffix(path,"_test.go"){
			return nil
		}
		dirPath := filepath.Dir(path)
		packagePart := strings.TrimPrefix(dirPath,rootPath)
		fullPackage := rootPkg + "/" + packagePart
		fset := token.NewFileSet()
		astRoot,err := parser.ParseFile(fset,path,nil,parser.AllErrors)
		if err != nil {
			return err
		}

		// 添加import
		importName := ""
		if totalImports[fullPackage] == ""{
			pkgNames := strings.Split(fullPackage,"/")
			pkgName:= pkgNames[len(pkgNames)-1]
			importName = fmt.Sprintf("%s_%d",pkgName,len(totalImports))
			totalImports[fullPackage]=importName
		} else {
			importName = totalImports[fullPackage]
		}

		// for 每个type
		anyUsed := false
		ast.Inspect(astRoot,func(n ast.Node) bool {
			no,ok := n.(*ast.TypeSpec)
			if !ok{
				return true
			}
			typeName := no.Name.Name
			// 加入注册
			registerStmt := "var _ = __registey.Register("+"(*"+importName+"."+typeName+")"+"(nil)"+")"
			registerStmts = append(registerStmts, registerStmt)
			anyUsed = true
			return false
		})
		if anyUsed {
			astutil.AddNamedImport(rfset,rastRoot,importName,fullPackage)
		}
		return err
	})

	rfile,err = os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0600)
	buf := new(bytes.Buffer)
	err = format.Node(buf, rfset, rastRoot)
	if err != nil {
		panic(err)
	}
	rfile.Write(buf.Bytes())
	rfile.Write([]byte("\n"))


	registerLines := strings.Join(registerStmts,"\n")
	rfile.Write([]byte(registerLines))
	rfile.Close()
}

// 非中心模式,适合依赖库使用,会在每个pkg下生成当前pkg的注册 
func DistributedMain(rootPkg string,rootPath string){
	// for每个子package
	filepath.Walk(rootPath,func(path string, info fs.FileInfo, err error) error {
		if info.IsDir(){
			return nil
		}
		if !strings.HasSuffix(path,".go"){
			return nil
		}
		if strings.HasSuffix(path,"_test.go"){
			return nil
		}
		
		// 创建文件,设置好package
		dirPath := filepath.Dir(path)
		outPath := dirPath+"/__type_registry.go"
		fileExists := false
		if _, err := os.Stat(outPath); err == nil {
			fileExists = true
		  } else if os.IsNotExist(err) {
			fileExists = false
		  } else {
			panic(err)
		  }
		rfile,err := os.OpenFile(outPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		if !fileExists{
			splits := strings.Split(dirPath,"/")
			packagePart := splits[len(splits)-1]
			rfile.WriteString("package "+packagePart+"\n")
		}
		rfile.Close()
		rfset := token.NewFileSet()

		// 添加对registry的依赖
		rastRoot,err := parser.ParseFile(rfset,outPath,nil,parser.AllErrors)
		astutil.AddNamedImport(rfset,rastRoot,"__registry","github.com/myhyh/go_type_registry/registry")
		rfile,err = os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0600)
		buf := new(bytes.Buffer)
		err = format.Node(buf, rfset, rastRoot)
		if err != nil {
			panic(err)
		}
		rfile.Write(buf.Bytes())
		rfile.Write([]byte("\n"))
		rfile.Close()

		fset := token.NewFileSet()
		astRoot,err := parser.ParseFile(fset,path,nil,parser.AllErrors)
		if err != nil {
			return err
		}
		// for 每个type
		registerStmts := []string{}
		ast.Inspect(astRoot,func(n ast.Node) bool {
			no,ok := n.(*ast.TypeSpec)
			if !ok{
				return true
			}
			typeName := no.Name.Name
			// 加入注册
			registerStmt := "var _ = __register.Register("+"(*"+typeName+")"+"(nil)"+")"
			registerStmts = append(registerStmts, registerStmt)
			return false
		})
		rfile,err = os.OpenFile(outPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
		rfile.WriteString(strings.Join(registerStmts,"\n")+"\n")
		rfile.Close()

		return err
	})
}