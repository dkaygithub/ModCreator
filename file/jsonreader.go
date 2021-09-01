package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// JSONReader allows for arbitrary reads and encoding of json
type JSONReader struct {
	basepath string
}

// NewJSONReader initializes our object on a directory
func NewJSONReader(base string) *JSONReader {
	return &JSONReader{
		basepath: base,
	}
}

// ReadObj pulls a file from configs and encodes it as a string.
func (j *JSONReader) ReadObj(filename string) (map[string]interface{}, error) {
	b, err := j.pullRawFile(filename)
	if err != nil {
		return map[string]interface{}{}, err
	}
	var v map[string]interface{}
	json.Unmarshal(b, &v)
	return v, nil
}

// ReadObjArray pulls a file from configs and encodes it as a string.
func (j *JSONReader) ReadObjArray(filename string) ([]map[string]interface{}, error) {
	b, err := j.pullRawFile(filename)
	if err != nil {
		return []map[string]interface{}{}, err
	}
	var v []map[string]interface{}
	json.Unmarshal(b, &v)
	return v, nil

}
func (j *JSONReader) pullRawFile(filename string) ([]byte, error) {
	p := path.Join(j.basepath, filename)
	jFile, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("os.Open(%s): %v", p, err)
	}
	defer jFile.Close()

	return ioutil.ReadAll(jFile)
}
