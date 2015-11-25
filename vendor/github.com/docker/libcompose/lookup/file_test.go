package lookup

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

type input struct {
	file       string
	relativeTo string
}

func TestLookupError(t *testing.T) {
	invalids := map[input]string{
		input{"", ""}:                             "read .: is a directory",
		input{"", "/tmp/"}:                        "read /tmp: is a directory",
		input{"file", "/does/not/exists/"}:        "open /does/not/exists/file: no such file or directory",
		input{"file", "/does/not/something"}:      "open /does/not/file: no such file or directory",
		input{"file", "/does/not/exists/another"}: "open /does/not/exists/file: no such file or directory",
		input{"/does/not/exists/file", "/tmp/"}:   "open /does/not/exists/file: no such file or directory",
		input{"does/not/exists/file", "/tmp/"}:    "open /tmp/does/not/exists/file: no such file or directory",
	}

	fileConfigLookup := FileConfigLookup{}

	for invalid, expectedError := range invalids {
		_, _, err := fileConfigLookup.Lookup(invalid.file, invalid.relativeTo)
		if err == nil || err.Error() != expectedError {
			t.Fatalf("Expected error with '%s', got '%v'", expectedError, err)
		}
	}
}

func TestLookupOK(t *testing.T) {
	tmpFolder, err := ioutil.TempDir("", "lookup-tests")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile1 := filepath.Join(tmpFolder, "file1")
	tmpFile2 := filepath.Join(tmpFolder, "file2")
	if err = ioutil.WriteFile(tmpFile1, []byte("content1"), 0755); err != nil {
		t.Fatal(err)
	}
	if err = ioutil.WriteFile(tmpFile2, []byte("content2"), 0755); err != nil {
		t.Fatal(err)
	}

	fileConfigLookup := FileConfigLookup{}

	valids := map[input]string{
		input{"file1", tmpFolder + "/"}:     "content1",
		input{"file2", tmpFolder + "/"}:     "content2",
		input{tmpFile1, tmpFolder}:          "content1",
		input{tmpFile1, "/does/not/exists"}: "content1",
		input{"file2", tmpFile1}:            "content2",
	}

	for valid, expectedContent := range valids {
		out, _, err := fileConfigLookup.Lookup(valid.file, valid.relativeTo)
		if err != nil || string(out) != expectedContent {
			t.Fatalf("Expected %s to contains '%s', got %s, %v.", valid.file, expectedContent, out, err)
		}
	}
}
