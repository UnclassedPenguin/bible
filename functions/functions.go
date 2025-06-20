package functions

import (
	"os"
	"fmt"
	"log"
	"time"
	"bufio"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"strconv"
	"math/rand"
	"database/sql"
	"encoding/json"
	"golang.org/x/term"
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


// Print Verse
// these are a string for a reason...I think beause random verse needs to return a []string, so it made it easier to do? because the bookname is a string,
// And I wanted it to return a single array
func PrintVerse(db *sql.DB, book string, chapter string, verse string) {
	chapterInt, _ := strconv.Atoi(chapter)
	verseInt, _ := strconv.Atoi(verse)

	var bibleVerse Bible
	err := db.QueryRow("SELECT id, bookName, chapter, verse, text FROM bible where bookName = ? AND chapter = ? AND verse = ?", book, chapterInt, verseInt).Scan(&bibleVerse.ID, &bibleVerse.BookName, &bibleVerse.Chapter, &bibleVerse.Verse, &bibleVerse.Text)
	if err != nil {
		fmt.Printf("Verse %s %s:%s not found\n", book, chapter, verse)
	}
	fmt.Printf("%s %s:%s\n", book, chapter, verse)
	WordWrap(bibleVerse.Text)
	fmt.Printf("\n")
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

func GetVerseFromId(db *sql.DB, id int) Bible {
	var verse Bible
	query := "SELECT bookName, chapter, verse, text FROM bible where id = ?"
	err := db.QueryRow(query, id).Scan(&verse.BookName, &verse.Chapter, &verse.Verse, &verse.Text)
	if err != nil {
		fmt.Printf("Can't get verse from id: %d\n", id)
		fmt.Println(err)
	}

	return verse
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


// -----------------------------------------------------------------------------
// Everything beneath here has to do with favorites/bookmarks
// -----------------------------------------------------------------------------

type SaveData struct {
	Bookmark  int   `json:"bookmark"`
	Favorites []int `json:"favorites"`
}

func (sd *SaveData) SetBookmark(id int) {
	sd.Bookmark = id
}

// AddFavorite function to add an integer to favorites if it's not already present
func (sd *SaveData) AddFavorite(item int) {
	if !sd.ContainsFavorite(item) {
		sd.Favorites = append(sd.Favorites, item)
		sort.Ints(sd.Favorites) // Sort the slice after adding
	}
}


// RemoveFavorite function to remove an integer from favorites
func (sd *SaveData) RemoveFavorite(item int) {
	for i, v := range sd.Favorites {
		if v == item {
			// Remove the item by slicing
			sd.Favorites = append(sd.Favorites[:i], sd.Favorites[i+1:]...)
			break // Exit the loop after removing the item
		}
	}
}


// ContainsFavorite function to check if an integer is already in favorites
func (sd *SaveData) ContainsFavorite(item int) bool {
	for _, v := range sd.Favorites {
		if v == item {
			return true
		}
	}
	return false
}


// Save function to save bookmarks and favorites to a file
func (sd *SaveData) Save(filename string) error {
	data, err := json.Marshal(sd)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}


// Load function to load bookmarks and favorites from a file
func (sd *SaveData) Load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, sd)
}


// GetDataFilePath function to get the data file path
func GetDataFilePath() string {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".local", "share", "bible")
	os.MkdirAll(dataDir, os.ModePerm) // Create the directory if it doesn't exist
	return filepath.Join(dataDir, "bible-data.json")
}

// This will be a fovorites feature. Need to save to a file and be able to read it back (probably "bible -f" will list all favorites)
func Favorites(db *sql.DB, id int) {

	saveData := &SaveData{}

	// Load existing data from file
	dataFilePath := GetDataFilePath()
	if err := saveData.Load(dataFilePath); err != nil && !os.IsNotExist(err) {
		fmt.Println("Error loading data:", err)
	}

	if saveData.ContainsFavorite(id) {
		verse := GetVerseFromId(db, id)
		fmt.Printf("%s %d:%d already in favorites\n", verse.BookName, verse.Chapter, verse.Verse)
		var choice string
		fmt.Printf("Remove from favorites? (y or N) ")
		fmt.Scanln(&choice)
		if choice == "y" {
			saveData.RemoveFavorite(id)
			fmt.Println("Verse WAS removed from favorites")

			// Save data to file
			if err := saveData.Save(dataFilePath); err != nil {
				fmt.Println("Error saving data:", err)
			}
		} else {
			fmt.Println("Verse WAS NOT removed from favorites")
		}
	} else {
		saveData.AddFavorite(id)
		fmt.Println("Added verse to favorites")

		// Save data to file
		if err := saveData.Save(dataFilePath); err != nil {
			fmt.Println("Error saving data:", err)
		}
	}
}


// This lists your favorites. (ie. bible -f)
func ListFavorites(db *sql.DB) {
	saveData := &SaveData{}

	// Load existing data from file
	dataFilePath := GetDataFilePath()
	if err := saveData.Load(dataFilePath); err != nil && !os.IsNotExist(err) {
		fmt.Println("Error loading data:", err)
	}

	for _, id := range saveData.Favorites {
		// Get info from id
		verse := GetVerseFromId(db, id)

		//Print Verse
		PrintVerse(db, verse.BookName, strconv.Itoa(verse.Chapter), strconv.Itoa(verse.Verse))
	}
}


// This will be a bookmark function in interactive mode
func BookMark(id int) int{
	var choice string
	for {
		fmt.Printf("Would you like to load or save bookmark? (l or s) ")
		fmt.Scanln(&choice)

		// Load the bookmark
		if choice == "l" {
			// LoadBookmark returns the id of the bookmark, so this just returns the id of the bookmark. Cyclical...
			return LoadBookmark()
		} else if choice == "s" {
			saveData := &SaveData{}

			// Load existing data from file
			dataFilePath := GetDataFilePath()
			if err := saveData.Load(dataFilePath); err != nil && !os.IsNotExist(err) {
				fmt.Println("Error loading data:", err)
			}

			saveData.SetBookmark(id)

			// Save data to file
			if err := saveData.Save(dataFilePath); err != nil {
				fmt.Println("Error saving data:", err)
			}
			
			fmt.Println("Saved Bookmark!")

			// Just keep the user on the same verse. 
			return id
		} else {
			fmt.Println("Please select either l or s")
		}
	}
	return -1 
}


// This is the load function for the very beginning of interactive mode. It needs to be seperate because I 
// don't want to have to send anything with it. It merely returns the int of the bookmark variable. 
func LoadBookmark() int {
	saveData := &SaveData{}

	// Load existing data from file
	dataFilePath := GetDataFilePath()
	if err := saveData.Load(dataFilePath); err != nil && !os.IsNotExist(err) {
		fmt.Println("Error loading data:", err)
	}

	// If a bookmark hasn't been saved yet, just send the user to Genesis 1 1
	if saveData.Bookmark == 0 {
		return 1
	}

	return saveData.Bookmark
}
