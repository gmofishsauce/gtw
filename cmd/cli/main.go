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

var registeredStrategies = []Strategy {
	Strategy{name: "ui", bot: GuesserFunc(UserGuess), interactive: true,},
	Strategy{name: "pathetic", bot: GuesserFunc(HopelessGuesser), interactive: false},
	Strategy{name: "amazing", bot: GuesserFunc(AmazingGuesser), interactive: false},
}
var NonInteractiveStrategyNames = []string{"pathetic","amazing"} // TODO construct dynamically

var corpusPath = flag.String("c", "", "specify the corpus")
var nGames = flag.Int("n", 0, "the number of games, default entire corpus")
var verbose = flag.Bool("v", false, "enable verbose output")
var rawStrategies = flag.String("s", "ui", "comma-separate list of strategies, default \"ui\".")

const MAX_TRIES = 10 // don't allow a bot to try forever (10 is too low though)

func main() {
	flag.Parse()

	corpus, err := gtw.LoadFile(*corpusPath)
	if err != nil {
		fmt.Printf("Cannot load corpus %s\n", corpusPath)
		return
	}

	var strategyNames []string
	if (*rawStrategies == "ALL") {
		strategyNames = NonInteractiveStrategyNames
	} else {
		strategyNames = strings.Split(*rawStrategies, ",")
	}

	engine := gtw.New(corpus)
	if *nGames == 0 {
		*nGames = len(engine.Corpus())
		fmt.Printf("Running corpus (%d words) in order\n", *nGames)
	} else if *nGames > len(corpus) {
		*nGames = len(corpus)
	}
	runAllSelectedBotsNGames(engine, strategyNames)

}

func stringInSlice(s string, slice []string) bool {
	for _, in := range slice {
		if s == in {
			return true
		}
	}
	return false
}

func runAllSelectedBotsNGames(engine *gtw.GtwEngine, strategyNames []string) {
	for i := 0; i < *nGames; i++ {
		engine.NewFixedGame(engine.Corpus()[i])
		goal := engine.Cheat()
		fmt.Printf("goal \"%s\"\n", goal)

		for _, s := range registeredStrategies {
			if ! stringInSlice(s.name, strategyNames) {
				continue
			}
			scores := []string{}
			nCorrect := 0
			for tries := 1; ; tries++ {
				guess := s.bot.Guess(engine.Corpus(), scores, nCorrect)
				score, nCorrectx := engine.Score(guess)
				if nCorrectx == 5 {
					if *verbose {
						fmt.Printf("PASS: bot \"%s\" goal %s n %d\n", s.name, goal, tries)
					}
					break
				}
				scores = append(scores, score)	
				if tries >= MAX_TRIES {
					fmt.Printf("FAIL: bot \"%s\" goal %s n %d\n", s.name, goal, tries)
					break
				}
			}
		}
	}
}

