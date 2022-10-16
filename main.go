package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := flag.String("root", "", "Root directory to start")
	restore := flag.String("restore", "", "Directory to restore gzip files to")

	flag.Parse()

	if err := run(os.Stdout, *root, *restore); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(out io.Writer, root, restore string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".gz" {
			return nil
		}

		if err := restoreFile(root, path, restore); err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, path)
		return err
	})
}

func restoreFile(root, path, restore string) error {
	info, err := os.Stat(restore)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", restore)
	}

	relDir, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return err
	}

	fName := strings.TrimSuffix(filepath.Base(path), ".gz")
	fPath := filepath.Join(restore, relDir, fName)

	if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(fPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer out.Close()

	gR, err := gzip.NewReader(in)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, gR); err != nil {
		return err
	}

	if err := gR.Close(); err != nil {
		return err
	}

	return out.Close()
}
