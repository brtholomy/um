package pipe

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

const Newline string = "\n"

func isStdinLoaded() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func GetStdin() ([]string, error) {
	if !isStdinLoaded() {
		return nil, errors.New("stdin not loaded")
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: there's got to be a better way:
	s, _ := strings.CutSuffix(string(data), Newline)
	return strings.Split(s, Newline), nil
}
