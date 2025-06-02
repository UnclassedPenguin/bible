package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"

    _ "github.com/mattn/go-sqlite3"
)

// Define a struct to match the JSON structure for each verse
type Bible struct {
    BookName string `json:"book_name"`
    Book     int    `json:"book"`
    Chapter  int    `json:"chapter"`
    Verse    int    `json:"verse"`
    Text     string `json:"text"`
}

// Define a struct to match the outer JSON structure
type BibleData struct {
    Verses []Bible `json:"verses"`
}

func main() {
    fmt.Println("Starting the JSON to SQLite conversion...")

    // Read the JSON file
    jsonFile, err := os.Open("kjv.json")
    if err != nil {
        log.Fatalf("Error opening JSON file: %v\n", err)
    }
    defer jsonFile.Close()
    fmt.Println("Successfully opened JSON file.")

    // Read the file content
    byteValue, err := ioutil.ReadAll(jsonFile)
    if err != nil {
        log.Fatalf("Error reading JSON file: %v\n", err)
    }
    fmt.Println("Successfully read JSON file content.")

    // Unmarshal the JSON data into the BibleData struct
    var bibleData BibleData
    if err := json.Unmarshal(byteValue, &bibleData); err != nil {
        log.Fatalf("Error unmarshaling JSON data: %v\n", err)
    }
    fmt.Printf("Successfully unmarshaled JSON data. Found %d verses.\n", len(bibleData.Verses))

    // Create or open the SQLite database
    db, err := sql.Open("sqlite3", "./kjv.db")
    if err != nil {
        log.Fatalf("Error opening SQLite database: %v\n", err)
    }
    defer db.Close()
    fmt.Println("Successfully opened SQLite database.")

    // Create a table
    createTableSQL := `CREATE TABLE IF NOT EXISTS bible (
        id INTEGER PRIMARY KEY,
        bookName TEXT,
        book INTEGER,
        chapter INTEGER,
        verse INTEGER,
        text TEXT
    );`
    if _, err := db.Exec(createTableSQL); err != nil {
        log.Fatalf("Error creating table: %v\n", err)
    }
    fmt.Println("Successfully created table in the database.")

    // Insert data into the table
    for _, verse := range bibleData.Verses {
        _, err := db.Exec("INSERT INTO bible (bookName, book, chapter, verse, text) VALUES (?, ?, ?, ?, ?)", verse.BookName, verse.Book, verse.Chapter, verse.Verse, verse.Text)
        if err != nil {
            log.Fatalf("Error inserting verse (Book: %s, Chapter: %d, Verse: %d): %v\n", verse.BookName, verse.Chapter, verse.Verse, err)
        }
        fmt.Printf("Inserted verse: %s %d:%d - %s\n", verse.BookName, verse.Chapter, verse.Verse, verse.Text)
    }

    fmt.Println("All verses successfully inserted into the SQLite database.")
}

