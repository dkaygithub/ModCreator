package objects

import (
	"ModCreator/file"
	"path"
	"regexp"

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
	if sp, ok := o.data["LuaScript_path"]; ok {
		if spstr, ok := sp.(string); ok {
			o.luascriptPath = spstr
			delete(o.data, "LuaScript_path")
		}
	}

	return nil
}

func (o *objConfig) parseFromJSON(data map[string]interface{}) error {
	o.data = data
	dguid, ok := o.data["GUID"]
	if !ok {
		return fmt.Errorf("object (%v) doesn't have a GUID field", data)
	}
	guid, ok := dguid.(string)
	if !ok {
		return fmt.Errorf("object (%v) doesn't have a string GUID (%s)", dguid, o.data["GUID"])
	}
	o.guid = guid
	o.subObj = []*objConfig{}
	if rawObjs, ok := o.data["ContainedObjects"]; ok {
		rawArr, ok := rawObjs.([]interface{})
		if !ok {
			return fmt.Errorf("type mismatch in ContainedObjects : %v", rawArr)
		}
		for _, rawSubO := range rawArr {
			subO, ok := rawSubO.(map[string]interface{})
			if !ok {
				return fmt.Errorf("type mismatch in ContainedObjects : %v", rawSubO)
			}
			so := objConfig{}
			if err := so.parseFromJSON(subO); err != nil {
				return fmt.Errorf("printing sub object of %s : %v", o.guid, err)
			}
			o.subObj = append(o.subObj, &so)
		}
		delete(o.data, "ContainedObjects")
	}
	return nil
}

func (o *objConfig) print(l file.LuaReader) (j, error) {
	if o.luascriptPath != "" {
		encoded, err := l.EncodeFromFile(o.luascriptPath)
		if err != nil {
			return j{}, fmt.Errorf("l.EncodeFromFile(%s) : %v", o.luascriptPath, err)
		}
		o.data["LuaScript"] = encoded
	}

	subs := []j{}
	for _, sub := range o.subObj {
		printed, err := sub.print(l)
		if err != nil {
			return nil, err
		}
		subs = append(subs, printed)
	}

	o.data["ContainedObjects"] = subs

	return o.data, nil
}

func (o *objConfig) printToFile(filepath string, l file.LuaWriter) error {
	// maybe convert LuaScript or LuaScriptState
	if rawscript, ok := o.data["LuaScript"]; ok {
		if script, ok := rawscript.(string); ok {
			if len(script) > 80 {
				createdFile := o.getAGoodFileName() + ".ttslua"
				o.data["LuaScript_path"] = createdFile
				l.EncodeToFile(script, createdFile)
				delete(o.data, "LuaScript")
			}
		}
	}
	if rawscript, ok := o.data["LuaScriptState"]; ok {
		if script, ok := rawscript.(string); ok {
			if len(script) > 80 {
				createdFile := o.getAGoodFileName() + ".txt"
				o.data["LuaScriptState_path"] = createdFile
				l.EncodeToFile(script, createdFile)
				delete(o.data, "LuaScriptState")
			}
		}
	}

	// recurse if need be
	if o.subObj != nil && len(o.subObj) > 0 {
		subDir := path.Join(filepath, o.guid)
		err := os.Mkdir(subDir, 0644)
		if err != nil {
			return err
		}
		for _, subo := range o.subObj {
			err = subo.printToFile(subDir, l)
			if err != nil {
				return err
			}
		}
	}

	// print self
	b, err := json.MarshalIndent(o.data, "", "  ")
	if err != nil {
		return err
	}
	fname := path.Join(filepath, o.getAGoodFileName()+".json")
	return ioutil.WriteFile(fname, b, 0644)

}

func (o *objConfig) getAGoodFileName() string {
	// only let alphanumberic, _, -, be put into names
	reg, err := regexp.Compile("[^a-zA-Z0-9_-]+")
	if err != nil {
		return o.guid
	}
	rawName, ok := o.data["Name"]
	if !ok {
		return o.guid
	}
	name, ok := rawName.(string)
	if !ok {
		return o.guid
	}
	n := reg.ReplaceAllString(name, "")
	return n + "." + o.guid
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

func (d *db) print(l file.LuaReader) (objArray, error) {
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
func ParseAllObjectStates(root string, l file.LuaReader) ([]map[string]interface{}, error) {
	d := db{
		all: map[string]*objConfig{},
	}
	err := parseFolder(root, true, &d)
	if err != nil {
		return []map[string]interface{}{}, fmt.Errorf("parseFolder(%s): %v", root, err)
	}
	return d.print(l)
}

func parseFolder(p string, isRoot bool, d *db) error {
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir(%s) : %v", p, err)
	}
	folders := make([]fs.FileInfo, 0)
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file)
		}
		parseFile(path.Join(p, file.Name()), isRoot, d)
	}
	for _, folder := range folders {
		parseFolder(path.Join(p, folder.Name()), false, d)
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

// PrintObjectStates takes a list of json objects and prints them in the
// expected format outlined by ParseAllObjectStates
func PrintObjectStates(root string, f file.LuaWriter, objs []map[string]interface{}) error {
	for _, rootObj := range objs {
		oc := objConfig{}
		err := oc.parseFromJSON(rootObj)
		if err != nil {
			return err
		}
		err = oc.printToFile(root, f)
		if err != nil {
			return err
		}
	}
	return nil
}
