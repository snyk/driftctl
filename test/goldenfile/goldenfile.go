package goldenfile

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	GoldenFilePath  = "test"
	ResultsFilename = "results.golden.json"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "../..")
)

var Update = flag.String("update", "", "name of test to update")

func ReadRootFile(p string, name string) []byte {
	p = path.Join(path.Join(Root, GoldenFilePath), p)

	_, err := os.Stat(path.Join(p, ResultsFilename))
	if os.IsNotExist(err) && name == ResultsFilename {
		return []byte("[]")
	}

	content, err := os.ReadFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)))
	if err != nil {
		panic(err)
	}
	return content
}

func WriteRootFile(p string, content []byte, name string) {
	output := content

	// Avoid creating golden files for empty results
	if name == ResultsFilename && string(output) == "[]" {
		return
	}

	p = path.Join(path.Join(Root, GoldenFilePath), p)
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

	if err := os.WriteFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)), output, os.ModePerm); err != nil {
		panic(err)
	}
}

func ReadFile(p string, name string) []byte {
	p = path.Join(GoldenFilePath, p)

	_, err := os.Stat(path.Join(p, ResultsFilename))
	if os.IsNotExist(err) && name == ResultsFilename {
		return []byte("[]")
	}

	content, err := os.ReadFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)))
	if err != nil {
		panic(err)
	}
	return content
}

func WriteFile(p string, content []byte, name string) {
	output := content

	// Avoid creating golden files for empty results
	if name == ResultsFilename && string(output) == "[]" {
		return
	}

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

	if err := os.WriteFile(fmt.Sprintf("%s%c%s", p, os.PathSeparator, sanitizeName(name)), output, os.ModePerm); err != nil {
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
