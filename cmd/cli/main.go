/*
Package cli implements a command line interface to play a word game.

*/
package main

import (
	"flag"
	"fmt"

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

var strategies = []Strategy {
	Strategy{name: "ui", bot: GuesserFunc(UserGuess), interactive: true,},
}

var corpusPath = flag.String("c", "", "specify the corpus")
var nGames = flag.Int("n", 0, "the number of games")
var verbose = flag.Bool("v", false, "enable verbose output")

const MAX_TRIES = 10 // don't allow a bot to try forever

func main() {
	flag.Parse()

	corpus, err := gtw.LoadFile(*corpusPath)
	if err != nil {
		fmt.Printf("Cannot load corpus %s\n", corpusPath)
		return
	}

	engine := gtw.New(corpus)
	if *nGames == 0 {
		*nGames = len(engine.Corpus())
		fmt.Printf("Running corpus (%d words) in order\n", *nGames)
	}
	runAllSelectedBotsNGames(engine)

}

func runAllSelectedBotsNGames(engine *gtw.GtwEngine) {
	for i := 0; i < *nGames; i++ {
		engine.NewFixedGame(engine.Corpus()[i])
		goal := engine.Cheat()
		fmt.Printf("goal %s\n", goal)

		for _, s := range strategies {
			scores := []string{}
			nCorrect := 0
			for tries := 1; ; tries++ {
				guess := s.bot.Guess(engine.Corpus(), scores, nCorrect)
				score, nCorrectx := engine.Score(guess)
				if nCorrectx == 5 {
					if *verbose {
						fmt.Printf("success bot %s goal %s n %d\n", s.name, goal, tries)
					}
					break
				}
				scores = append(scores, score)	
				if tries >= MAX_TRIES {
					fmt.Printf("fail bot %s goal %s n %d\n", s.name, goal, tries)
					break
				}
			}
		}
	}
}

