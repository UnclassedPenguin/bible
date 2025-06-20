package functions

import (
	"os"
	"fmt"
	"bufio"
	"strings"
	"database/sql"
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


//  Takes the command from interactive mode (if it is more than a single character), returns the id of the verse to go to
func ParseInteractiveCommand(db *sql.DB, split []string) int {
	var bookName string
	var chapter string
	var verse string
	var id int

	// This is a special case for "Song of Solomon", which is the only 3 word book
	if split[0][0] == '"' && split[0] == "\"Song" {
		bookName = strings.Trim(split[0] + " " + split[1] + " " + split[2], "\"")
		chapter = split[3]
		verse = split[4]
		id = GetIdOfVerse(db, bookName, chapter, verse)
		return id
	// This is a special case for 'Song of Solomon' (if single quotes are used)
	} else if split[0][0] == '\'' && split[0] == "'Song" {
		bookName = strings.Trim(split[0] + " " + split[1] + " " + split[2], "'")
		chapter = split[3]
		verse = split[4]
		id = GetIdOfVerse(db, bookName, chapter, verse)
		return id
	// Detect if it is a numbered book ("1 John" etc)
	} else if split[0][0] == '"' {
		bookName = strings.Trim(split[0] + " " + split[1], "\"")
		chapter = split[2]
		verse = split[3]
		id = GetIdOfVerse(db, bookName, chapter, verse)
		return id
	// This is just to catch if the user uses single quotes instead of double
	} else if  split[0][0] == '\'' {
		bookName = strings.Trim(split[0] + " " + split[1], "'")
		chapter = split[2]
		verse = split[3]
		id = GetIdOfVerse(db, bookName, chapter, verse)
		return id
	// If not a numbered book, get the verse
	} else {
		bookName = split[0]
		chapter = split[1]
		verse = split[2]
		id = GetIdOfVerse(db, bookName, chapter, verse)
		return id
	}
	return -1
}


// Get the id of a verse. Would be useful in interactive mode, so that then you could just go next or previous based on id.
func GetIdOfVerse(db *sql.DB, bookName string, chapter string, verse string) int {
	var id int
	query := "SELECT id FROM bible WHERE bookName = ? AND chapter = ? AND verse = ?"
	err := db.QueryRow(query, bookName, chapter, verse).Scan(&id)
	if err != nil {
		fmt.Printf("Can't get id of %s %s %s: ", bookName, chapter, verse)
		fmt.Println(err)
	}
	
	return id
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


