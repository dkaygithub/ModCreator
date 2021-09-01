package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// LuaOps allows for arbitrary reads and writes of luascript
type LuaOps struct {
	basepath string
}

// LuaReader serves to describe all ways to read luascripts
type LuaReader interface {
	EncodeFromFile(string) (string, error)
}

// LuaWriter serves to describe all ways to write luascripts
type LuaWriter interface {
	EncodeToFile(script, file string) error
}

// NewLuaOps initializes our object on a directory
func NewLuaOps(base string) *LuaOps {
	return &LuaOps{
		basepath: base,
	}
}

// EncodeFromFile pulls a file from configs and encodes it as a string.
func (l *LuaOps) EncodeFromFile(filename string) (string, error) {
	p := path.Join(l.basepath, filename)
	sFile, err := os.Open(p)
	if err != nil {
		return "", fmt.Errorf("os.Open(%s): %v", p, err)
	}
	defer sFile.Close()

	b, err := ioutil.ReadAll(sFile)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// EncodeToFile takes a single string and decodes escape characters; writes it.
func (l *LuaOps) EncodeToFile(script, file string) error {
	p := path.Join(l.basepath, file)
	return os.WriteFile(p, []byte(script), 0644)
}
