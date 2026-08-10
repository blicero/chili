// /home/krylon/go/src/github.com/blicero/chili/probe/01_dmi_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 10. 08. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-08-10 12:56:30 krylon>

package probe

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blicero/chili/model"
)

func TestParseDMI(t *testing.T) {
	var (
		err      error
		files    []fs.DirEntry
		testdata fs.ReadDirFS
	)

	if tprobe == nil {
		t.SkipNow()
	}

	testdata = os.DirFS("testdata").(fs.ReadDirFS)

	if files, err = testdata.ReadDir("."); err != nil {
		t.Fatalf("Cannot open testdata folder: %s",
			err.Error())
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "dmidecode_") {
			var fullPath = filepath.Join("testdata", f.Name())
			if err = testParseFile(t, fullPath); err != nil {
				t.Fatalf("Failed to process %s: %s",
					fullPath,
					err.Error())
			}
		}
	}
} // func TestParseDMI(t *testing.T)

func testParseFile(t *testing.T, path string) error {
	var (
		err   error
		fh    *os.File
		buf   []byte
		lines []string
		dmi   *model.DMI
	)

	if fh, err = os.Open(path); err != nil {
		t.Errorf("Failed to open %s: %s",
			path,
			err.Error())
		return err
	}

	defer fh.Close() // nolint: errcheck

	if buf, err = io.ReadAll(fh); err != nil {
		t.Errorf("Failed to read %s: %s",
			path,
			err.Error())
		return err
	}

	lines = strings.Split(string(buf), "\n")

	if dmi, err = tprobe.parseDMI(lines); err != nil {
		t.Errorf("Failed to parse %s: %s",
			path,
			err.Error())
		return err
	}

	t.Logf("Found this info: %s",
		dmi)

	return nil
} // func testParseFile(t *testing.T, path string)
