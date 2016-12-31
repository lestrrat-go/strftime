package main

import (
	"flag"
	"fmt"
	"os"

	strftime "github.com/lestrrat/go-strftime"
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func _main() error {
	var name string
	flag.StringVar(&name, "name", "Format", "")
	flag.Parse()

	s, err := strftime.New(flag.Arg(0))
	if err != nil {
		return err
	}

	if err := s.Generate(os.Stdout, name); err != nil {
		return err
	}
	return nil
}
