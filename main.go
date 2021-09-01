package main

import (
	file "ModCreator/file"
	objects "ModCreator/objects"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var (
	config = flag.String("config", "testdata/simple", "a directory containing tts mod configs")
)

const (
	luaSubdir  = "luascript"
	jsonSubdir = "json"
)

// Config is how users will specify their mod's configuration.
type Config struct {
	Name            string `json:"Name"`
	Version         string `json:"version"`
	ImagePath       string `json:"ImagePath"`
	LuaScriptPath   string `json:"LuaScriptPath"`
	LuaScriptState  string `json:"LuaScriptState"`
	TabStatesPath   string `json:"TabStatesPath"`
	MusicPlayerPath string `json:"MusicPlayerPath"`
	GridPath        string `json:"GridPath"`
	LightingPath    string `json:"LightingPath"`
	DecalPalletPath string `json:"DecalPalletPath"`
	SnapPointsPath  string `json:"SnapPointsPath"`
	ObjectDir       string `json:"ObjectDir"`
	Raw             Obj    `json:"-"`
}

// Obj is a simpler way to refer to a json map.
type Obj map[string]interface{}

// ObjArray is a simple way to refer to an array of json maps
type ObjArray []map[string]interface{}

// Mod is used as the accurate representation of what gets printed when
// module creation is done
type Mod struct {
	Data Obj
}

func main() {
	flag.Parse()

	lua := file.NewLuaReader(path.Join(*config, luaSubdir))
	j := file.NewJSONReader(path.Join(*config, jsonSubdir))

	c, err := readConfig(*config)
	if err != nil {
		fmt.Printf("readConfig(%s) : %v\n", *config, err)
		return
	}

	m, err := generateMod(*config, lua, j, c)
	if err != nil {
		fmt.Printf("generateMod(<config>) : %v\n", err)
		return
	}
	err = printMod(*config, m)
	if err != nil {
		log.Fatalf("printMod(...) : %v", err)
	}
}

func readConfig(cPath string) (*Config, error) {
	// Open our jsonFile
	cFile, err := os.Open(path.Join(cPath, "config.json"))
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, fmt.Errorf("os.Open(%s): %v", path.Join(cPath, "config.json"), err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer cFile.Close()

	b, err := ioutil.ReadAll(cFile)
	if err != nil {
		return nil, fmt.Errorf("ioutil.Readall(%s) : %v", path.Join(cPath, "config.json"), err)
	}
	var c Config

	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal(%s) : %v", b, err)
	}
	err = json.Unmarshal(b, &c.Raw)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal(%s) : %v", b, err)
	}
	return &c, nil
}

func generateMod(p string, lua *file.LuaReader, j *file.JSONReader, c *Config) (*Mod, error) {
	if c == nil {
		return nil, fmt.Errorf("nil config")
	}
	var m Mod

	m.Data = c.Raw

	plainObj := func(s string) (interface{}, error) {
		return j.ReadObj(s)
	}
	objArray := func(s string) (interface{}, error) {
		return j.ReadObjArray(s)
	}
	luaGet := func(s string) (interface{}, error) {
		return lua.EncodeFromFile(s)
	}

	tryPut(&m, c, "TabStatesPath", "TabStates", plainObj)
	tryPut(&m, c, "MusicPlayerPath", "MusicPlayer", plainObj)
	tryPut(&m, c, "GridPath", "Grid", plainObj)
	tryPut(&m, c, "LightingPath", "Lighting", plainObj)

	tryPut(&m, c, "DecalPalletPath", "DecalPallet", objArray)
	tryPut(&m, c, "SnapPointsPath", "SnapPoints", objArray)

	tryPut(&m, c, "LuaScriptPath", "LuaScript", luaGet)

	allObjs, err := objects.ParseAllObjectStates(path.Join(p, c.ObjectDir), lua)
	if err != nil {
		return nil, fmt.Errorf("objects.ParseAllObjectStates(%s) : %v", path.Join(p, c.ObjectDir), err)
	}
	m.Data["ObjectStates"] = allObjs
	delete(m.Data, "ObjectDir")
	return &m, nil
}

func tryPut(m *Mod, c *Config, from, to string, fun func(string) (interface{}, error)) {
	if m == nil || c == nil {
		log.Println("Nil objects")
		return
	}
	if m.Data == nil || c.Raw == nil {
		log.Println("Nil data inside objects")
		return
	}
	var o interface{}
	fromFile, ok := c.Raw[from]
	if !ok {
		fromFile = ""
	}
	filename, ok := fromFile.(string)
	if !ok {
		log.Printf("non string filename found: %s", fromFile)
		filename = ""
	}

	o, _ = fun(filename)
	// ignore error for now

	m.Data[to] = o
	delete(m.Data, from)
}

func printMod(p string, m *Mod) error {
	b, err := json.MarshalIndent(m.Data, "", "  ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent(<mod>) : %v", err)
	}

	return ioutil.WriteFile(path.Join(p, "output.json"), b, 0644)
}
