//------------------------------------------------------------------------------
// Bible program
// Written by UnclassedPenguin
// https://github.com/unclassedpenguin/bible.git
//------------------------------------------------------------------------------

package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	_ "embed"
	"os/exec"
	"strconv"
	"strings"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	f "bible/functions"
)

// This struct is to reference the sql database
type Bible struct {
	ID       	int
	BookName	string
	Book		int
	Chapter  	int
	Verse    	int
	Text     	string
}

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

// Embed the sql database into the binary
//go:embed kjv.db
var embeddedDb []byte


// The main function :p (The more comments the better!)
func main() {
	// Version number
	versionNumber := "v0.2.5"

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

	// Command line flags
	interactive := flag.Bool("i", false, "Enable interactive mode")
	list := flag.Bool("l", false, "List Info")
	version := flag.Bool("v", false, "Print Version")
	random := flag.Bool("r", false, "Print random verse")
	search := flag.Bool("s", false, "search for term")
	exact := flag.Bool("e", false, "search for exact term, use with -s")
	favorite := flag.Bool("f", false, "List favorite verses")
	//test := flag.Bool("t", false, "Test function, for testing.")

  	// This changes the help/usage info when -h is used.
	flag.Usage = func() {
		w := flag.CommandLine.Output() // may be os.Stderr - but not necessarily
		description := "%s\n\n" +
		"This program lets you read the bible in the command line.\n\n" +
		" Basic Usage:\n\n" +
		" \"bible Genesis 1 1\" or \"bible -i\"\n\n" +
		"Available arguments:\n"
		fmt.Fprintf(w, description, os.Args[0])
		flag.PrintDefaults()
		//fmt.Fprintf(w, "...custom postamble ... \n")
	}	

	flag.Parse()

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// These are all the different "modes"
	switch {
	case *interactive:
		interactiveMode(db)
	case *list:
		listMode(db)
	case *version:
		fmt.Println(versionNumber)
	case *random:
		printRandomVerse(db)
	case *search:
		searchForTerm(db, *exact)
	//case *test:
		//testFunction(db)
	case *favorite:
		favoriteMode(db)
	default:
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
		userInputSplit := f.GetUserInput("Enter Book Chapter Name(ie Genesis 1 1): ")

		// Check if it was 'r' for random, and if so, get id of random verse to start at
		if len(userInputSplit) == 1 && userInputSplit[0] == "r" {
			passage := f.RandomVerse(db)
			id = f.GetIdOfVerse(db, passage.BookName, passage.Chapter, passage.Verse)
			break
		// Load bookmark
		} else if len(userInputSplit) == 1 && userInputSplit[0] == "b" {
			id = f.LoadBookmark()
			break
		} else if len(userInputSplit) == 1 && userInputSplit[0] == "q" {
			return
		} else if len(userInputSplit) == 1 && userInputSplit[0] == "?" {
			f.PrintInteractiveHelp()
		} else if len(userInputSplit) == 1 && userInputSplit[0] == "h" {
			f.PrintInteractiveHelp()
		// If any other single character, prompt proper usage
		} else if len(userInputSplit) == 1 {
			f.WordWrap("Please enter either a book chapter verse(ie Genesis 1 1) or 'r' for random verse")
		// If Specific book chapter verse to start at, get the id
		} else {
			id = f.ParseInteractiveCommand(db, userInputSplit)
			// Check if not valid input ParseInteractiveCommand returns -1 on failure.
			if id == -1 {
				fmt.Println("Please enter a valid verse\n")
			} else {
				break
			}
		}
	}

	// Print info for usage for user 1 time at beginning
	f.WordWrap("\nPress 'n' for next verse, 'p' for prev, 'r' for random, '?' for help, or 'q' to quit: \n\n")

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
		f.WordWrap(bibleVerse.Text)
		
		// Prompt for next command
		inputSplit := f.GetUserInput(": ")

		if len(inputSplit) == 1 {
			switch strings.ToLower(inputSplit[0]) {
			case "n": // Go to next verse
					id++
			case "p": // Go to prev verse
				if id > 1 {
					id--
				} else {
					fmt.Println("You are at the first verse.")
				}
			case "b":
				id = f.BookMark(bibleVerse.ID)
			case "f":
				f.Favorites(db, bibleVerse.ID)
			case "?":
				f.PrintInteractiveHelp()
			case "h":
				f.PrintInteractiveHelp()
			case "r": // Get a random verse
				passage := f.RandomVerse(db)
				//Get id of random verse
				id = f.GetIdOfVerse(db, passage.BookName, passage.Chapter, passage.Verse)
			case "q": // quit :p
				return
			case "x":
				return
			default:
				fmt.Println("Invalid input. Please enter 'n', 'p', 'r' or 'q'.")
			}
		} else {
			// Capture current id incase of failure
			oldid := id
			id = f.ParseInteractiveCommand(db, inputSplit)
			// Check if failure. ParseInteractiveCommand returns -1 on failure, so prompt user, and set id back to the current verse id
			if id == -1 {
				fmt.Printf("Please enter a valid verse\n")
				id = oldid
			}
		}

		// Clear the console for better readability
		// This doesn't even work so I'm going to comment it out for now
		//clearConsole()
	}
}


