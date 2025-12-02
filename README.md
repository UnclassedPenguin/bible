This program is basically done. Check out [the docs](https://docs.unclassed.ca/bible)  

## Todo:  
- [ ] When you do bible -l I don't think it has wordwrap
- [ ] Need to find any more error handling that needs to be done  

## Finished:
- [x] Add a way to list all books  
- [x] Add a random verse function  
- [x] Add search function. Search for a word, give all verses with it in it  
- [x] randomVerse should be a function that returns [string, string, string] ie [bookName, chapter, verse]  
- [x] Add random function to interactive mode  
- [x] Add a parseInteractiveCommand() function to deal with the interactive command. Maybe only if its more than a single character or maybe just deal with everything...even single letter commands  
- [x] Need to add a special case for "Song of Solomon" in -i. Can't reach it for now. or maybe chqnge current function so it just grabs what is inbetween quotes  
- [x] Maybe change word wrap so that it always leaves at least 1 space on the right side...Basically have to just make termWidth = termWidth - 1 I think..  
- [x] Need some basic restructuring. Anytime a verse is printed, it should call a function (printVerse) that takes the book, the chapter, the verse. I think this will simplify things in the long run...  
- [x] Write documentation for program  
- [x] Clean up interactive mode, make it not so ugly  
    - [x] Interactive should be able to basically have a command line, sort of like vim. That at anytime you can do a book or a chapter or a verse  
- [x] Add a print function that wraps the print, so It doesn't split words
- [x] Combine if statements in infoMode function. Should be if else if  
- [x] Combine if statements in singleShotMode function. Should be if else if  
- [x] I think maybe i could change passage struct to have chapter and verse as int. It might clean up some code where I dont have to convert back and forth? Could change printVerse function  
    - [x] I tried this, I wasn't really happy with it...  
- [x] Add bookmark function  
- [x] Add favorite function  
- [ ] Add a print all?  
    - [x] nah
- [x] Need to fix in interactive mode what happens when you enter a verse that doesnt exist
- [x] Need to update help usage (-h flag) to add info about how to use single shot mode  
- [x] Add f.PrintInteractiveHelp function to print help usage when "?" entered at interactive prompt.  
- [x] Comment out test function  
- [x] Change description in PrintInteractiveHelp to be shorter, or use wordwrap.  
