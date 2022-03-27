/*
Package cli implements a command line interface to play a word game.

When given a file containing a corpus of 5-letter words, one per line,
allows the game to be played interactively from the command line. The
interactive component produces output as described below. By default,
the game core produces no output for guesses but produces a summary
at the end of its run. Use -v for more output. The primary purpose of
the cli is to run a large number of games over a set of "guessers"
(bots) built into the code. Use -h to see other command line options.

Interactive component: after each incorrect guess, a signature will
be displayed. In the signature, the character '-' means the letter
is not in the word. Lower case letters are not in the right place,
while upper case letters are correctly placed.  Example:

guess> tears
       --ers (0 letters in the correct place)
guess> cloud
       -l-u- (0 letters in the correct place)
guess> aural
       *URAL (4 letters in the correct place)
guess> rural

*/

package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/gmofishsauce/gtw/lib"
)

// Guesser is the interface to be implemented by GTW bots. A Guesser
// is passed a corpus of words, a history of past "scores" for this
// game, and the number of correct letters in the previous guess.
// The corpus is a slice of 5-letter strings. The scores are a slice
// of signatures, 5-letter strings in which "+" indicates a correct
// letter, "#" indicates an incorrect letter, and "*" indicates a
// letter in the word but not in the correct location. The bot can
// deduce the start of a new game (i.e. new goal word) when the scores
// slice is 0-length.
type Guesser interface{
	Guess(corpus []string, scores []string, nCorrect int) string
}

// Golang allows functions to implement interfaces. This adapter with
// the signature of a Guesser supports this, making it unnecessary to
// create an object to implement the interface. See the "ui" strategy
// in the strategies array below for an example
type GuesserFunc func([]string, []string, int) string

func (f GuesserFunc) Guess(c []string, s []string, n int) string {
	return f(c, s, n)
}

// Each Guesser bot is defined by a Strategy instance.
type Strategy struct {
	name string
	bot  Guesser 
	interactive bool
}

// Add your strategy here. Names must be unique and should be short for
// convenience when constructing command lines. The bots "pathetic" and
// "amazing" are intended for early testing and will be removed.
var registeredStrategies = []Strategy {
	Strategy{name: "gmobot", bot: GuesserFunc(GmoGuess), interactive: false},
	Strategy{name: "ui", bot: GuesserFunc(UserGuess), interactive: true},
	Strategy{name: "pathetic", bot: GuesserFunc(HopelessGuesser), interactive: false},
	Strategy{name: "amazing", bot: GuesserFunc(AmazingGuesser), interactive: false},
}

// Command line flags
var corpusPath = flag.String("c", "", "required: `corpus-file` to load")
var nGames = flag.Int("n", 0, "the `number` of games to run, default entire corpus")
var verbose = flag.Bool("v", false, "enable verbose output")
var strategyNames = flag.String("s", "ui", "comma-separated list of `strategy-names` or ALL for all noninteractive strategies")
var goals = flag.String("g", "", "list of `goal-words`, default entire corpus")

// This is used to size the slice that holds the distribution of results for each
// bot, so enormous numbers are not advisable. It will work fine, but the output
// will be ridiculously hard to read if there are stupid bots that make many guesses.
const MAX_TRIES = 20

func main() {
	flag.Parse()

	if *corpusPath == "" {
		flag.PrintDefaults()
		return
	}
	corpus, err := gtw.LoadFile(*corpusPath)
	if err != nil {
		fmt.Printf("Cannot load corpus: %s\n", err)
		return
	}

	var selectedStrategies []Strategy
	if *strategyNames == "ALL" {
		for _, s := range(registeredStrategies) {
			if !s.interactive {
				selectedStrategies = append(selectedStrategies, s)
			}
		}
	} else {
		selectedNames := strings.Split(*strategyNames, ",")
		for _, s := range(registeredStrategies) {
			if stringInSlice(s.name, selectedNames) {
				selectedStrategies = append(selectedStrategies, s)
			}
		}
	}
	if len(selectedStrategies) == 0 {
		fmt.Printf("No strategies (bots) selected by the command line options\n")
		return
	}

	var goalWords []string
	if *goals == "" {
		goalWords = corpus 
	} else {
		goalWords, err = gtw.LoadFile(*goals)
		if err != nil {
			fmt.Printf("Cannot load goal words from %s\n", *goals)
			return
		}
	}

	games := *nGames
	if games == 0 || games >= len(goalWords) {
		games = len(goalWords)
	}
	if *verbose {
		fmt.Printf("Running %d games\n", games)
	}

	runAllSelectedBotsNGames(gtw.New(corpus), games, selectedStrategies, goalWords)
}

func runAllSelectedBotsNGames(engine *gtw.GtwEngine, games int, selectedStrategies []Strategy, goalWords []string) {
	statistics := make(map[string][]int)

	for _, s := range selectedStrategies {
		statistics[s.name] = make([]int, MAX_TRIES, MAX_TRIES)
	}
	
	for i := 0; i < games; i++ {
		engine.NewFixedGame(goalWords[i])
		goal := engine.Cheat()
		fmt.Printf("cheat: \"%s\"\n", goal)

		for _, s := range selectedStrategies {
			var guessResults []string
			nCorrect := 0
			var signature string

			for tries := 1; ; tries++ {
				guess := s.bot.Guess(engine.Corpus(), guessResults, nCorrect)
				signature, nCorrect = engine.Score(guess)
				if nCorrect == 5 {
					if *verbose {
						fmt.Printf("PASS: bot \"%s\" goal %s n %d\n", s.name, goal, tries)
					}
					statistics[s.name][tries]++
					break
				}
				guessResults = append(guessResults, signature)	
				if tries >= MAX_TRIES {
					if *verbose {
						fmt.Printf("FAIL: bot \"%s\" goal %s n %d\n", s.name, goal, tries)
					}
					statistics[s.name][MAX_TRIES-1]++
					break
				}
			}
		}
	}
	for name := range statistics {
		fmt.Printf("STATS bot %s : %v\n", name, statistics[name])
	}
}

func stringInSlice(s string, slice []string) bool {
	for _, in := range slice {
		if s == in {
			return true
		}
	}
	return false
}

// --- For test purposes - won't leave permanently ---
func HopelessGuesser(corpus []string, results []string, nCorrect int) string {
	return "xvqzw"
}

var amazingGuesserMagic int
func AmazingGuesser(corpus []string, results []string, nCorrect int) string {
	result := corpus[amazingGuesserMagic]
	amazingGuesserMagic++
	return result
}
// --- End "for test purposes" ---

