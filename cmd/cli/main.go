/*
Package cli implements a command line interface to play a word game.

*/
package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/gmofishsauce/gtw/lib"
)

type Guesser interface{
	Guess(corpus []string, scores []string, nCorrect int) string
}

// Allow the Guesser interface to be implemented with just functions
// See the "ui" strategy in the strategies array below for an example
type GuesserFunc func([]string, []string, int) string

func (f GuesserFunc) Guess(c []string, s []string, n int) string {
	return f(c, s, n)
}

type Strategy struct {
	name string
	bot  Guesser 
	interactive bool
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

// Add your strategy here
var registeredStrategies = []Strategy {
	Strategy{name: "ui", bot: GuesserFunc(UserGuess), interactive: true,},
	Strategy{name: "pathetic", bot: GuesserFunc(HopelessGuesser), interactive: false},
	Strategy{name: "amazing", bot: GuesserFunc(AmazingGuesser), interactive: false},
}

// Command line flags
var corpusPath = flag.String("c", "", "specify the corpus")
var nGames = flag.Int("n", 0, "the number of games, default entire corpus")
var verbose = flag.Bool("v", false, "enable verbose output")
var rawStrategies = flag.String("s", "ui", "comma-separate list of strategies, default \"ui\".")
var goals = flag.String("g", "", "specify a set of goal words")

const MAX_TRIES = 10 // don't allow a bot to try forever (10 is too low though)

func main() {
	flag.Parse()

	corpus, err := gtw.LoadFile(*corpusPath)
	if err != nil {
		fmt.Printf("Cannot load corpus %s\n", corpusPath)
		return
	}

	// XXX this is silly, just a pass a []Strategy to the runner
	var selectedStrategies []string
	if (*rawStrategies == "ALL") {
		selectedStrategies = []string{}
		for _, s := range(registeredStrategies) {
			if !s.interactive {
				selectedStrategies = append(selectedStrategies, s.name)
			}
		}
	} else {
		selectedStrategies = strings.Split(*rawStrategies, ",")
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
	fmt.Printf("Running %d games\n", games)

	runAllSelectedBotsNGames(gtw.New(corpus), games, selectedStrategies, goalWords)
}

func runAllSelectedBotsNGames(engine *gtw.GtwEngine, games int, selectedStrategies []string, goalWords []string) {
	for i := 0; i < games; i++ {
		engine.NewFixedGame(goalWords[i])
		goal := engine.Cheat()
		fmt.Printf("cheat: \"%s\"\n", goal)

		for _, s := range registeredStrategies {
			if ! stringInSlice(s.name, selectedStrategies) {
				continue
			}
			var guessResults []string
			nCorrect := 0
			for tries := 1; ; tries++ {
				guess := s.bot.Guess(engine.Corpus(), guessResults, nCorrect)
				result, nCorrect := engine.Score(guess)
				if nCorrect == 5 {
					if *verbose {
						fmt.Printf("PASS: bot \"%s\" goal %s n %d\n", s.name, goal, tries)
					}
					break
				}
				guessResults = append(guessResults, result)	
				if tries >= MAX_TRIES {
					fmt.Printf("FAIL: bot \"%s\" goal %s n %d\n", s.name, goal, tries)
					break
				}
			}
		}
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

