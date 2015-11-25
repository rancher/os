package glob

import "testing"

func testGlobMatch(t *testing.T, pattern, subj string) {
	if !Glob(pattern, subj) {
		t.Fatalf("%s should match %s", pattern, subj)
	}
}

func testGlobNoMatch(t *testing.T, pattern, subj string) {
	if Glob(pattern, subj) {
		t.Fatalf("%s should not match %s", pattern, subj)
	}
}

func TestEmptyPattern(t *testing.T) {
	testGlobMatch(t, "", "")
	testGlobNoMatch(t, "", "test")
}

func TestPatternWithoutGlobs(t *testing.T) {
	testGlobMatch(t, "test", "test")
}

func TestGlob(t *testing.T) {
	for _, pattern := range []string{
		"*test",           // Leading glob
		"this*",           // Trailing glob
		"*is *",           // String in between two globs
		"*is*a*",          // Lots of globs
		"**test**",        // Double glob characters
		"**is**a***test*", // Varying number of globs
	} {
		testGlobMatch(t, pattern, "this is a test")
	}

	for _, pattern := range []string{
		"test*", // Implicit substring match should fail
		"*is",   // Partial match should fail
		"*no*",  // Globs without a match between them should fail
	} {
		testGlobNoMatch(t, pattern, "this is a test")
	}
}
