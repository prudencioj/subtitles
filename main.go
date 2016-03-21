package main

import (
	"os"
	"flag"
	"github.com/prudencioj/subtitles/subs"
	"strings"
)

func main() {
	// Path to search videos will be the current dir
	// Or the user can add a -p option
	var srcVar string
	// Get current dir
	dir, _ := os.Getwd()
	flag.StringVar(&srcVar, "p", dir, "path to search for videos")

	// -l to supply a list of languages
	var langVar string
	flag.StringVar(&langVar, "l", "en", "list of languages")
	flag.Parse()

	src := srcVar
	langs := strings.Split(langVar, ",")

	s := subs.NewDownloader()
	s.Download(src, langs)
}