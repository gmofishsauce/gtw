package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gmofishsauce/gtw/lib"
)

var console *bufio.Reader
var previousGuess string

func UserGuess(corpus []string, scores []string, nCorrect int) string {
	if console == nil {
		console = bufio.NewReader(os.Stdin)
	}

	if len(scores) == 0 { // new game
		fmt.Println("New goal word selected")
	} else {
		// Not a new game - report the results of the user's previous guess
		score := scores[len(scores) - 1]
		fmt.Printf("       %s (%d letters in the correct place)\n", gtw.Humanize(score, previousGuess), nCorrect)
	}

	for { // loop over illegal guesses
		fmt.Printf("guess> ")
		text, _ := console.ReadString('\n')
		text = strings.TrimSpace(text)
		if len(text) == 5 {
			previousGuess = text
			return previousGuess
		}
		fmt.Println("5-letter words only")
	}
}
