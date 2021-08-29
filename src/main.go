package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

var (
	path = flag.String("path", "testdata/simple", "a directory containing tts mod configs")
)

const (
	luaSubdir  = "luascript"
	jsonSubdir = "json"
)

type Config struct {
	Name            string `json:"Name"`
	Version         string `json:"version"`
	ImagePath       string `json:"ImagePath"`
	LuaScriptPath   string `json:"LuaScriptPath"`
	LuaScriptState  string `json:"LuaScriptState"`
	TabStatesPath   string `json:"TabStatesPath"`
	MusicPlayerPath string `json:"MusicPlayerPath"`
	GridPath        string `json:"GridPath"`
	ObjectDir       string
}

// Mod is used as the accurate representation of what gets printed when
// module creation is done
type Mod struct {
	SaveName       string
	EpochTime      int64
	Date           string
	Tags           []string
	TabStates      map[string]interface{}
	MusicPlayer    map[string]interface{}
	Grid           map[string]interface{}
	LuaScript      string
	LuaScriptState string
	Decals         []*Decal
	ObjectStates   []*Object
	SnapPoints     []*SnapPoint
}

type Decal struct {
	DecalField string
}

type Object struct {
	FooObj string
}
type SnapPoint struct {
	SnapField string
}

func main() {
	flag.Parse()
	c, err := readConfig(*path)
	if err != nil {
		fmt.Printf("readConfig(%s) : %v\n", *path, err)
		return
	}
	m, err := generateMod(*path, c)
	if err != nil {
		fmt.Printf("generateMod(<config>) : %v\n", err)
		return
	}
	err = printMod(*path, m)
	if err != nil {
		log.Fatalf("printMod(...) : %v", err)
	}
}

func readConfig(cPath string) (*Config, error) {
	// Open our jsonFile
	cFile, err := os.Open(cPath + "/config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, fmt.Errorf("os.Open(%s): %v", cPath+"/config.json", err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer cFile.Close()

	b, err := ioutil.ReadAll(cFile)
	if err != nil {
		return nil, fmt.Errorf("ioutil.Readall(%s) : %v", cPath+"/config.json", err)
	}
	var c Config

	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal(%s) : %v", b, err)
	}
	return &c, nil
}

func generateMod(p string, c *Config) (*Mod, error) {
	if c == nil {
		return nil, fmt.Errorf("nil config")
	}
	var m Mod

	m.SaveName = c.Name

	err := putEncodedJSON(&m.TabStates, p, c.TabStatesPath)
	if err != nil {
		return nil, err
	}

	err = putEncodedJSON(&m.MusicPlayer, p, c.MusicPlayerPath)
	if err != nil {
		return nil, err
	}

	err = putEncodedJSON(&m.Grid, p, c.GridPath)
	if err != nil {
		return nil, err
	}

	encoded, err := encodeLuaScript(p, c.LuaScriptPath)
	if err != nil {
		return nil, fmt.Errorf("encodeLuaScript(%s) : %v", c.LuaScriptPath, err)
	}
	m.LuaScript = encoded
	m.LuaScriptState = c.LuaScriptState

	return &m, nil
}

func putEncodedJSON(to *map[string]interface{}, p, f string) error {
	jsonEnc, err := encodeJSON(p, f)
	if err != nil {
		return err
	}
	*to = jsonEnc
	return nil
}

func printMod(p string, m *Mod) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent(<mod>) : %v", err)
	}

	return ioutil.WriteFile(p+"/output.json", b, 0644)
}

func encodeLuaScript(p, f string) (string, error) {
	path := p + "/" + luaSubdir + "/" + f
	sFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		return "", fmt.Errorf("os.Open(%s): %v", path, err)
	}
	defer sFile.Close()

	b, err := ioutil.ReadAll(sFile)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func encodeJSON(p, f string) (map[string]interface{}, error) {
	path := p + "/" + jsonSubdir + "/" + f
	jFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, fmt.Errorf("os.Open(%s): %v", path, err)
	}
	defer jFile.Close()

	b, err := ioutil.ReadAll(jFile)
	if err != nil {
		return nil, err
	}

	var v map[string]interface{}
	json.Unmarshal([]byte(b), &v)

	return v, nil
}
