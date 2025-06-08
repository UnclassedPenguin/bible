//------------------------------------------------------------------------------
// Bible program
// Written by UnclassedPenguin
// https://github.com/unclassedpenguin
//------------------------------------------------------------------------------

package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"math/rand"
	"time"
	"golang.org/x/term"
	_ "embed"
	_ "github.com/mattn/go-sqlite3"
)

// This struct is to reference the sql database
type Bible struct {
	ID       int
	BookName string
	Book		 int
	Chapter  int
	Verse    int
	Text     string
}

// This struct is to hold the command line argurments
type Args struct {
	BookName	string
	Chapters  string
	Verses		string
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

// Embed the sql database into the binary
//go:embed kjv.db
var embeddedDb []byte


// The main function :p (The more comments the better!)
func main() {
	// Version number
	version := "v0.1.1"

	// Create a temporary file to hold the embedded database
	tmpFile, err := os.CreateTemp("", "kjv.db")
	if err != nil {
			log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temp file afterwards

	// Write the embedded database to the temporary file
	if _, err := tmpFile.Write(embeddedDb); err != nil {
			log.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
			log.Fatal(err)
	}

	// Define the -i flag for interactive mode
	interactive := flag.Bool("i", false, "Enable interactive mode")
	// Define -l flag for the list mode
	listMode := flag.Bool("l", false, "List Info")
	// Define -v flag for the version mode
	versionMode := flag.Bool("v", false, "Print Version")
	randomMode := flag.Bool("r", false, "Print random verse")
	searchMode := flag.Bool("s", false, "search for term")
	exactMode := flag.Bool("e", false, "search for exact term, use with -s")
	testMode := flag.Bool("t", false, "Test function, for testing.")
	flag.Parse()

	//db, err := sql.Open("sqlite3", "./kjv.db")
	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
			log.Fatal(err)
	}
	defer db.Close()

	// These are all the different "modes"
	if *interactive {
		interactiveMode(db)
	} else if *listMode {
		infoMode(db)
	} else if *versionMode {
		fmt.Println(version)
	} else if *randomMode {
		printRandomVerse(db)
	} else if *searchMode {
		searchForTerm(db, *exactMode)
	} else if *testMode {
		testFunction(db)
	} else {
		singleShotMode(db)
	}
}


// This is the main interactive mode that opens up a "command line" that you can interact with and change verses.
func interactiveMode(db *sql.DB) {
	var bookName string
	var chapter string
	var verse string
	var id int
	
	// Loop to get initial input from user. 
	for {
		// Get user input 
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter Book Chapter Verse(ie Genesis 1 1): ")
		bookChapterVerse, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		// Clean the white space (including the newline character)
		userInput := strings.TrimSpace(bookChapterVerse)

		// Split the user input into its parts (should be book chapter verse or 'r' for random
		bookChapterVerseSplit := strings.Split(userInput, " ")

		// Check if it was 'r' for random, and if so, get id of random verse to start at
		if len(bookChapterVerseSplit) == 1 && bookChapterVerseSplit[0] == "r" {
                  rVerse := randomVerse(db)
                  id = getIdOfVerse(db, rVerse[0], rVerse[1], rVerse[2])
                  break
		// If any other single character, prompt proper usage
		} else if len(bookChapterVerseSplit) == 1 {
                  fmt.Println("Please enter either a book chapter verse(ie Genesis 1 1) or 'r' for random verse")
		// If Specific book chapter verse to start at, get the id
		} else {
                  // Detect if it is a numbered book ("1 John" etc)
                  if bookChapterVerseSplit[0][0] == '"' {
                    bookName = strings.Trim(bookChapterVerseSplit[0] + " " + bookChapterVerseSplit[1], "\"")
                    chapter = bookChapterVerseSplit[2]
                    verse = bookChapterVerseSplit[3]
                    id = getIdOfVerse(db, bookName, chapter, verse)
                    break
                  } else if  bookChapterVerseSplit[0][0] == '\'' {
                    bookName = strings.Trim(bookChapterVerseSplit[0] + " " + bookChapterVerseSplit[1], "'")
                    chapter = bookChapterVerseSplit[2]
                    verse = bookChapterVerseSplit[3]
                    id = getIdOfVerse(db, bookName, chapter, verse)
                    break
                  // If not a numbered book, get the verse
                  } else {
                    bookName = bookChapterVerseSplit[0]
                    chapter = bookChapterVerseSplit[1]
                    verse = bookChapterVerseSplit[2]
                    id = getIdOfVerse(db, bookName, chapter, verse)
                    break
                  }
		}
	}

	// Diagnostics
	//fmt.Println("split: ", bookChapterVerseSplit)
	//fmt.Printf("split type: %T\n", bookChapterVerseSplit)
	//fmt.Println("split length: ", len(bookChapterVerseSplit))

	//fmt.Println("bookName: ", bookName)
	//fmt.Printf("bookName Type: %T\n", bookName)
	
	//fmt.Println("chapter: ", chapter)
	//fmt.Printf("chapter Type: %T\n", chapter)

	//fmt.Println("startVerse: ", startVerse)
	//fmt.Printf("startVerse Type: %T\n", startVerse)

	//fmt.Println("currentVerse: ", currentVerse)
	//fmt.Printf("currentVerse Type: %T\n", currentVerse)

	// Print info for usage for user 1 time at beginning
	fmt.Print("\nPress 'n' for next verse, 'p' for prev, 'r' for random, or 'q' to quit: \n\n")

	// This is the main loop of interactive mode. Prints out the verse based on the id number
	for {
		fmt.Printf("\n")
		var bibleVerse Bible
		err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible WHERE id = ?", id).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
		if err != nil {
			fmt.Printf("Verse %s %d:%d not found.\n", bookName, chapter, verse)
			break
		}

		// This actually prints the verse
		fmt.Printf("%s %d:%d\n", bibleVerse.BookName, bibleVerse.Chapter, bibleVerse.Verse)
		wordWrap(bibleVerse.Text)
		

		// Prompt for next command
		reader2 := bufio.NewReader(os.Stdin)
		fmt.Print(": ")
		input, err := reader2.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		// Clean the white space (including the newline character)
		userInput := strings.TrimSpace(input)

		// Split the user input into its parts (should be book chapter verse or 'r' for random
		inputSplit := strings.Split(userInput, " ")


		// If length of inputSplit is 3 it should mean the user has input book chapter verse (ie James 2 24)
		// Which means it should jump to that verse
		if len(inputSplit) == 3 {
			id = getIdOfVerse(db, inputSplit[0], inputSplit[1], inputSplit[2])
		// If length of inputSplit is 1, it is probably just a command 
		} else if len(inputSplit) == 1 {
			switch strings.ToLower(inputSplit[0]) {
			// Go to next verse
			case "n":
				id++
			// Go to prev verse
			case "p":
				if id > 1 {
					id--
				} else {
					fmt.Println("You are at the first verse.")
				}
			//Get a random verse
			case "r":
				rVerse := randomVerse(db)
				//Get id of random verse
				id = getIdOfVerse(db, rVerse[0], rVerse[1], rVerse[2])
			// quit :p
			case "q":
				//fmt.Println("Exiting interactive mode.")
				return
			case "x":
				return
			default:
				fmt.Println("Invalid input. Please enter 'n', 'p', 'r' or 'q'.")
			}
		}

		// Clear the console for better readability
		clearConsole()
	}
}


// This is just to give info. If no other arguments, list all books. If only book, give number of chapters. If book and chapter, give number of verses.
func infoMode(db *sql.DB) {
	// Print all books
	if len(os.Args) == 2 {
		for i := 0; i < len(allBooks); i++ {
			// This is just for formatting. No comma and newline on last one
			if i == len(allBooks)-1 {
				fmt.Printf("%s\n", allBooks[i])
			} else {
				fmt.Printf("%s, ", allBooks[i])
			}
		}
	}

	// If just a book is provided, print Number of chapters
	if len(os.Args) == 3 {
		var args Args
		args.BookName = os.Args[2]

		chapters := getAllChaptersInBook(db, args.BookName)
		fmt.Printf("Chapters in %s: %d\n", args.BookName, chapters)
	}

	// if a book and a chapter, print number of verses
	if len(os.Args) == 4 {
		var args Args
		args.BookName = os.Args[2]
		args.Chapters = os.Args[3]
		verses := getAllVersesInChapter(db, args.BookName, args.Chapters)
		fmt.Printf("Verses in %s %s: %d\n", args.BookName, args.Chapters, verses)
	}
}


// Fucntion to print a random verse. use -r on command line
func printRandomVerse(db *sql.DB) {
	// Get random verse
	random := randomVerse(db)
	// Print random verse
	printVerse(db, random[0], random[1], random[2])
}


// This returns a random book, chapter and verse in a string array
func randomVerse(db *sql.DB) []string {
	rand.Seed(time.Now().UnixNano())

	// Get a random book
	randomBook := allBooks[rand.Intn(66)]

	// Get number of chapters in random book
	chapters := getAllChaptersInBook(db, randomBook)

	// Get random chapter
	// i think this needs the +1 to not try to pick chapter "0". needs to be 1-chapters It might be a bug?
	randomChapter := rand.Intn(chapters) + 1

	// Get number of verses in chapter
	verses := getAllVersesInChapter(db, randomBook, strconv.Itoa(randomChapter))

	// Get random verse
	// I think this needs the plus 1 to not try to get verse "0". it might be a bug?
	randomVerse := rand.Intn(verses) + 1

	return []string{randomBook, strconv.Itoa(randomChapter), strconv.Itoa(randomVerse)}
}


// Print Verse
// these are a string for a reason...I think beause random verse needs to return a []string, so it made it easier to do? because the bookname is a string,
// And I wanted it to return a single array
func printVerse(db *sql.DB, book string, chapter string, verse string) {
	
	chapterInt, _ := strconv.Atoi(chapter)
	verseInt, _ := strconv.Atoi(verse)

	var bibleVerse Bible
	err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible where bookName = ? AND chapter = ? AND verse = ?", book, chapterInt, verseInt).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
	if err != nil {
		fmt.Printf("Verse %s %s:%s not found\n", book, chapter, verse)
	}
	fmt.Printf("%s %s:%s\n", book, chapter, verse)
	wordWrap(bibleVerse.Text)
	fmt.Printf("\n")
}


// Search for a term or an exact term
func searchForTerm(db *sql.DB, exactMode bool) {
	// This executes an exact search for the search term
	if exactMode {
		query := "select bookName, chapter, verse, text from bible where text like ?"
		rows, err := db.Query(query, "% "+os.Args[3]+" %")
		if err != nil {
			fmt.Println("Error in query of exact search: ", err)
			os.Exit(1)
		}
    defer rows.Close()

		if !rows.Next() {
			fmt.Println("No search found matching: ", os.Args[3])
			os.Exit(0)
		}

		for {
			var bible Bible
			err = rows.Scan(&bible.BookName, &bible.Chapter, &bible.Verse, &bible.Text)
			if err != nil {
				fmt.Println("Error reading row(exact search): ", err)
				os.Exit(1)
			}

			fmt.Printf("%s %d:%d\n", bible.BookName, bible.Chapter, bible.Verse)
			wordWrap(bible.Text)
			fmt.Printf("\n")

			if !rows.Next() {
				break
			}
		}

	// This executes a search that takes any match, ie love with match with loved
	} else {
		query := "select bookName, chapter, verse, text from bible where text like ?"
		// This is the only thing that is different from exact search. No spaces around the search term
		rows, err := db.Query(query, "%"+os.Args[2]+"%")
		if err != nil {
			fmt.Println("Error in query of search: ", err)
			os.Exit(1)
		}
    defer rows.Close()

		if !rows.Next() {
			fmt.Println("No search found matching: ", os.Args[2])
			os.Exit(0)
		}

		for {
			var bible Bible
			err = rows.Scan(&bible.BookName, &bible.Chapter, &bible.Verse, &bible.Text)
			if err != nil {
				fmt.Println("Error reading row(search): ", err)
				os.Exit(1)
			}

			fmt.Printf("%s %d:%d\n", bible.BookName, bible.Chapter, bible.Verse)
			wordWrap(bible.Text)
			fmt.Printf("\n")

			if !rows.Next() {
				break
			}
		}
	}
}


// This runs if no "flags" are provided, but there may be arguments. 
func singleShotMode(db *sql.DB) {

	// if no argurments provided, print all books
	if len(os.Args) == 1 {
		for i := 0; i < len(allBooks); i++ {
			// This is just for formatting. No comma and newline on last one
			if i == len(allBooks)-1 {
				fmt.Printf("%s\n", allBooks[i])
			} else {
				fmt.Printf("%s, ", allBooks[i])
			}
		}
		os.Exit(0)
	}

	var args Args
	args.BookName = os.Args[1]

	// If just a book is provided, Print number of chapters.
	if len(os.Args) == 2 {
		chapters := getAllChaptersInBook(db, args.BookName)
		fmt.Printf("Chapters in %s: %d\n", args.BookName, chapters)
		fmt.Println()
	}

	// if a book and a chapter, print the entire chapter
	if len(os.Args) == 3 {
		args.Chapters = os.Args[2]
		printChapters(db, args)
	}

	// if book and chapter and verse(s), print the verse(s)
	if len(os.Args) == 4 {
		args.Chapters = os.Args[2]
		args.Verses = os.Args[3]
		printVerses(db, args)
	}
}


// This function runs if you provide only a book
// I don't really want to use this. I don't want to print entire books...
func printBook() {

}


// This function runs if you provide 2 arguments, ie a book and a chapter or range of chapters
func printChapters(db *sql.DB, args Args) {
	// This is for a range of chapters ie "bible "1 Cornthians" 1-3"
	if strings.Contains(args.Chapters, "-") {
		chapters, err := getIntsStartAndEnd(args.Chapters)
		if err != nil {
			fmt.Println("Error getting all chapters: ", err)
		}

		// For every chapter
		for i := 0 ; i < len(chapters); i++{
			fmt.Println("Chapter ", chapters[i])
			// We need to get the number of verses for the chapter
			verses := getAllVersesInChapter(db, args.BookName, strconv.Itoa(chapters[i]))

			// For every verse
			for j := 1; j <= verses; j++ {
				printVerse(db, args.BookName, strconv.Itoa(chapters[i]), strconv.Itoa(j))
			}
		}

	// This is for a single chapter ie "bible "1 Corinthians" 1"
	} else {

		verses := getAllVersesInChapter(db, args.BookName, args.Chapters)

		for i := 1; i <= verses; i++ {
			printVerse(db, args.BookName, args.Chapters, strconv.Itoa(i))
		}
	}
}


// This function runs if you provide all 3 arguments. Ie  a book, a chapter, and a verse or range of verses.
func printVerses(db *sql.DB, args Args) {
	// This is for a range of verses ie "bible "1 Corinthians" 1 1-5"
	if strings.Contains(args.Verses, "-") {
		verses, err := getIntsStartAndEnd(args.Verses)
		if err != nil {
			fmt.Println("Error gettings all verses: ", err)
		}
		for i := 0; i < len(verses); i ++ {
			printVerse(db, args.BookName, args.Chapters, strconv.Itoa(verses[i]))
		}
	// This is for a single verse
	} else {
		printVerse(db, args.BookName, args.Chapters, args.Verses)
	}
}


// Get the id of a verse. Would be useful in interactive mode, so that then you could just go next or previous based on id.
func getIdOfVerse(db *sql.DB, bookName string, chapter string, verse string) int {
	var id int
	query := "SELECT id FROM bible WHERE bookName = ? AND chapter = ? AND verse = ?"
	err := db.QueryRow(query, bookName, chapter, verse).Scan(&id)
	if err != nil {
		fmt.Printf("Can't get id of %s %s %s: ", bookName, chapter, verse)
		fmt.Println(err)
	}
	
	return id
}


// This gives number of chapters in a book
func getAllChaptersInBook(db *sql.DB, bookName string) int {
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

// This gives the number of verses in a chapter
func getAllVersesInChapter(db *sql.DB, bookName string, chapter string) int {
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


// What is this? I don't use it...
// REMOVE
func printAllVerses(db *sql.DB, bookName string, chapter int, verse int) {
	fmt.Println("printAllVerses func")
}


// getIntsStartAndEnd parses a string in the format "start-end" and returns a slice of integers.
func getIntsStartAndEnd(s string) ([]int, error) {
	split := strings.Split(s, "-")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid format, expected 'start-end'")
	}

	start, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, fmt.Errorf("error converting start to int: %v", err)
	}

	end, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, fmt.Errorf("error converting end to int: %v", err)
	}

	// Create a slice to hold the integers from start to end
	var ints []int
	for i := start; i <= end; i++ {
		ints = append(ints, i)
	}

	return ints, nil
}


