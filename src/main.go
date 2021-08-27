package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"os"
)

var (
	path = flag.String("path", "", "a directory containing tts mod configs")
)

type Config struct {
	Name      string `json:"SaveName"`
	Version   string `json:"version,omitempty"`
	ImagePath string `json:"-"`
	ObjectDir string
}

// Mod is used as the accurate representation of what gets printed when
// module creation is done
type Mod struct {
	SaveName string

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
	log.Println(*path)
	printSample()
}

func printSample() {
	test := Config{}
	b, err := json.Marshal(test)
	if err != nil {
		log.Fatalf("Marshal() : %v\n", err)
	}
	log.Println(b)

	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")

	out.WriteTo(os.Stdout)
	os.Stdout.Write([]byte{'\n'})
	log.Println("used config found in :")
}
