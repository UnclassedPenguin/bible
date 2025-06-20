package functions

import (
	"fmt"
	"strings"
	"golang.org/x/term"
)


// Function to ask the user for input in interactive mode
func GetUserInput(prompt string) []string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	bookChapterVerse, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return []string{}
	}

	// Clean the white space (including the newline character)
	userInput := strings.TrimSpace(bookChapterVerse)

	return strings.Split(userInput, " ")
}


// This Returns the width of the terminal (used for wordwrap)
func termWidth() int {
	termWidth, _, err := term.GetSize(0)
	if err != nil {
		fmt.Println("Error geting terminal width: ", err)
		return 0
  	}

	// Return with -1 so that it always has a gap of at least one spot on the right side. Just better readability.
	return termWidth - 1
}


// Wraps the text so that it doesn't split a word in the middle
func WordWrap(str string) {
	lineWidth := termWidth()
	words := strings.Fields(str)
	if len(words) == 0 {
		fmt.Println(str)
	}

	wrapped := words[0]
	spaceLeft := lineWidth - len(wrapped)

	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + word
			spaceLeft = lineWidth - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}

	fmt.Println(wrapped)
}


