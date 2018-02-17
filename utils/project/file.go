package project

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/xlab/treeprint"

	"github.com/gomeet/gomeet/utils/project/helpers"
	"github.com/gomeet/gomeet/utils/project/templates"
	tmplHelpers "github.com/gomeet/gomeet/utils/project/templates/helpers"
)

type grpcMethod struct {
	File    *descriptor.FileDescriptorProto
	Service *descriptor.ServiceDescriptorProto
	Method  *descriptor.MethodDescriptorProto
}

type viewData struct {
	*Project
	GrpcMethod *grpcMethod
	//GrpcMethod *descriptor.MethodDescriptorProto
}

type file struct {
	Name         string
	AbsPath      string
	Template     string
	KeepIfExists bool
	GrpcMethod   *grpcMethod
	//GrpcMethod   *descriptor.MethodDescriptorProto
}

type folder struct {
	Name    string
	AbsPath string

	// Unexported so you can't set them without methods
	files   map[string]*file
	folders map[string]*folder
}

func newFolder(name, absPath string) *folder {
	return &folder{
		Name:    name,
		AbsPath: absPath,
		files:   make(map[string]*file),
		folders: make(map[string]*folder),
	}
}

func (f *folder) existFolder(name string) bool {
	for _, v := range f.folders {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (f *folder) existFile(name string) bool {
	for _, v := range f.files {
		if v.Name == name {
			return true
		}
	}
	return false
}

func (f *folder) addFolder(name string) *folder {
	if !f.existFolder(name) {
		newF := newFolder(name, filepath.Join(f.AbsPath, name))
		f.folders[name] = newF
		return newF
	}
	return f
}

func (f *folder) getFolder(name string) *folder {
	if nextF, ok := f.folders[name]; ok {
		return nextF
	} else if f.Name == name {
		return f
	} else {
		return &folder{}
	}
}

func (f *folder) getFile(name string) (*file, error) {
	if nextF, ok := f.files[name]; ok {
		return nextF, nil
	}

	return nil, fmt.Errorf("%s file not found", name)
}

func (f *folder) delete(name string) {
	dir := filepath.Dir(name)
	if dir != "." {
		split := strings.Split(dir, string(filepath.Separator))
		if len(split) > 1 {
			for _, s := range split {
				if f.existFolder(s) {
					f = f.getFolder(s)
				}
			}
		}
		name = filepath.Base(name)
	}

	if f.existFile(name) {
		delete(f.files, name)
	}
	if f.existFolder(name) {
		delete(f.folders, name)
	}
}

func (f *folder) renameFile(sName, dName string) error {
	if sName == "" || dName == "" {
		return errors.New("empty values")
	}
	// src nfo
	sSeg := strings.Split(sName, string(filepath.Separator))
	sL := len(sSeg)
	sF := f
	sShortName := sSeg[sL-1]
	if sL > 1 {
		for _, s := range sSeg[:sL-1] {
			if !sF.existFolder(s) {
				return fmt.Errorf("source file %s not found 1", sName)
			}
			sF = sF.getFolder(s)
		}
	}

	if !sF.existFile(sShortName) {
		return fmt.Errorf("source file %s not found 2", sName)
	}

	// dest nfo
	dSeg := strings.Split(dName, string(filepath.Separator))
	dL := len(dSeg)
	dF := f
	dShortName := dSeg[sL-1]
	if dL > 1 {
		for _, s := range sSeg[:dL-1] {
			if !dF.existFolder(s) {
				return fmt.Errorf("destination folder %s not found", filepath.Join(dSeg[:dL-1]...))
			}
			dF = dF.getFolder(s)
		}
	}

	if dF.existFolder(dShortName) || dF.existFile(dShortName) {
		return fmt.Errorf("destination file %s exist", dName)
	}

	if sF.existFolder(sShortName) {
		// rename folder
		dF.folders[sShortName] = sF.folders[sShortName]
		dF.folders[dShortName].Name = dShortName
		sF.delete(sShortName)

		return nil
	}

	// src is file and dest doen't exist
	file, err := sF.getFile(sShortName)
	if err != nil {
		return err
	}
	dF.addFile(dShortName, file.Template, file.GrpcMethod, file.KeepIfExists)
	sF.delete(sShortName)

	return nil
}

//func (f *folder) addFile(name, tmpl string, grpcMethod *descriptor.MethodDescriptorProto, keepIfExists bool) {
func (f *folder) addFile(name, tmpl string, grpcM *grpcMethod, keepIfExists bool) {
	f.files[name] = &file{
		Name:         name,
		Template:     tmpl,
		AbsPath:      filepath.Join(f.AbsPath, name),
		KeepIfExists: keepIfExists,
		GrpcMethod:   grpcM,
	}
}

//func (f *folder) addTree(name, s string, grpcMethod *descriptor.MethodDescriptorProto, keepFilesIfExists bool) (*folder, error) {
func (f *folder) addTree(name, s string, grpcM *grpcMethod, keepFilesIfExists bool) (*folder, error) {
	files, err := templates.AssetDir(s)
	if err != nil {
		return nil, err
	}
	segments := strings.Split(s, string(filepath.Separator))
	baseTemplatePath := strings.Join(segments[:len(segments)-1], string(filepath.Separator))
	// added root folder
	var rootF *folder
	if name == "." {
		rootF = f
	} else {
		rootF = f.addFolder(name)
	}
	//manage all folder tree
	for _, file := range files {
		path := filepath.Join(s, file)
		if rsrc, err := filepath.Rel(baseTemplatePath, path); err == nil {
			segments := strings.Split(rsrc, string(filepath.Separator))
			subF := rootF
			for _, v := range segments[1:] {
				if subF.existFolder(v) {
					subF = subF.getFolder(v)
				} else {
					fi, err := templates.AssetInfo(path)
					if err != nil || fi.IsDir() {
						// template file is dir
						_, err := templates.AssetDir(s)
						if err != nil {
							continue
						}
						// recursion
						subF, _ = subF.addTree(v, filepath.Join(baseTemplatePath, rsrc), grpcM, keepFilesIfExists)
					} else {
						// template file is file
						n := filepath.Base(fi.Name())
						tmplSuf := ".tmpl"
						if strings.HasSuffix(n, tmplSuf) {
							n = strings.TrimSuffix(n, tmplSuf)
						}
						subF.addFile(n, filepath.Join(baseTemplatePath, rsrc), grpcM, keepFilesIfExists)
					}
				}
			}
		}
	}
	return rootF, nil
}

func (f folder) render(p Project) error {
	for _, v := range f.files {
		fileExist := false
		contents, err := templates.Asset(v.Template)
		if err != nil {
			helpers.Log(helpers.LogError, err.Error())
			continue
		}

		if strings.HasSuffix(v.AbsPath, ".png") {
			if _, err := os.Stat(v.AbsPath); err == nil {
				fileExist = true
				if v.KeepIfExists {
					helpers.Log(helpers.LogSkipping, v.AbsPath)
					continue
				}
			}
			err := ioutil.WriteFile(v.AbsPath, contents, os.ModePerm)
			if err != nil {
				helpers.Log(helpers.LogError, err.Error())
			}
			continue
		}

		t, err := template.
			New(v.Template).
			Funcs(tmplHelpers.ProtoHelpersFuncMap()).
			Parse(string(contents))
		if err != nil {
			continue
		}

		if _, err := os.Stat(v.AbsPath); err == nil {
			fileExist = true
			if v.KeepIfExists {
				helpers.Log(helpers.LogSkipping, v.AbsPath)
				continue
			}
		}

		file, err := os.Create(v.AbsPath)
		if err != nil {
			helpers.Log(helpers.LogError, err.Error())
			continue
		}
		defer file.Close()

		vData := &viewData{&p, v.GrpcMethod}
		if strings.HasSuffix(v.AbsPath, ".go") {
			var out bytes.Buffer
			err = t.Execute(&out, vData)
			if err != nil {
				helpers.Log(helpers.LogError, err.Error())
				continue
			}

			b, err := format.Source(out.Bytes())
			if err != nil {
				b = out.Bytes()
				log.Println(string(b))
				helpers.Log(helpers.LogError, err.Error())
				continue
			}

			_, err = file.Write(b)
			if err != nil {
				helpers.Log(helpers.LogError, err.Error())
				continue
			}
		} else {
			err = t.Execute(file, vData)
			if err != nil {
				helpers.Log(helpers.LogError, err.Error())
				continue
			}
		}

		if fileExist {
			helpers.Log(helpers.LogReplacing, v.AbsPath)
		} else {
			helpers.Log(helpers.LogCreating, v.AbsPath)
		}
	}

	for _, v := range f.folders {
		if _, err := os.Stat(v.AbsPath); os.IsNotExist(err) {
			err = os.Mkdir(v.AbsPath, os.ModePerm)
			if err != nil {
				helpers.Log(helpers.LogError, err.Error())
				continue
			}
			helpers.Log(helpers.LogCreating, v.AbsPath)
		}

		if err := v.render(p); err != nil {
			helpers.Log(helpers.LogError, err.Error())
			continue
		}
	}

	return nil
}

func (f folder) print() {
	t := f.tree(true, treeprint.New())
	fmt.Println(t.String())
}

func (f folder) tree(root bool, tree treeprint.Tree) treeprint.Tree {
	if !root {
		tree = tree.AddBranch(f.Name)
	}

	for _, v := range f.folders {
		v.tree(false, tree)
	}

	for _, v := range f.files {
		tree.AddNode(v.Name)
	}

	return tree
}
