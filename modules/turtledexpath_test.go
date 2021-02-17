package modules

import (
	"runtime"
	"testing"

	"github.com/turtledex/errors"
)

var (
	// TestGlobalTurtleDexPathVar tests that the NewGlobalTurtleDexPath initialization
	// works.
	TestGlobalTurtleDexPathVar TurtleDexPath = NewGlobalTurtleDexPath("/testdir")
)

// TestGlobalTurtleDexPath checks that initializing a new global siapath does not
// cause any issues.
func TestGlobalTurtleDexPath(t *testing.T) {
	sp, err := TestGlobalTurtleDexPathVar.Join("testfile")
	if err != nil {
		t.Fatal(err)
	}
	mirror, err := NewTurtleDexPath("/testdir")
	if err != nil {
		t.Fatal(err)
	}
	expected, err := mirror.Join("testfile")
	if err != nil {
		t.Fatal(err)
	}
	if !sp.Equals(expected) {
		t.Error("the separately spawned siapath should equal the global siapath")
	}
}

// TestRandomTurtleDexPath tests that RandomTurtleDexPath always returns a valid TurtleDexPath
func TestRandomTurtleDexPath(t *testing.T) {
	for i := 0; i < 1000; i++ {
		err := RandomTurtleDexPath().Validate(false)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestTurtleDexpathValidate verifies that the validate function correctly validates
// TurtleDexPaths.
func TestTurtleDexpathValidate(t *testing.T) {
	var pathtests = []struct {
		in    string
		valid bool
	}{
		{"valid/siapath", true},
		{"../../../directory/traversal", false},
		{"testpath", true},
		{"valid/siapath/../with/directory/traversal", false},
		{"validpath/test", true},
		{"..validpath/..test", true},
		{"./invalid/path", false},
		{".../path", true},
		{"valid./path", true},
		{"valid../path", true},
		{"valid/path./test", true},
		{"valid/path../test", true},
		{"test/path", true},
		{"/leading/slash", false},
		{"foo/./bar", false},
		{"", false},
		{"blank/end/", false},
		{"double//dash", false},
		{"../", false},
		{"./", false},
		{".", false},
	}
	for _, pathtest := range pathtests {
		err := ValidatePathString(pathtest.in, false)
		if err != nil && pathtest.valid {
			t.Fatal("validateTurtleDexpath failed on valid path: ", pathtest.in)
		}
		if err == nil && !pathtest.valid {
			t.Fatal("validateTurtleDexpath succeeded on invalid path: ", pathtest.in)
		}
	}
}

// TestTurtleDexpath tests that the NewTurtleDexPath, LoadString, and Join methods function correctly
func TestTurtleDexpath(t *testing.T) {
	var pathtests = []struct {
		in    string
		valid bool
		out   string
	}{
		{`\\some\\windows\\path`, true, `\\some\\windows\\path`}, // if the os is not windows this will not update the separators
		{"valid/siapath", true, "valid/siapath"},
		{`\some\back\slashes\path`, true, `\some\back\slashes\path`},
		{"../../../directory/traversal", false, ""},
		{"testpath", true, "testpath"},
		{"valid/siapath/../with/directory/traversal", false, ""},
		{"validpath/test", true, "validpath/test"},
		{"..validpath/..test", true, "..validpath/..test"},
		{"./invalid/path", false, ""},
		{".../path", true, ".../path"},
		{"valid./path", true, "valid./path"},
		{"valid../path", true, "valid../path"},
		{"valid/path./test", true, "valid/path./test"},
		{"valid/path../test", true, "valid/path../test"},
		{"test/path", true, "test/path"},
		{"/leading/slash", true, "leading/slash"}, // clean will trim leading slashes so this is a valid input
		{"foo/./bar", false, ""},
		{"", false, ""},
		{`\`, true, `\`},
		{`\\`, true, `\\`},
		{`\\\`, true, `\\\`},
		{`\\\\`, true, `\\\\`},
		{`\\\\\`, true, `\\\\\`},
		{"/", false, ""},
		{"//", false, ""},
		{"///", false, ""},
		{"////", false, ""},
		{"/////", false, ""},
		{"blank/end/", true, "blank/end"}, // clean will trim trailing slashes so this is a valid input
		{"double//dash", false, ""},
		{"../", false, ""},
		{"./", false, ""},
		{".", false, ""},
		{"dollar$sign", true, "dollar$sign"},
		{"and&sign", true, "and&sign"},
		{"single`quote", true, "single`quote"},
		{"full:colon", true, "full:colon"},
		{"semi;colon", true, "semi;colon"},
		{"hash#tag", true, "hash#tag"},
		{"percent%sign", true, "percent%sign"},
		{"at@sign", true, "at@sign"},
		{"less<than", true, "less<than"},
		{"greater>than", true, "greater>than"},
		{"equal=to", true, "equal=to"},
		{"question?mark", true, "question?mark"},
		{"open[bracket", true, "open[bracket"},
		{"close]bracket", true, "close]bracket"},
		{"open{bracket", true, "open{bracket"},
		{"close}bracket", true, "close}bracket"},
		{"carrot^top", true, "carrot^top"},
		{"pipe|pipe", true, "pipe|pipe"},
		{"tilda~tilda", true, "tilda~tilda"},
		{"plus+sign", true, "plus+sign"},
		{"minus-sign", true, "minus-sign"},
		{"under_score", true, "under_score"},
		{"comma,comma", true, "comma,comma"},
		{"apostrophy's", true, "apostrophy's"},
		{`quotation"marks`, true, `quotation"marks`},
	}
	// If the OS is windows then the windows path is valid and will be updated
	if runtime.GOOS == "windows" {
		pathtests[0].valid = true
		pathtests[0].out = `some/windows/path`
	}

	// Test NewTurtleDexPath
	for _, pathtest := range pathtests {
		sp, err := NewTurtleDexPath(pathtest.in)
		// Verify expected Error
		if err != nil && pathtest.valid {
			t.Fatal("validateTurtleDexpath failed on valid path: ", pathtest.in)
		}
		if err == nil && !pathtest.valid {
			t.Fatal("validateTurtleDexpath succeeded on invalid path: ", pathtest.in)
		}
		// Verify expected path
		if err == nil && pathtest.valid && sp.String() != pathtest.out {
			t.Fatalf("Unexpected TurtleDexPath From New; got %v, expected %v, for test %v", sp.String(), pathtest.out, pathtest.in)
		}
	}

	// Test LoadString
	var sp TurtleDexPath
	for _, pathtest := range pathtests {
		err := sp.LoadString(pathtest.in)
		// Verify expected Error
		if err != nil && pathtest.valid {
			t.Fatal("validateTurtleDexpath failed on valid path: ", pathtest.in)
		}
		if err == nil && !pathtest.valid {
			t.Fatal("validateTurtleDexpath succeeded on invalid path: ", pathtest.in)
		}
		// Verify expected path
		if err == nil && pathtest.valid && sp.String() != pathtest.out {
			t.Fatalf("Unexpected TurtleDexPath from LoadString; got %v, expected %v, for test %v", sp.String(), pathtest.out, pathtest.in)
		}
	}

	// Test Join
	sp, err := NewTurtleDexPath("test")
	if err != nil {
		t.Fatal(err)
	}
	for _, pathtest := range pathtests {
		newTurtleDexPath, err := sp.Join(pathtest.in)
		// Verify expected Error
		if err != nil && pathtest.valid {
			t.Fatal("validateTurtleDexpath failed on valid path: ", pathtest.in)
		}
		if err == nil && !pathtest.valid {
			t.Fatal("validateTurtleDexpath succeeded on invalid path: ", pathtest.in)
		}
		// Verify expected path
		if err == nil && pathtest.valid && newTurtleDexPath.String() != "test/"+pathtest.out {
			t.Fatalf("Unexpected TurtleDexPath from Join; got %v, expected %v, for test %v", newTurtleDexPath.String(), "test/"+pathtest.out, pathtest.in)
		}
	}
}

// TestTurtleDexpathRebase tests the TurtleDexPath.Rebase method.
func TestTurtleDexpathRebase(t *testing.T) {
	var rebasetests = []struct {
		oldBase string
		newBase string
		siaPath string
		result  string
	}{
		{"a/b", "a", "a/b/myfile", "a/myfile"}, // basic rebase
		{"a/b", "", "a/b/myfile", "myfile"},    // newBase is root
		{"", "b", "myfile", "b/myfile"},        // oldBase is root
		{"a/a", "a/b", "a/a", "a/b"},           // folder == oldBase
	}

	for _, test := range rebasetests {
		var oldBase, newBase TurtleDexPath
		var err1, err2 error
		if test.oldBase == "" {
			oldBase = RootTurtleDexPath()
		} else {
			oldBase, err1 = newTurtleDexPath(test.oldBase)
		}
		if test.newBase == "" {
			newBase = RootTurtleDexPath()
		} else {
			newBase, err2 = newTurtleDexPath(test.newBase)
		}
		file, err3 := newTurtleDexPath(test.siaPath)
		expectedPath, err4 := newTurtleDexPath(test.result)
		if err := errors.Compose(err1, err2, err3, err4); err != nil {
			t.Fatal(err)
		}
		// Rebase the path
		res, err := file.Rebase(oldBase, newBase)
		if err != nil {
			t.Fatal(err)
		}
		// Check result.
		if !res.Equals(expectedPath) {
			t.Fatalf("'%v' doesn't match '%v'", res.String(), expectedPath.String())
		}
	}
}

// TestTurtleDexpathDir probes the Dir function for TurtleDexPaths.
func TestTurtleDexpathDir(t *testing.T) {
	var pathtests = []struct {
		path string
		dir  string
	}{
		{"one/dir", "one"},
		{"many/more/dirs", "many/more"},
		{"nodir", ""},
		{"/leadingslash", ""},
		{"./leadingdotslash", ""},
		{"", ""},
		{".", ""},
	}
	for _, pathtest := range pathtests {
		siaPath := TurtleDexPath{
			Path: pathtest.path,
		}
		dir, err := siaPath.Dir()
		if err != nil {
			t.Errorf("Dir should not return an error %v, path %v", err, pathtest.path)
			continue
		}
		if dir.Path != pathtest.dir {
			t.Errorf("Dir %v not the same as expected dir %v ", dir.Path, pathtest.dir)
			continue
		}
	}
}

// TestTurtleDexpathName probes the Name function for TurtleDexPaths.
func TestTurtleDexpathName(t *testing.T) {
	var pathtests = []struct {
		path string
		name string
	}{
		{"one/dir", "dir"},
		{"many/more/dirs", "dirs"},
		{"nodir", "nodir"},
		{"/leadingslash", "leadingslash"},
		{"./leadingdotslash", "leadingdotslash"},
		{"", ""},
		{".", ""},
	}
	for _, pathtest := range pathtests {
		siaPath := TurtleDexPath{
			Path: pathtest.path,
		}
		name := siaPath.Name()
		if name != pathtest.name {
			t.Errorf("name %v not the same as expected name %v ", name, pathtest.name)
		}
	}
}
