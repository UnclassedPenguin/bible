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

func main() {
	// Version number
	version := "0.0.3"

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
	flag.Parse()

	//db, err := sql.Open("sqlite3", "./kjv.db")
	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
			log.Fatal(err)
	}
	defer db.Close()

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
	} else {
		singleShotMode(db)
	}
}

// This returns a random book, chapter and verse in a string array
func randomVerse(db *sql.DB) []string {
	rand.Seed(time.Now().UnixNano())

	// Get a random book
	randomBook := allBooks[rand.Intn(66)]

	// Get number of chapters in random book
	chapters := getAllChaptersInBook(db, randomBook)

	// Get random chapter
	randomChapter := chapters[rand.Intn(len(chapters))]

	// Get number of verses in chapter
	verses := getAllVersesInChapter(db, randomBook, strconv.Itoa(randomChapter))

	// Get random verse
	randomVerse := verses[rand.Intn(len(verses))]

	return []string{randomBook, strconv.Itoa(randomChapter), strconv.Itoa(randomVerse)}
}

// Print Verse
func printVerse(db *sql.DB, book string, chapter string, verse string) {
	
	chapterInt, _ := strconv.Atoi(chapter)
	verseInt, _ := strconv.Atoi(verse)

	var bibleVerse Bible
	err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible where bookName = ? AND chapter = ? AND verse = ?", book, chapterInt, verseInt).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
	if err != nil {
		fmt.Printf("Verse %s %s:%s not found\n", book, chapter, verse)
	}
	fmt.Printf("%s %s %s:\n%s\n", book, chapter, verse, bibleVerse.Text)
}

// Fucntion to print a random verse. use -r on command line
func printRandomVerse(db *sql.DB) {
	// Get random verse
	random := randomVerse(db)
	// Print random verse
	printVerse(db, random[0], random[1], random[2])
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

	// If just a book is provided
	if len(os.Args) == 3 {
		var args Args
		args.BookName = os.Args[2]

		chapters := getAllChaptersInBook(db, args.BookName)
		fmt.Printf("Chapters in %s: %d\n", args.BookName, len(chapters))
	}

	// if a book and a chapter
	if len(os.Args) == 4 {
		var args Args
		args.BookName = os.Args[2]
		args.Chapters = os.Args[3]
		verses := getAllVersesInChapter(db, args.BookName, args.Chapters)
		fmt.Printf("Verses in %s %s: %d\n", args.BookName, args.Chapters, len(verses))
	}
}


// This 'worked" But not great...
// //Search for a term or an exact term
//func searchForTerm(db *sql.DB, exactMode bool) {
//	// This executes an exact search for the search term
//	if exactMode {
//		query := "select bookName, chapter, verse, text from bible where text like ?"
//		rows, err := db.Query(query, "% "+os.Args[3]+" %")
//		if err != nil {
//			log.Fatal(err)
//		}
//    defer rows.Close()
//
//		if !rows.Next() {
//			fmt.Println("No search found matching: ", os.Args[3])
//		}
//
//		for rows.Next() {
//			var bible Bible
//			err = rows.Scan(&bible.BookName, &bible.Chapter, &bible.Verse, &bible.Text)
//			if err != nil {
//				fmt.Println("Error reading row(exact search): ", err)
//				os.Exit(1)
//			}
//
//			fmt.Printf("%s %d %d: %s\n", bible.BookName, bible.Chapter, bible.Verse, bible.Text)
//		}
//
//	// This executes a search that takes any match, ie love with match with loved
//	} else {
//		query := "select bookName, chapter, verse, text from bible where text like ?"
//		rows, err := db.Query(query, "%"+os.Args[2]+"%")
//		if err != nil {
//			log.Fatal(err)
//		}
//    defer rows.Close()
//
//		if !rows.Next() {
//			fmt.Println("No search found matching: ", os.Args[3])
//		}
//
//
//		for rows.Next() {
//			var bible Bible
//			err = rows.Scan(&bible.BookName, &bible.Chapter, &bible.Verse, &bible.Text)
//			if err != nil {
//				fmt.Println("Error reading row(exact search): ", err)
//				os.Exit(1)
//			}
//
//			fmt.Printf("%s %d %d: %s\n", bible.BookName, bible.Chapter, bible.Verse, bible.Text)
//		}
//
//		}
//	os.Exit(0)
//}


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

			fmt.Printf("%s %d %d: %s\n", bible.BookName, bible.Chapter, bible.Verse, bible.Text)

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
			fmt.Println("Error in query of exact search: ", err)
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
				fmt.Println("Error reading row(exact search): ", err)
				os.Exit(1)
			}

			fmt.Printf("%s %d %d: %s\n", bible.BookName, bible.Chapter, bible.Verse, bible.Text)

			if !rows.Next() {
				break
			}
		}
	}
}


