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
	ObjectDir       string
}

// Obj is a simpler way to refer to a json map.
type Obj map[string]interface{}

// ObjArray is a simple way to refer to an array of json maps
type ObjArray []Obj

// Mod is used as the accurate representation of what gets printed when
// module creation is done
type Mod struct {
	SaveName       string
	EpochTime      int64
	Date           string
	Tags           []string
	TabStates      Obj
	MusicPlayer    Obj
	Grid           Obj
	Lighting       Obj
	DecalPallet    ObjArray
	LuaScript      string
	LuaScriptState string
	Decals         []*Decal
	ObjectStates   []*Object
	SnapPoints     ObjArray
}

type Decal struct {
	DecalField string
}

type Object struct {
	FooObj string
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

	putEncodedJSON(&m.TabStates, p, c.TabStatesPath)

	putEncodedJSON(&m.MusicPlayer, p, c.MusicPlayerPath)

	putEncodedJSON(&m.Grid, p, c.GridPath)

	putEncodedJSON(&m.Lighting, p, c.LightingPath)

	putEncodedJSONArray(&m.DecalPallet, p, c.DecalPalletPath)
	putEncodedJSONArray(&m.SnapPoints, p, c.DecalPalletPath)

	encoded, err := encodeLuaScript(p, c.LuaScriptPath)
	if err != nil {
		return nil, fmt.Errorf("encodeLuaScript(%s) : %v", c.LuaScriptPath, err)
	}
	m.LuaScript = encoded
	m.LuaScriptState = c.LuaScriptState

	return &m, nil
}

func putEncodedJSON(to *Obj, p, f string) {
	jsonEnc, err := encodeJSON(p, f)
	if err != nil {
		log.Printf("encodeJson(%s,%s) : %v\n", p, f, err)
		*to = Obj{}
		return
	}
	*to = jsonEnc
	return
}

func putEncodedJSONArray(to *ObjArray, p, f string) {
	jsonEnc, err := encodeJSONArray(p, f)
	if err != nil {
		log.Printf("encodeJsonArray(%s,%s) : %v\n", p, f, err)
		*to = ObjArray{}
		return
	}
	*to = jsonEnc
	return
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

func encodeJSON(p, f string) (Obj, error) {
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

	var v Obj
	json.Unmarshal([]byte(b), &v)

	return v, nil
}

func encodeJSONArray(p, f string) (ObjArray, error) {
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

	var v ObjArray
	json.Unmarshal([]byte(b), &v)

	return v, nil
}