// it doesn't appear I ever use this?
// REMOVE
func parseVerseInput(input string) []int {
    var verses []int
    if strings.Contains(input, "-") {
        parts := strings.Split(input, "-")
        start, _ := strconv.Atoi(parts[0])
        end, _ := strconv.Atoi(parts[1])
        for i := start; i <= end; i++ {
            verses = append(verses, i)
        }
    } else {
        verse, _ := strconv.Atoi(input)
        verses = append(verses, verse)
    }
    return verses
}


// This doesn't even work...
func clearConsole() {
	cmd := exec.Command("clear") // For Unix/Linux
	if err := cmd.Run(); err != nil {
		// If clearing fails, try for Windows
		cmd = exec.Command("cmd", "/c", "cls")
		if err := cmd.Run(); err != nil {
			fmt.Println("Unable to clear console:", err)
		}
	}
}


// This Returns the width of the terminal (used for wordwrap)
func termWidth() int {
	termWidth, _, err := term.GetSize(0)
  if err != nil {
		fmt.Println("Error geting terminal width: ", err)
    os.Exit(1)
  }

	return termWidth
}


// Wraps the text so that it doesn't split a word in the middle
func wordWrap(str string) {
	lineWidth := termWidth()
	//words := strings.Fields(strings.TrimSpace(text))
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


// This is just for testing random things...
func testFunction(db *sql.DB) {
	var book string
	var chapter string
	var verse string
	fmt.Println("Book: ")
	fmt.Scan(&book)
	fmt.Println("Chatper: ")
	fmt.Scan(&chapter)
	fmt.Println("Verse: ")
	fmt.Scan(&verse)

	id := getIdOfVerse(db, book, chapter, verse)
	fmt.Printf("ID: %d\n", id)
}