// This is just to give info. If no other arguments, list all books. If only book, give number of chapters. If book and chapter, give number of verses.
func listMode(db *sql.DB) {
	// Print all books
	if len(os.Args) == 2 {
		var allBooksString string
		for i := 0; i < len(allBooks); i++ {
			// This is just for formatting. No comma and newline on last one
			if i == len(allBooks)-1 {
				//fmt.Printf("%s\n", allBooks[i])
				allBooksString += allBooks[i] + "\n"
			// If not last one, just append the book with a comma
			} else {
				//fmt.Printf("%s, ", allBooks[i])
				allBooksString += allBooks[i] + ", "
			}
		}

		f.WordWrap(allBooksString)	
	
	// If just a book is provided, print Number of chapters
	} else if len(os.Args) == 3 {
		var passage Passage
		passage.BookName = os.Args[2]

		chapters := f.GetAllChaptersInBook(db, passage.BookName)
		fmt.Printf("Chapters in %s: %d\n", passage.BookName, chapters)

	// if a book and a chapter, print number of verses
	} else if len(os.Args) == 4 {
		var passage Passage 
		passage.BookName = os.Args[2]
		passage.Chapter = os.Args[3]

		verses := f.GetAllVersesInChapter(db, passage.BookName, passage.Chapter)
		fmt.Printf("Verses in %s %s: %d\n", passage.BookName, passage.Chapter, verses)
	}
}


// Fucntion to print a random verse. use -r on command line
func printRandomVerse(db *sql.DB) {
	// Get random verse
	passage := f.RandomVerse(db)
	// Print random verse
	f.PrintVerse(db, passage.BookName, passage.Chapter, passage.Verse)
}



// Search for a term or an exact term
func searchForTerm(db *sql.DB, exact bool) {
	// This executes an exact search for the search term
	if exact {
		query := "SELECT bookName, chapter, verse, text FROM bible WHERE text LIKE ?"
		rows, err := db.Query(query, "% "+os.Args[3]+" %")
		if err != nil {
			fmt.Println("Error in query of exact search: ", err)
			return
		}

    	defer rows.Close()

		if !rows.Next() {
			fmt.Println("No search found matching: ", os.Args[3])
			return
		}

		for {
			var bible Bible
			err = rows.Scan(&bible.BookName, &bible.Chapter, &bible.Verse, &bible.Text)
			if err != nil {
				fmt.Println("Error reading row(exact search): ", err)
				return
			}

			fmt.Printf("%s %d:%d\n", bible.BookName, bible.Chapter, bible.Verse)
			f.WordWrap(bible.Text)
			fmt.Printf("\n")

			if !rows.Next() {
				break
			}
		}

	// This executes a search that takes any match, ie love with match with loved
	} else {
		query := "SELECT bookName, chapter, verse, text FROm bible WHERE text LIKE ?"
		// This is the only thing that is different from exact search. No spaces around the search term
		rows, err := db.Query(query, "%"+os.Args[2]+"%")
		if err != nil {
			fmt.Println("Error in query of search: ", err)
			return
		}

    	defer rows.Close()

		if !rows.Next() {
			fmt.Println("No search found matching: ", os.Args[2])
			return
		}

		for {
			var bible Bible
			err = rows.Scan(&bible.BookName, &bible.Chapter, &bible.Verse, &bible.Text)
			if err != nil {
				fmt.Println("Error reading row(search): ", err)
				return
			}

			fmt.Printf("%s %d:%d\n", bible.BookName, bible.Chapter, bible.Verse)
			f.WordWrap(bible.Text)
			fmt.Printf("\n")

			if !rows.Next() {
				break
			}
		}
	}
}

func favoriteMode(db *sql.DB) {
	f.ListFavorites(db)
}

