package main

import (
	"log"
	"fmt"
	"net/http"
	"html/template"
	// "reflect"
	"strings"
	// "strconv"
	"io/ioutil"
	"time"
	"math/rand"
)

type Clue struct {
	Id		string
	Kind	string
}

type Context struct {
	Clues []Clue
	LastUpdated time.Time
	PuzzleId string
}

type Puzzles map[string]*Context

var P = make(Puzzles)
var NewestPuzzleId string

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

func getNewId() string {
	return getAdjective() + "-" + getAdjective() + "-" + getPokemon()
}

func create(w http.ResponseWriter, r *http.Request) {
	C := new(Context)
	puzzleId := getNewId()
	fmt.Printf("New PuzzleId %s\n", puzzleId)
	C.PuzzleId = puzzleId
	P[puzzleId] = C
	t,_ := template.ParseFiles("create.html")
	err := t.Execute(w,C)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
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
	fmt.Printf("PuzzleId %s\n", puzzleId)
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
	C.PuzzleId = puzzleId
	P[puzzleId] = C

	// output response
	t,_ := template.ParseFiles("view.html")
	err = t.Execute(w, C)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}
}

func main() {
	// create
	http.HandleFunc("/create", create)
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

func getAdjective() string {
	adjectives := []string{
		"good",
		"new",
		"first",
		"last",
		"long",
		"great",
		"little",
		"other",
		"old",
		"right",
		"big",
		"high",
		"different",
		"small",
		"large",
		"next",
		"early",
		"young",
		"important",
		"public",
		"bad",
		"able",
		"adorable",
		"beautiful",
		"clean",
		"drab",
		"elegant",
		"fancy",
		"glamorous",
		"handsome",
		"long",
		"magnificent",
		"plain",
		"quaint",
		"sparkling",
		"ugliest",
		"unsightly",
		"red",
		"orange",
		"yellow",
		"green",
		"blue",
		"purple",
		"gray",
		"black",
		"white",
		"alive",
		"better",
		"careful",
		"clever",
		"dead",
		"easy",
		"famous",
		"gifted",
		"helpful",
		"important",
		"inexpensive",
		"mushy",
		"odd",
		"powerful",
		"rich",
		"shy",
		"tender",
		"uninterested",
		"vast",
		"wrong",
		"agreeable",
		"brave",
		"calm",
		"delightful",
		"eager",
		"faithful",
		"gentle",
		"happy",
		"jolly",
		"kind",
		"lively",
		"nice",
		"obedient",
		"proud",
		"relieved",
		"silly",
		"thankful",
		"victorious",
		"witty",
		"zealous",
		"angry",
		"bewildered",
		"clumsy",
		"defeated",
		"embarrassed",
		"fierce",
		"grumpy",
		"helpless",
		"itchy",
		"jealous",
		"lazy",
		"mysterious",
		"nervous",
		"obnoxious",
		"panicky",
		"repulsive",
		"scary",
		"thoughtless",
		"uptight",
		"worried",
		"broad",
		"chubby",
		"crooked",
		"curved",
		"deep",
		"flat",
		"high",
		"hollow",
		"low",
		"narrow",
		"round",
		"shallow",
		"skinny",
		"square",
		"steep",
		"straight",
		"wide",
		"big",
		"colossal",
		"fat",
		"gigantic",
		"great",
		"huge",
		"immense",
		"large",
		"little",
		"mammoth",
		"massive",
		"miniature",
		"petite",
		"puny",
		"scrawny",
		"short",
		"small",
		"tall",
		"teeny",
		"tiny",
		"cooing",
		"deafening",
		"faint",
		"hissing",
		"loud",
		"melodic",
		"noisy",
		"purring",
		"quiet",
		"raspy",
		"screeching",
		"thundering",
		"voiceless",
		"whispering",
		"ancient",
		"brief",
		"early",
		"fast",
		"late",
		"long",
		"modern",
		"old",
		"quick",
		"rapid",
		"short",
		"slow",
		"swift",
		"young",
		"bitter",
		"delicious",
		"fresh",
		"greasy",
		"juicy",
		"hot",
		"icy",
		"loose",
		"melted",
		"nutritious",
		"prickly",
		"rainy",
		"rotten",
		"salty",
		"sticky",
		"strong",
		"sweet",
		"tart",
		"tasteless",
		"uneven",
		"weak",
		"wet",
		"wooden",
		"yummy",
		"boiling",
		"breezy",
		"broken",
		"bumpy",
		"chilly",
		"cold",
		"cool",
		"creepy",
		"crooked",
		"cuddly",
		"curly",
		"damaged",
		"damp",
		"dirty",
		"dry",
		"dusty",
		"filthy",
		"flaky",
		"fluffy",
		"freezing",
		"hot",
		"warm",
		"wet",
		"abundant",
		"empty",
		"few",
		"full",
		"heavy",
		"light",
		"many",
		"numerous",
		"sparse",
		"substantial",
	}
	return adjectives[rand.Intn(len(adjectives))]
}

func getPokemon() string {
	pokemon := []string {
		"Bulbasaur",
		"Ivysaur",
		"Venusaur",
		"Charmander",
		"Charmeleon",
		"Charizard",
		"Squirtle",
		"Wartortle",
		"Blastoise",
		"Caterpie",
		"Metapod",
		"Butterfree",
		"Weedle",
		"Kakuna",
		"Beedrill",
		"Pidgey",
		"Pidgeotto",
		"Pidgeot",
		"Rattata",
		"Raticate",
		"Spearow",
		"Fearow",
		"Ekans",
		"Arbok",
		"Pikachu",
		"Raichu",
		"Sandshrew",
		"Sandslash",
		"Nidoran",
		"Nidorina",
		"Nidoqueen",
		"Nidorino",
		"Nidoking",
		"Clefairy",
		"Clefable",
		"Vulpix",
		"Ninetales",
		"Jigglypuff",
		"Wigglytuff",
		"Zubat",
		"Golbat",
		"Oddish",
		"Gloom",
		"Vileplume",
		"Paras",
		"Parasect",
		"Venonat",
		"Venomoth",
		"Diglett",
		"Dugtrio",
		"Meowth",
		"Persian",
		"Psyduck",
		"Golduck",
		"Mankey",
		"Primeape",
		"Growlithe",
		"Arcanine",
		"Poliwag",
		"Poliwhirl",
		"Poliwrath",
		"Abra",
		"Kadabra",
		"Alakazam",
		"Machop",
		"Machoke",
		"Machamp",
		"Bellsprout",
		"Weepinbell",
		"Victreebel",
		"Tentacool",
		"Tentacruel",
		"Geodude",
		"Graveler",
		"Golem",
		"Ponyta",
		"Rapidash",
		"Slowpoke",
		"Slowbro",
		"Magnemite",
		"Magneton",
		"Farfetchd",
		"Doduo",
		"Dodrio",
		"Seel",
		"Dewgong",
		"Grimer",
		"Muk",
		"Shellder",
		"Cloyster",
		"Gastly",
		"Haunter",
		"Gengar",
		"Onix",
		"Drowzee",
		"Hypno",
		"Krabby",
		"Kingler",
		"Voltorb",
		"Electrode",
		"Exeggcute",
		"Exeggutor",
		"Cubone",
		"Marowak",
		"Hitmonlee",
		"Hitmonchan",
		"Lickitung",
		"Koffing",
		"Weezing",
		"Rhyhorn",
		"Rhydon",
		"Chansey",
		"Tangela",
		"Kangaskhan",
		"Horsea",
		"Seadra",
		"Goldeen",
		"Seaking",
		"Staryu",
		"Starmie",
		"MrMime",
		"Scyther",
		"Jynx",
		"Electabuzz",
		"Magmar",
		"Pinsir",
		"Tauros",
		"Magikarp",
		"Gyarados",
		"Lapras",
		"Ditto",
		"Eevee",
		"Vaporeon",
		"Jolteon",
		"Flareon",
		"Porygon",
		"Omanyte",
		"Omastar",
		"Kabuto",
		"Kabutops",
		"Aerodactyl",
		"Snorlax",
		"Articuno",
		"Zapdos",
		"Moltres",
		"Dratini",
		"Dragonair",
		"Dragonite",
		"Mewtwo",
		"Mew",
	}
	return pokemon[rand.Intn(151)]
}
