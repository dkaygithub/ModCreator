package objects

import (
	"ModCreator/file"

	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
)

type j map[string]interface{}
type objArray []map[string]interface{}

type objConfig struct {
	guid          string
	data          j
	luascriptPath string
	subObj        []*objConfig
}

func (o *objConfig) parseFromFile(filepath string) error {
	jFile, err := os.Open(filepath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return fmt.Errorf("os.Open(%s): %v", filepath, err)
	}
	defer jFile.Close()

	b, err := ioutil.ReadAll(jFile)
	if err != nil {
		return err
	}

	json.Unmarshal([]byte(b), &o.data)

	dguid, ok := o.data["GUID"]
	if !ok {
		return fmt.Errorf("object at (%s) doesn't have a GUID field", filepath)
	}
	guid, ok := dguid.(string)
	if !ok {
		return fmt.Errorf("object at (%s) doesn't have a string GUID (%s)", filepath, o.data["GUID"])
	}
	o.guid = guid

	// TODO nead ability to read from script folder
	if sp, ok := o.data["LuaScriptPath"]; ok {
		if spstr, ok := sp.(string); ok {
			o.luascriptPath = spstr
			delete(o.data, "LuaScriptPath")
		}
	}

	return nil
}

func (o *objConfig) print(l *file.LuaReader) (j, error) {
	if o.luascriptPath != "" {
		encoded, err := l.EncodeFromFile(o.luascriptPath)
		if err != nil {
			return j{}, fmt.Errorf("l.EncodeFromFile(%s) : %v", o.luascriptPath, err)
		}
		o.data["LuaScript"] = encoded
	}
	return o.data, nil
}

type db struct {
	root []*objConfig

	all map[string]*objConfig
}

func (d *db) addObj(filepath string, isRoot bool) error {
	var o objConfig
	err := o.parseFromFile(filepath)
	if err != nil {
		return fmt.Errorf("objConfig.parseFromFile(%s) : %v", filepath, err)
	}

	if isRoot {
		d.root = append(d.root, &o)
	} else {
		// find parent based on filepath name

		paths := strings.Split(filepath, "/")
		if len(paths) < 2 {
			return fmt.Errorf("could not identify parent path in %s", filepath)
		}
		folderName := paths[len(paths)-2]
		parent, ok := d.all[folderName]
		if !ok {
			return fmt.Errorf("could not find object with guid %s, looking from %s", folderName, filepath)
		}

		parent.subObj = append(parent.subObj, &o)
	}
	d.all[o.guid] = &o
	return nil
}

func (d *db) print(l *file.LuaReader) (objArray, error) {
	var oa objArray
	for _, o := range d.root {
		printed, err := o.print(l)
		if err != nil {
			return objArray{}, fmt.Errorf("obj (%s) did not print : %v", o.guid, err)
		}
		oa = append(oa, printed)
	}
	return oa, nil
}

// ParseAllObjectStates looks at a folder and creates a json map from it.
// It assumes that folder names under the 'objects' directory are valid guids
// of existing Objects.
// like:
// objects/
// --foo.json (guid=1234)
// --bar.json (guid=888)
// --888/
//    --baz.json (guid=999) << this is a child of bar.json
func ParseAllObjectStates(root string, l *file.LuaReader) ([]map[string]interface{}, error) {
	d := db{
		all: map[string]*objConfig{},
	}
	err := parseFolder(root, true, &d)
	if err != nil {
		return []map[string]interface{}{}, fmt.Errorf("parseFolder(%s): %v", root, err)
	}
	return d.print(l)
}

func parseFolder(path string, isRoot bool, d *db) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir(%s) : %v", path, err)
	}
	folders := make([]fs.FileInfo, 0)
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file)
		}
		parseFile(path+"/"+file.Name(), isRoot, d)
	}
	for _, folder := range folders {
		parseFolder(path+"/"+folder.Name(), false, d)
	}
	return nil
}

func parseFile(filepath string, isRoot bool, d *db) error {
	var o objConfig
	err := o.parseFromFile(filepath)
	if err != nil {
		return fmt.Errorf("parseFromFile(%s) : %v", filepath, err)
	}

	return d.addObj(filepath, isRoot)
}
