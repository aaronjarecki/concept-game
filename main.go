package main

import (
	"log"
	"fmt"
	"net/http"
	"html/template"
	// "reflect"
	"strings"
	"io/ioutil"
	"time"
)

type Clue struct {
	Id		string
	Kind	string
}

type Context struct {
	Clues []Clue
	LastUpdated time.Time
}

type Puzzles map[string]*Context

var P = make(Puzzles)
var NewestPuzzleId string

func (C *Context) GetClues(kind string) []Clue {
	clueSubset := make([]Clue, 0, len(C.Clues))
	for _, c := range C.Clues {
		if c.Kind == kind {
			clueSubset = append(clueSubset, c)
		}
	}
	return clueSubset
}

func (C *Context) toString() string {
	theStr := ""
	for _, c := range C.Clues {
		theStr = theStr + c.Id + ":" + c.Kind + ","
	}
	return theStr
}

func contextFromString(theStr string) *Context{
	newContext := new(Context)
	for _,c := range strings.Split(theStr, ",") {
		if c != "" {
			parts := strings.Split(c, ":")	
			newContext.Clues = append(newContext.Clues, Clue{Id: parts[0], Kind: parts[1]})
		}
	}
	return newContext
}

func pushItem(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")
	clueId := r.FormValue("clueId")
	clueKind := r.FormValue("clueKind")

	// update Context
	if P[puzzleId] == nil {
		P[puzzleId] = new(Context)
		NewestPuzzleId = puzzleId
	}
	C  := P[puzzleId]
	C.Clues = append(C.Clues, Clue{Id: clueId, Kind: clueKind})
	C.LastUpdated = time.Now()

	// debug
	fmt.Printf("Primary %v\n", C.GetClues("0"))
	fmt.Printf("Secondary %v\n", C.GetClues("1"))
	fmt.Printf("Tertiary %v\n\n", C.GetClues("2"))

	// output response
	t,_ := template.ParseFiles("view.html")
	err := t.Execute(w, C)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
}

func popItem(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")
	if P[puzzleId] == nil {
		// Error
		fmt.Printf("Error: No such puzzle Id\n")
	}

	C := P[puzzleId]
	C.Clues = C.Clues[:len(C.Clues)-1] 
	fmt.Printf("Primary %v\n", C.GetClues("0"))
	fmt.Printf("Secondary %v\n", C.GetClues("1"))
	fmt.Printf("Tertiary %v\n\n", C.GetClues("2"))

	// output response
	t,_ := template.ParseFiles("view.html")
	err := t.Execute(w, C)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
}

func clear(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")
	P[puzzleId] = new(Context)

	// output response
	t,_ := template.ParseFiles("view.html")
	err := t.Execute(w, P[puzzleId])
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
}

func watch(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")

	// get Context
	if P[puzzleId] == nil {
		// Error
		fmt.Printf("Error: No such puzzle Id\n")
	}
	C := P[puzzleId]

	// output response
	t,_ := template.ParseFiles("watch.html")
	err := t.Execute(w, C)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
}

func watchRecent(w http.ResponseWriter, r *http.Request) {
	// get Context
	if P[NewestPuzzleId] == nil {
		// Error
		fmt.Printf("Error: No such puzzle Id\n")
	}
	C := P[NewestPuzzleId]

	// output response
	t,_ := template.ParseFiles("watch.html")
	err := t.Execute(w, C)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
}

func save(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")

	// save context to file
	str := P[puzzleId].toString()
	filename := puzzleId + ".con"
	ioutil.WriteFile(filename, []byte(str), 0600)
}

func load(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")

	// get Context from file
	filename := puzzleId + ".con"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error loading Context %s: %v", puzzleId, err)
	}
	C := contextFromString(string(body))

	// output response
	t,_ := template.ParseFiles("view.html")
	err = t.Execute(w, C)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
}

func main() {
	// pushItem?puzzleId=001&clueId=001&clueKind=0
	http.HandleFunc("/pushItem", pushItem)
	// popItem?puzzleId=001
	http.HandleFunc("/popItem", popItem)
	// clear?puzzleId=001
	http.HandleFunc("/clear", clear)
	// watch?puzzleId=001
	http.HandleFunc("/watch", watch)
	// watchRecent
	http.HandleFunc("/watchRecent", watchRecent)
	// save?puzzleId=001
	http.HandleFunc("/save", save)
	// load?puzzleId=001
	http.HandleFunc("/load", load)

	http.Handle("/", http.FileServer(http.Dir("assets")))
	log.Fatal(http.ListenAndServe(":8888", nil))
}