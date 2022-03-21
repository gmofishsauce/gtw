/*
Package gtw implements a word game.

Artifacts in this package are suitable for use when
implementing a user interface for the word game or
when creating bots to play the word game.

*/
package gtw

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
	"time"
)

// LoadFile loads a "corpus file" having one word per newline-separated
// line and returns the file as an array of strings, one per line.
func LoadFile(filepath string) ([]string, error) {
	words, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(words)

	wordlist := make([]string, 0, 100)
	for scanner.Scan() {
		wordlist = append(wordlist, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return wordlist, nil
}

type GtwEngine struct {
	corpus []string
	rng    *rand.Rand
	goal   string
}

// New creates a new GtW evaluation engine given a corpus of words.
// The corpus may be constructed by LoadFile.
func New(corpus []string) *GtwEngine {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	goal := corpus[rng.Int31n(int32(len(corpus)))]
	result := &GtwEngine{corpus, rng, goal}
	return result
}

// NewGame reinitializes the goal word of the engine to a uniformly-
// selected random word from the engine's corpus.
func (e *GtwEngine) NewGame() {
	e.goal = e.corpus[e.rng.Int31n(int32(len(e.corpus)))]
}

// Cheat returns the the engine's current goal word.
func (e *GtwEngine) Cheat() string {
	return e.goal
}

const LETTER_CORRECT = '+' // This letter is correct and in position
const LETTER_IN_WORD = '*' // This letter is in the word, but out of position
const LETTER_WRONG = '#'   // This letter is not in the word at any position

// Score returns two values indicating the goodness of a guess.
// The first return value is a string describing the match result
// for the guess. In this string, '+' means the letter is in the
// correct position in the goal word, '*' indicates the letter is
// in the goal word but not in the correct position, and '#' means
// the letter is not in the word. The integer value is the number
// of '+' characters in the match result string.
func (e *GtwEngine) Score(guess string) (string, int) {
	var nCorrect int
	var signature strings.Builder

	for i, r := range guess {
		if r == rune(e.goal[i]) {
			signature.WriteRune(LETTER_CORRECT)
			nCorrect += 1
		} else if strings.Contains(e.goal, string(r)) {
			signature.WriteRune(LETTER_IN_WORD)
		} else {
			signature.WriteRune(LETTER_WRONG)
		}
	}
	return signature.String(), nCorrect
}
