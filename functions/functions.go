package functions

import (
	"os"
	"fmt"
	"log"
	"time"
	"bufio"
	"strings"
	"strconv"
	"math/rand"
	"database/sql"
	"golang.org/x/term"
)

// This struct is to hold the command line argurments
type Passage struct {
	BookName	string
	Chapter  	string
	Verse		string
}


var allBooks = []string{
    "Genesis", "Exodus", "Leviticus", "Numbers", "Deuteronomy",
    "Joshua", "Judges", "Ruth", "1 Samuel", "2 Samuel",
    "1 Kings", "2 Kings", "1 Chronicles", "2 Chronicles", "Ezra",
    "Nehemiah", "Esther", "Job", "Psalms", "Proverbs",
    "Ecclesiastes", "Song of Solomon", "Isaiah", "Jeremiah", "Lamentations",
    "Ezekiel", "Daniel", "Hosea", "Joel", "Amos",
    "Obadiah", "Jonah", "Micah", "Nahum", "Habakkuk",
    "Zephaniah", "Haggai", "Zechariah", "Malachi", "Matthew",
    "Mark", "Luke", "John", "Acts", "Romans",
    "1 Corinthians", "2 Corinthians", "Galatians", "Ephesians", "Philippians",
    "Colossians", "1 Thessalonians", "2 Thessalonians", "1 Timothy", "2 Timothy",
    "Titus", "Philemon", "Hebrews", "James", "1 Peter",
    "2 Peter", "1 John", "2 John", "3 John", "Jude",
    "Revelation",
}


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


// This returns a random book, chapter and verse in a string array
func RandomVerse(db *sql.DB) Passage {
	var passage Passage

	rand.Seed(time.Now().UnixNano())

	// Get a random book
	passage.BookName = allBooks[rand.Intn(66)]

	// Get number of chapters in random book
	chapters := GetAllChaptersInBook(db, passage.BookName)

	// Get random chapter
	// i think this needs the +1 to not try to pick chapter "0". needs to be 1-chapters It might be a bug?
	passage.Chapter = strconv.Itoa(rand.Intn(chapters) + 1)

	// Get number of verses in chapter
	verses := GetAllVersesInChapter(db, passage.BookName, passage.Chapter)

	// Get random verse
	// I think this needs the plus 1 to not try to get verse "0". it might be a bug?
	passage.Verse = strconv.Itoa(rand.Intn(verses) + 1)

	return passage
}


// This gives the number of verses in a chapter
func GetAllVersesInChapter(db *sql.DB, bookName string, chapter string) int {
    var verses []int
    query := "SELECT verse FROM bible WHERE bookName = ? AND chapter = ?"
    rows, err := db.Query(query, bookName, chapter)
    if err != nil {
        log.Fatal(err)
    }

    defer rows.Close()

    for rows.Next() {
        var verse int
        if err := rows.Scan(&verse); err != nil {
            log.Fatal(err)
        }
        verses = append(verses, verse)
    }

    return len(verses)
}


// This gives number of chapters in a book
func GetAllChaptersInBook(db *sql.DB, bookName string) int {
	query := "SELECT chapter FROM bible WHERE bookName = ?"
	rows, err := db.Query(query, bookName)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	
	uniqueMap := make(map[int]struct{})
	var uniqueChapters []int
	
	for rows.Next() {
		var chapter int

		if err := rows.Scan(&chapter); err != nil {
			log.Fatal(err)
		}

		if _, exists := uniqueMap[chapter]; !exists {
			uniqueMap [chapter] = struct{}{}
			uniqueChapters = append(uniqueChapters, chapter)
		}
	}

	return len(uniqueChapters)
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


