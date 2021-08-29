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

type Config struct {
	Name      string `json:"Name"`
	Version   string `json:"version"`
	ImagePath string `json:"ImagePath"`
	ObjectDir string
}

// Mod is used as the accurate representation of what gets printed when
// module creation is done
type Mod struct {
	SaveName     string
	EpochTime    int64
	Date         string
	Tags         []string
	LuaScript    string
	Decals       []*Decal
	ObjectStates []*Object
	SnapPoints   []*SnapPoint
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
	m, err := generateMod(c)
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

	b, _ := ioutil.ReadAll(cFile)

	var c Config

	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal(%s) : %v", b, err)
	}
	return &c, nil
}

func generateMod(c *Config) (*Mod, error) {
	if c == nil {
		return nil, fmt.Errorf("nil config")
	}
	var m Mod

	m.SaveName = c.Name

	return &m, nil
}

func printMod(p string, m *Mod) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent(<mod>) : %v", err)
	}

	return ioutil.WriteFile(p+"/output.json", b, 0644)
}
