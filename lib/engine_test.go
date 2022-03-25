package gtw

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Do not lightly change the test data ... it has very specific
// properties ... e.g. see the string reverse in TestScore.
const testData = "three\nblind\nmices\n"
var loadedTestData = []string{"three", "blind", "mices"}

func createTestFile(data string) (string, error) {
	tmpFile, err := ioutil.TempFile("", fmt.Sprintf("%s-", filepath.Base(os.Args[0])))
	if err != nil {
		return "", err
	}
	if _, err = tmpFile.WriteString(data); err != nil {
		tmpFile.Close()
		return "", err
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}

func TestLoadFile(t *testing.T) {
	testFileName, err := createTestFile(testData)
	if err != nil {
		t.Error("internal error creating test data", err)
	}
	words, err := LoadFile(testFileName)
	if err != nil {
		t.Error("LoadFile", err)
	}
	if len(words) != len(loadedTestData) {
		t.Errorf("LoadFile: length")
	}
	// uncomment to test the test
	// loadedTestData[1] = "wrong"
	for i, w := range loadedTestData {
		if w != words[i] {
			t.Errorf("LoadFile: word %d: got %s, expected %s\n", i, words[i], w)
		}
	}
}

func loadTestCorpus(t *testing.T) []string {
	filename, err := createTestFile(testData)
	if err != nil {
		t.Error("creating test data file (internal error?)", err)
	}

	corpus, err := LoadFile(filename)
	if err != nil {
		t.Error("loading test data file (internal error?)", err)
	}
	return corpus
}

func TestNew(t *testing.T) {
	engine := New(loadTestCorpus(t))
	goal := engine.Cheat()
	if goal != "three" && goal != "blind" && goal != "mices" {
		t.Error("after New(), goal word is not in the test data")
	}
}

const doubleTest = "taken\n"
var loadedDoubleTest = []string{"taken"}

func TestBugScore(t *testing.T) {
	testFileName, err := createTestFile(doubleTest)
	if err != nil {
		t.Error("internal error creating test data", err)
	}
	corpus, err := LoadFile(testFileName)
	if err != nil {
		t.Error("LoadFile", err)
	}
	engine := New(corpus)
	expected := "++#+#"
	signature, _ := engine.Score("tater")
	if signature != expected {
		t.Error("got ", signature, "expected", expected)
	}
}

const thirdTest = "cross\n"
var loadedThirdTest = []string{"cross"}

func TestBug3Score(t *testing.T) {
	testFileName, err := createTestFile(thirdTest)
	if err != nil {
		t.Error("internal error creating test data", err)
	}
	corpus, err := LoadFile(testFileName)
	if err != nil {
		t.Error("LoadFile", err)
	}
	engine := New(corpus)
	expected := "#+#+#"
	signature, _ := engine.Score("brush")
	if signature != expected {
		t.Error("got ", signature, "expected", expected)
	}
}

func TestScore(t *testing.T) {
	engine := New(loadTestCorpus(t))
	signature, score := engine.Score("xyzzy")
	if signature != "#####" || score != 0 {
		t.Error("wrong score for xyzzy", signature, score)
	}
	signature, score = engine.Score(engine.Cheat())
	if signature != "+++++" || score != 5 {
		t.Error("wrong score for blind", signature, score)
	}
	engine.NewGame()
	answer := engine.Cheat()
	var reversed strings.Builder
	for i := 4; i >= 0; i -= 1 {
		r := rune(answer[i])
		if i == 2 {
			r = 'z'
		}
		reversed.WriteRune(r)
	}
	signature, score = engine.Score(reversed.String())
	if signature != "**#**" || score != 0 {
		t.Error("wrong score for reversed goal", signature, score)
	}
}

func TestFixedGame(t *testing.T) {
	engine := New(loadTestCorpus(t))
	aWord := engine.Corpus()[0]
	engine.NewFixedGame(aWord)
	signature, score := engine.Score(aWord)
	if signature != "+++++" || score != 5 {
		t.Error("wrong score for specified goal word", signature, score, aWord)
	}
}

func TestHumanize(t *testing.T) {
	result := Humanize("++##*", "after")
	if result != "AF--r" {
		t.Error("Humanize(++##*, after): bad result", result)
	}
}
