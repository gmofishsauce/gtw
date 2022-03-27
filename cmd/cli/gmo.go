package main

// GTW bot. Usage: ./cli -c wordle.corpus  -s gmobot -v -- gmobot:wordle.corpus

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	// Guessers don't have access to the "game engine" but we need the
	// function LoadFile for our word frequency list which is organized
	// just like a corpus, one five-letter word per line.
	"github.com/gmofishsauce/gtw/lib"
)

const prefix = "gmobot:"

// initialization state machine
var initialized bool
var initFailed bool

// one-time initialization - all guesses come from these words
var masterWordList []string

// per-game
var guesses []string

func GmoGuess(corpus []string, scores []string, nCorrect int) string {
	if ! initialized {
		botInit(corpus)
		if len(masterWordList) == 0 {
			fmt.Fprintf(os.Stderr, "gmobot: bot initialized failed\n")
			initFailed = true
		}
		initialized = true
	}

	if initFailed {
		return "?????"
	}

	if len(scores) == 0 { //new game
		guesses = make([]string, 0, 0)
	}
	
	// fmt.Printf("gmobot: scores: %v\n", scores)
	remaining := masterWordList
	for i := range(guesses) {
		remaining = filter(remaining, guesses[i], scores[i])
	}

	frequencies := make([]float32, 0, 0) // computeLetterFrequencies(remaining) not used yet
	guess := choose(remaining, frequencies)
	guesses = append(guesses, guess)
	fmt.Printf("gmobot: guess: %s\n", guess)
	return guess
}

// filter returns a subset of the argument word list. The subset is constructed
// by removing all the words that are no longer possible given the score and
// the guess. The guess is a 5-letter word and the score is a signature returned
// from GtwEngine.Score().

func filter(words []string, guess string, score string) []string {
	var regexComponents = []string{"", "", "", "", ""}

	// Create a stoplist having all the out-of-place letters. This handles a
	// corner case where a letter is both out-of-place and wrong (this can
	// happen, another aspect of the double-letter issue). Any letter marked
	// "out of place" cannot be included in the patterns that specify "not in
	// word" later.
	stopList := ""
	for i, v := range(score) {
		if v == gtw.LETTER_IN_WORD {
			stopList = stopList + string(guess[i])
		}
	}

	// Construct negative patterns for letters neither in word nor in stoplist.
	// If the letters r, t, and e are not in the solution, we want to construct:
	// [^rte][^rte][^rte][^rte][^rte] as the regex (and then add other features)
	// fmt.Printf("filter: scores %v | guesses %v\n", scores, guesses)
	for i, r := range(score) {
		if r == gtw.LETTER_WRONG {
			letter := string(guess[i])
			for k := range regexComponents {
				// Add the letter to each growing component if it's neither present nor stopped
				// fmt.Printf("filter: k %d letter %s regexComponents[k] %s\n", k, letter, regexComponents[k])
				if !strings.Contains(regexComponents[k], letter) && !strings.Contains(stopList, letter) {
					regexComponents[k] = regexComponents[k] + letter
				}
			}
		}
	}

	// Now any letter that is out of place in the guess can be added to the
	// negative pattern at its position (only). Later, we'll deal with the
	// fact that we must also find the letter in the word at some other place.
	for i, r := range(score) {
		if r == gtw.LETTER_IN_WORD {
			regexComponents[i] = regexComponents[i] + string(r)
		}
	}

	// Now convert all the letter collections "xyz" into "[^xyz]", or
	// replace with "." (match any) if there are no letters to ^match.
	for i, s := range(regexComponents) {
		if s == "" {
			regexComponents[i] = "."
		} else {
			regexComponents[i] = "[^" + s + "]"
		}
	}

	// Next, install all the "correct" letters. Each correct letter
	// we see in the score will replace one of the pattern components
	// we just constructed.
	for i, r := range(score) {
		if r == gtw.LETTER_CORRECT {
			regexComponents[i] = string(guess[i])
		}
	}

	// Compile the regex
	re := regexComponents[0] + regexComponents[1] + regexComponents[2] + regexComponents[3] + regexComponents[4]
	matcher, err := regexp.Compile(re)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gmobot: regex.Compile(): %s\n", err);
		return masterWordList
	}

	// Run the regex over the master word list
	var result = make([]string, 0, 100)
	for _, s := range(words) {
		if matcher.FindString(s) != "" {
			result = append(result, s)
		}
	}

	// Remove the guess from the possible words
	if i := findStringInSlice(guess, result); i >= 0 {
		result[i] = result[len(result)-1]
		result = result[:len(result)-1]
	}

	fmt.Printf("gmobot: regex: %s in %d out %d\n", re, len(words), len(result))
	// fmt.Printf("gmobot: remaining possibles: %d\n", len(result))
	return result
}

func findStringInSlice(s string, slice []string) int {
	for i, in := range slice {
		if s == in {
			return i
		}
	}
	return -1
}


func choose(possible []string, letterFreqs []float32) string {
	if len(possible) == 0 || len(possible[0]) != 5 {
		return "badly"
	}
	return possible[0]
}

// Initialize the bot. Caller determines success or failure by checking
// the top level variables we're supposed to set.
func botInit(corpus []string) {
	// First load the word frequency list
	for _, s := range os.Args {
		if strings.HasPrefix(s, prefix) {
			params := strings.Split(strings.Replace(s, prefix, "", -1), ",")
			if len(params) >= 1 {
				wf, err := gtw.LoadFile(params[0])
				if err == nil {
					masterWordList = wf
				} else {
					fmt.Fprintf(os.Stderr, "%s %s\n", prefix, err)
				}
			}
			break
		}
	}
}
