package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// LuaReader allows for arbitrary reads and encoding of luascript
type LuaReader struct {
	basepath string
}

// NewReader initializes our object on a directory
func NewReader(base string) *LuaReader {
	return &LuaReader{
		basepath: base,
	}
}

// EncodeFromFile pulls a file from configs and encodes it as a string.
func (l *LuaReader) EncodeFromFile(filename string) (string, error) {
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