// This runs if no "flags" are provided, but there may be arguments. 
func singleShotMode(db *sql.DB) {
	// if no argurments provided, print all books
	if len(os.Args) == 1 {
	var allBooksString string
		for i := 0; i < len(allBooks); i++ {
			// This is just for formatting. No comma and newline on last one
			if i == len(allBooks)-1 {
				allBooksString += allBooks[i] + "\n"
			// If not last one, just append the book with a comma
			} else {
				allBooksString += allBooks[i] + ", "
			}
		}

		f.WordWrap(allBooksString)	
		return
	}

	var passage Passage
	passage.BookName = os.Args[1]

	// If just a book is provided, Print number of chapters.
	if len(os.Args) == 2 {
		chapters := f.GetAllChaptersInBook(db, passage.BookName)
		if chapters == 0 {
			fmt.Printf("Can't find book \"%s\"\n\n", passage.BookName)
			return
		} else {
			fmt.Printf("Chapters in %s: %d\n", passage.BookName, chapters)
			fmt.Println()
		}

	// if a book and a chapter, print the entire chapter
	} else if len(os.Args) == 3 {
		passage.Chapter = os.Args[2]
		printChapters(db, passage)

	// if book and chapter and verse(s), print the verse(s)
	} else if len(os.Args) == 4 {
		passage.Chapter = os.Args[2]
		passage.Verse = os.Args[3]
		printVerses(db, passage)
	} else {
		fmt.Println("Please enter a correct verse\n")
	}
}


// This function runs if you provide only a book
// I don't really want to use this. I don't want to print entire books...But maybe. I'll leave it for now
func printBook() {

}


// This function runs if you provide 2 arguments, ie a book and a chapter or range of chapters
func printChapters(db *sql.DB, passage Passage) {
	// This is for a range of chapters ie "bible "1 Cornthians" 1-3"
	if strings.Contains(passage.Chapter, "-") {
		chapters, err := getIntsStartAndEnd(passage.Chapter)
		if err != nil {
			fmt.Println("Error getting all chapters: ", err)
		}

		// For every chapter
		for i := 0 ; i < len(chapters); i++{
			// We need to get the number of verses for the chapter
			verses := f.GetAllVersesInChapter(db, passage.BookName, strconv.Itoa(chapters[i]))

			// Check if returned 0. This means the chapter doens't exist
			if verses == 0 {
				fmt.Printf("Can't find chapter %d in book \"%s\"\n\n", chapters[i], passage.BookName)
				return
			}

			fmt.Printf("%s Chapter %d\n\n", passage.BookName, chapters[i])
			// For every verse
			for j := 1; j <= verses; j++ {
				f.PrintVerse(db, passage.BookName, strconv.Itoa(chapters[i]), strconv.Itoa(j))
			}
		}

	// This is for a single chapter ie "bible "1 Corinthians" 1"
	} else {
		verses := f.GetAllVersesInChapter(db, passage.BookName, passage.Chapter)
		// Check if returned 0. This means that the chapter doesn't exist. 
		if verses == 0 {
			fmt.Printf("Can't find chapter %s in book \"%s\"\n\n", passage.Chapter, passage.BookName)
			return
		}

		for i := 1; i <= verses; i++ {
			f.PrintVerse(db, passage.BookName, passage.Chapter, strconv.Itoa(i))
		}
	}
}


// This function runs if you provide all 3 arguments. Ie  a book, a chapter, and a verse or range of verses.
func printVerses(db *sql.DB, passage Passage) {
	// This is for a range of verses ie "bible "1 Corinthians" 1 1-5"
	if strings.Contains(passage.Verse, "-") {
		verses, err := getIntsStartAndEnd(passage.Verse)
		if err != nil {
			fmt.Println("Error gettings all verses: ", err)
		}
		for i := 0; i < len(verses); i ++ {
			f.PrintVerse(db, passage.BookName, passage.Chapter, strconv.Itoa(verses[i]))
		}

	// This is for a single verse
	} else {
		f.PrintVerse(db, passage.BookName, passage.Chapter, passage.Verse)
	}
}


// getIntsStartAndEnd parses a string in the format "start-end" (ie 5-10) and returns a slice of integers (ie [5,6,7,8,9,10]).
func getIntsStartAndEnd(s string) ([]int, error) {
	split := strings.Split(s, "-")
	if len(split) != 2 {
		return nil, fmt.Errorf("Invalid format, expected 'start-end'")
	}

	start, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, fmt.Errorf("Error converting start to int: %v", err)
	}

	end, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, fmt.Errorf("Error converting end to int: %v", err)
	}

	// Create a slice to hold the integers from start to end
	var ints []int
	for i := start; i <= end; i++ {
		ints = append(ints, i)
	}

	return ints, nil
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


// This is just for testing random things...
//func testFunction(db *sql.DB) {
	//var book string
	//var chapter string
	//var verse string
	//fmt.Println("Book: ")
	//fmt.Scan(&book)
	//fmt.Println("Chapter: ")
	//fmt.Scan(&chapter)
	//fmt.Println("Verse: ")
	//fmt.Scan(&verse)

	//id := f.GetIdOfVerse(db, book, chapter, verse)
	//fmt.Printf("ID: %d\n", id)
	//f.GetAllChaptersInBook(db, "gensis")
//}