func singleShotMode(db *sql.DB) {

	// Print all books
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

	// If just a book is provided
	if len(os.Args) == 2 {
		chapters := getAllChaptersInBook(db, args.BookName)
		fmt.Printf("Chapters in %s:\n", args.BookName)
		for i, num := range chapters {
			if i > 0 {
				fmt.Print(" ") // Print a space before each number except the first
			}
			fmt.Print(num)
		}
		fmt.Println()
	}

	// if a book and a chapter
	if len(os.Args) == 3 {
		args.Chapters = os.Args[2]
		printChapters(db, args)
	}

	// if book and chapter and verse(s)
	if len(os.Args) == 4 {
		args.Chapters = os.Args[2]
		args.Verses = os.Args[3]
		printVerses(db, args)
	}
}


// This function runs if you provide only a book
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

		for i := 0 ; i < len(chapters); i++{
			fmt.Println("Chapter ", chapters[i])
			// We need to get all the verses for the chapter
			verses := getAllVersesInChapter(db, args.BookName, strconv.Itoa(chapters[i]))
			var bibleVerse Bible
			for j := 0; j < len(verses); j++ {
				err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible where bookName = ? AND chapter = ? AND verse = ?", args.BookName, chapters[i], verses[j]).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
				if err != nil {
					fmt.Printf("Verse %s %d:%d not found\n", args.BookName, args.Chapters, verses[j])
				}
				fmt.Printf("%d: %s\n", verses[j], bibleVerse.Text)
			}
		}
	// This is for a single chapter ie "bible "1 Corinthians" 1"
	} else {
	verses := getAllVersesInChapter(db, args.BookName, args.Chapters)
	fmt.Println(verses)
	var bibleVerse Bible
		for i := 0; i < len(verses); i++ {
			err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible where bookName = ? AND chapter = ? AND verse = ?", args.BookName, args.Chapters, verses[i]).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
			if err != nil {
				fmt.Printf("Verse %s %d:%d not found\n", args.BookName, args.Chapters, verses[i])
			}
			fmt.Printf("%d: %s\n", verses[i], bibleVerse.Text)
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
		var bibleVerse Bible
		for i := 0; i < len(verses); i ++ {
			err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible where bookName = ? AND chapter = ? AND verse = ?", args.BookName, args.Chapters, verses[i]).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
			if err != nil {
				fmt.Printf("Verse %s %d:%d not found\n", args.BookName, args.Chapters, verses[i])
			}
			fmt.Printf("%d: %s\n", verses[i], bibleVerse.Text)
		}
	// This is for a single verse
	} else {
		var bibleVerse Bible
		err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible where bookName = ? AND chapter = ? AND verse = ?", args.BookName, args.Chapters, args.Verses).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
		if err != nil {
			fmt.Printf("Verse %s %d:%d not found\n", args.BookName, args.Chapters, args.Verses)
		}
		fmt.Printf("%s\n", bibleVerse.Text)
	}
}

// This gives all chapters in a book
func getAllChaptersInBook(db *sql.DB, bookName string) []int {
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
	return uniqueChapters
}

// This gives the number of verses in a chapter
func getAllVersesInChapter(db *sql.DB, bookName string, chapter string) []int {
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
    return verses
}

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


func interactiveMode(db *sql.DB) {
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

	bookChapterVerseSplit := strings.Split(userInput, " ")
	bookName := bookChapterVerseSplit[0]
	chapter, _ := strconv.Atoi(bookChapterVerseSplit[1])
	startVerse, _ := strconv.Atoi(bookChapterVerseSplit[2])
	currentVerse := startVerse

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

	fmt.Print("\nPress 'n' for next verse, 'p' for prev, or 'q' to quit: \n\n")

	for {
		var bibleVerse Bible
		err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible WHERE bookName = ? AND chapter = ? AND verse = ?", bookName, chapter, currentVerse).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
		if err != nil {
			fmt.Printf("Verse %s %d:%d not found.\n", bookName, chapter, currentVerse)
			break
		}

		fmt.Printf("%s %d:%d - %s\n", bibleVerse.BookName, bibleVerse.Chapter, bibleVerse.Verse, bibleVerse.Text)

		fmt.Print(": ")
		var input string
		fmt.Scan(&input)

		switch strings.ToLower(input) {
		case "n":
			currentVerse++
		case "p":
			if currentVerse > 1 {
				currentVerse--
			} else {
				fmt.Println("You are at the first verse.")
			}
		case "q":
			fmt.Println("Exiting interactive mode.")
			return
		default:
			fmt.Println("Invalid input. Please enter 'n', 'p', or 'q'.")
		}

		// Clear the console for better readability
		clearConsole()
	}
}

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
