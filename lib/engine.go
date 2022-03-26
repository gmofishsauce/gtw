/*
Package gtw implements a word game.

Artifacts in this package are suitable for use when
implementing a user interface for the word game or
when creating bots to play the word game.

*/
package gtw

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"
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

// GtwEngine is a "game engine" for Guess the Word
type GtwEngine struct {
	corpus []string
	rng    *rand.Rand
	goal   string
}

// New creates a new GtW evaluation engine given a corpus of words.
// The corpus may be constructed by LoadFile.
func New(corpus []string) *GtwEngine {
	if len(corpus) == 0 {
		panic("0-length corpus ... ouch, don't do that")
	}
	result := &GtwEngine{corpus, nil, ""}
	result.SetSeed(-1) // random
	result.NewGame()
	return result
}

// Get the Corpus
func (e *GtwEngine) Corpus() []string {
	return e.corpus
}

// Set the seed for the RNG
func (e *GtwEngine) SetSeed(seed int64) {
	if seed < 0 {
		seed = time.Now().UnixNano()
	}
	e.rng = rand.New(rand.NewSource(seed))
}

// NewGame reinitializes the goal word of the engine to a uniformly-
// selected random word from the engine's corpus.
func (e *GtwEngine) NewGame() {
	e.goal = e.corpus[e.rng.Int31n(int32(len(e.corpus)))]
}

// NewFixedGame reinitializes the goal word to the argument
// The argument is not necessarily in the corpus.
func (e *GtwEngine) NewFixedGame(aWord string) error {
	e.goal = aWord
	return nil
}

// Cheat returns the the engine's current goal word.
func (e *GtwEngine) Cheat() string {
	return e.goal
}

const LETTER_CORRECT = '+' // This letter is correct and in position
const LETTER_IN_WORD = '*' // This letter is in the word, but out of position
const LETTER_WRONG = '#'   // This letter is not in the word at any position
const LETTER_INVALID = 0   // This can't ever occur in a guess or a goal

// Score returns two values indicating the goodness of a guess.
// The first return value is a string describing the match result
// for the guess. In this string, '+' means the letter is in the
// correct position in the goal word, '*' indicates the letter is
// in the goal word but not in the correct position, and '#' means
// the letter is not in the word. The integer value is the number
// of '+' characters in the match result string. Note: the function
// Humanize(signature, guess) can be used to produce a result string
// is easier for humans to read from the result of this method.

func (e *GtwEngine) Score(guess string) (string, int) {
	var aGuess, aGoal, signature [5]rune

	for i, _ := range(guess) {
		aGuess[i] = rune(guess[i])
		aGoal[i] = rune(e.goal[i])
		signature[i] = LETTER_WRONG
	}

	// First find all the correct matches. Once found, they
	// play no further role in matching either in the goal
	// or in the guess. Then make a second pass over the guess
	// and score any letter that still exists in the goal as
	// an out-of-place letter.

	unsolvedLetterCounts := make(map[rune]int)

	nCorrect := 0
	for i, g := range(aGuess) {
		if g == aGoal[i] {
			aGuess[i] = 0
			aGoal[i] = 0
			signature[i] = LETTER_CORRECT
			nCorrect++
		} else {
			count, exists := unsolvedLetterCounts[aGoal[i]]

			if !exists {
				count = 0
			}

			unsolvedLetterCounts[aGoal[i]] = count + 1
		}
	}

	for i, g := range(aGuess) {
		if aGuess[i] != 0 {
			count, exists := unsolvedLetterCounts[g]
			if exists && count > 0 {
				unsolvedLetterCounts[g] = count - 1
				signature[i] = LETTER_IN_WORD
			}
		}
	}

	return string(signature[:]), nCorrect
}

// Humanize the result of a guess. Given a signature like "++##*"
// and guess like "after", the result is AF--r meaning the A and F
// are correcly placed, TE are not in the goal, and r is present
// but out of place.
func Humanize(signature string, guess string) string {
	var result strings.Builder
	for i, r := range signature {
		switch r {
		case LETTER_CORRECT:
			result.WriteRune(unicode.ToUpper(rune(guess[i])))
		case LETTER_IN_WORD:
			result.WriteRune(rune(guess[i]))
		case LETTER_WRONG:
			result.WriteRune('-')
		default:
			result.WriteRune('?')
			fmt.Fprintf(os.Stderr, "humanizing result string: invalid character %c in signature\n", r)
		}
	}
	return result.String()
}

