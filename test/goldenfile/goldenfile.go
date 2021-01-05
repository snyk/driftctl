package goldenfile

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const GoldenFilePath = "test"

var Update = flag.String("update", "", "name of test to update")

func ReadFile(p string, name string) []byte {
	p = path.Join(GoldenFilePath, p)

	content, err := ioutil.ReadFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)))
	if err != nil {
		panic(err)
	}
	return content
}

func WriteFile(p string, content []byte, name string) {
	output := content
	p = path.Join(GoldenFilePath, p)
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		panic(err)
	}
	var indentBuffer bytes.Buffer
	err := json.Indent(&indentBuffer, output, "", " ")
	if err == nil {
		output = indentBuffer.Bytes()
	}
	if err != nil {
		logrus.Error(err)
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)), output, os.ModePerm); err != nil {
		panic(err)
	}
}

// Remove forbidden characters like / in file name
func sanitizeName(name string) string {
	substitution := "_"
	replacer := strings.NewReplacer(
		"/", substitution,
		"\\", substitution,
		"<", substitution,
		">", substitution,
		":", substitution,
		"\"", substitution,
		"|", substitution,
		"?", substitution,
		"*", substitution,
	)
	return replacer.Replace(name)
}

func FileExists(dirname, f string) bool {
	fileName := path.Join(GoldenFilePath, dirname, sanitizeName(f))
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
