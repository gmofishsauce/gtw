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
	remaining := filter(scores, guesses)
	frequencies := make([]float32, 0, 0) // computeLetterFrequencies(remaining) not used yet
	guess := choose(remaining, frequencies)
	guesses = append(guesses, guess)
	fmt.Printf("gmobot: guess: %s\n", guess)
	return guess
}

// filter returns a subset of the master word list. The
// subset includes all the words left in play.
func filter(scores []string, guesses []string) []string {
	var regexComponents = []string{"", "", "", "", ""}
	if len(scores) > 0 {
		// fmt.Printf("gmobot: score: %s\n", scores[len(scores)-1])
	}

	// Create a stoplist of letters for the following stanza. This handles the
	// corner case where a letter is both out-of-place in one of the guesses
	// but also wrong in another place (this can happen, another aspect of
	// the double-letter issue). So any letter marked "out of place" anywhere
	// cannot be included in the patterns that specify "not in word" later.
	stopList := ""
	for i := 0; i < len(scores); i++ {
		for j := 0; j < len(scores[i]); j++ {
			if scores[i][j] == gtw.LETTER_IN_WORD {
				liw := string(guesses[i][j])
				if !strings.Contains(stopList, liw) {
					stopList = stopList + liw
				}
			}
		}
	}

	// Start by setting the negative patterns for letters not in the word.
	// We are contructing a pattern like '[^rte][^rte][^rte][^rte][^rte]
	// which would make sense after one guess where r, t, e were all absent.
	// fmt.Printf("filter: scores %v | guesses %v\n", scores, guesses)
	for i, s := range(scores) {
		for j, r := range(s) {
			if r == gtw.LETTER_WRONG {
				letter := string(guesses[i][j])
				for k := range regexComponents {
					// fmt.Printf("filter: k %d letter %s regexComponents[k] %s\n", k, letter, regexComponents[k])
					if !strings.Contains(regexComponents[k], letter) && !strings.Contains(stopList, letter) {
						regexComponents[k] = regexComponents[k] + letter
					}
				}
			}
		}
	}

	// Now convert all the letter collections "xyz" into "[^xyz]", or
	// replace with "." (match any) if there are no letters to ^match.
	for i, s := range(regexComponents) {
		if s == "" {
			regexComponents[i] = "."
		} else {
			regexComponents[i] = "[^" + regexComponents[i] + "]"
		}
	}

	// Next, install all the correct letters. In some cases this will
	// overwrite negative patterns created just above.
	for i, s := range(scores) {
		for j, r := range(s) {
			if r == gtw.LETTER_CORRECT {
				regexComponents[j] = string(guesses[i][j])
			}
		}
	}

	// What we have now is sufficient for minimal gameplay, but does not
	// make use of the letters that are known to be in the word but were
	// out of position in all guesses made so far. At the command lines,
	// we do this with a pipeline. I do not know how to write a single
	// regex to express "match all of the above and then match 'x' and
	// then match 'y', etc. I think the fastest solution is to accumulate
	// the valid-but-out-of-position letters in a slice of one-character
	// strings then make a pass over the remaining candidates doing a
	// contains() for each of the letters. TODO.

	re := regexComponents[0] + regexComponents[1] + regexComponents[2] + regexComponents[3] + regexComponents[4]
	// fmt.Printf("gmobot: regex: %s\n", re)
	matcher, err := regexp.Compile(re)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gmobot: regex.Compile(): %s\n", err);
		return masterWordList
	}

	// Run the regex over the master word list
	var result = make([]string, 0, 100)
	for _, s := range(masterWordList) {
		if matcher.FindString(s) != "" {
			result = append(result, s)
		}
	}

	// Remove all the previous guesses from the result
	for _, s := range guesses {
		if i := findStringInSlice(s, result); i >= 0 {
			result[i] = result[len(result)-1]
			result = result[:len(result)-1]
		}
	}

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

/*
func computeLetterFrequencies(words []string) []float32 {
	rawFrequencies := make([]float32, 26, 26)
	for _, s := range words {
		for _, r := range(s) {
			i := r - 'a'
			rawFrequencies[i] = rawFrequencies[i] + 1.0
		}
	}
	//fmt.Printf("raw frequencies: %v\n", rawFrequencies)
	leastLikely := float32(5 * len(words))
	for _, v := range(rawFrequencies) {
		if v < leastLikely {
			leastLikely = v
		}
	}
	nf := make([]float32, 26, 26)
	// When the list of words gets small, many of the
	// raw frequencies will be 0. In this case we just
	// divide by a value which is very roughly the typical
	// difference between the least and most common letters
	// in English, about 50:1.
	divisor := float32(leastLikely)
	if divisor == 0.0 {
		divisor = 0.02
	}
	for i := range(nf) {
		nf[i] = rawFrequencies[i] / divisor
	}
	//fmt.Printf("Normalized frequencies: %v\n", nf)
	return nf
}
*/
