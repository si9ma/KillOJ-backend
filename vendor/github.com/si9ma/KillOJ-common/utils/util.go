package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/si9ma/KillOJ-common/constants"
)

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)

	// Config file not found
	if err != nil {
		return nil, fmt.Errorf("Open file error: %s", err)
	}

	// Config file found, let's try to read it
	data := make([]byte, 10000)
	count, err := file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("Read from file error: %s", err)
	}

	return data[:count], nil
}

func IsDebug() bool {
	// on debug mode, write log to stdout at the same time
	debug := false
	var err error

	if e := os.Getenv(constants.EnvDebug); e != "" {
		// if error , return false
		if debug, err = strconv.ParseBool(e); err != nil {
			return false
		}
	}

	return debug
}

func MkDirAll4RelativePath(relativePath string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	absolutePath := strings.Join([]string{pwd, relativePath}, "/")
	err = MkDirAll4Path(absolutePath)

	return absolutePath, err
}

func MkDirAll4Path(p string) error {
	dir := path.Dir(p)

	// create directory if directory not exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func Lower1stCharacter(ipt string) string {
	for i, v := range ipt {
		return string(unicode.ToLower(v)) + ipt[i+1:]
	}

	return ipt
}

func Upper1stCharacter(ipt string) string {
	for i, v := range ipt {
		return string(unicode.ToUpper(v)) + ipt[i+1:]
	}

	return ipt
}

func CheckEmail(email string) bool {
	exp := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
	ok, err := regexp.Match(exp, []byte(email))
	if !ok || err != nil {
		return false
	}

	return true
}

// refer : https://studygolang.com/articles/2118
func DeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func GetUrlRoot(port string) string {
	url := os.Getenv(constants.EnvURL)
	if url == "" {
		url = "http://127.0.0.1" // default to localhost
	}

	if url != "80" && url != "" { // 80 and empty port should be ignored
		url = url + ":" + port
	}

	return url
}

func BothZeroOrNot(v1, v2 interface{}) bool {
	v1IsZero := IsZeroOfUnderlyingType(v1)
	v2IsZero := IsZeroOfUnderlyingType(v2)

	return (v1IsZero && v2IsZero) || (!v1IsZero && !v2IsZero)
}

// refer : https://stackoverflow.com/a/13906031/8042600
func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func EscapeDoubleQuotes(str string) string {
	return strings.ReplaceAll(str, `"`, `\"`)
}
