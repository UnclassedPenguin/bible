## Todo:  
[ ] - add bookmark function  
[ ] - add favorite function  
[ ] - Add a print all?  

## Finished:
[x] - add a way to list all books.  
[x] - Add a random verse function  
[x] - add search function. Search for a word, give all verses with it in it  
[x] - randomVerse should be a fucntion that returns [string, int, int] ie [bookName, chapter, verse]  
[x] - Add random function to interactive mode  
[x] - add a parseInteractiveCommand() function to deal with the interavtive command. maybe only if its more than a single character. or maybe just dral with everuthing...even single letter commands  
[x] - need to add a special case for "song of solomon" in -i. cant reach it for now. or maybe chqnge current funxrion ao it juat grabs whats inbetween quotes.  
[x] - Maybe change word wrap so that it always leaves at least 1 space on the right side...Basically have to just make termWidth = termWidth - 1 I think..  
[x] - Need some basic restructuring. Anytime a verse is printed, it should call a function (printVerse) that takes the book, the chapter, the verse. I think this will simplify things in the long run...  
[x] - Write documentation for program  
[x] - Clean up interactive mode, make it not so ugly.   
   [x] - Interactive should be able to basically have a command line, sort of like vim. That at anytime you can do a book or a chapter or a verse.  
[x] - Add a print function that wraps the print, so It doesn't split words.  
