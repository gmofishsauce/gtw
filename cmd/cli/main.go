/*
Package cli implements a command line interface to play a word game.

*/
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gmofishsauce/gtw/lib"
)

const defaultCorpusName = "webster-2-all-five-letter.corpus"

const help = `
--------
After each guess, a signature will be displayed. In the signature,
the character '+' means the guess character above it is correct and
in the correct location. '*' means the character above is in the
word, but not in the correct location, while '#' means the character
above is not in the word. Example:

guess> tears
       ##*** (0 letters in the correct place)
guess> cloud
       #*#*# (0 letters in the correct place)
guess> aural
       *++++ (4 letters in the correct place)
guess> rural

Success!
--------
`

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: cli corpus-path")
		return
	}
	corpusPath := os.Args[1]
	corpus, err := gtw.LoadFile(corpusPath)
	if err != nil {
		fmt.Printf("Cannot load corpus %s\n", corpusPath)
		return
	}
	fmt.Printf("Loaded corpus: %d words\n", len(corpus))
	fmt.Println(help)

	engine := gtw.New(corpus)

	reader := bufio.NewReader(os.Stdin)
	for { // one game per loop. Runs until ^C.
		engine.NewGame()
		fmt.Println("New goal word selected")
		for { // one guess per loop. Runs until success
			fmt.Printf("guess> ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if len(text) != 5 {
				fmt.Println("5-letter words only")
			} else {
				signature, score := engine.Score(text)
				if score == 5 {
					fmt.Println("\nSuccess!\n")
					break
				}
				fmt.Printf("       %s (%d letters in the correct place)\n", signature, score)
			}
		}
	}
}
