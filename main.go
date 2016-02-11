package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"database/sql"
	"encoding/json"
	"github.com/aaronjarecki/concept-game/Godeps/_workspace/src/github.com/go-sql-driver/mysql"
	"math/rand"
	"os"
	"time"
	"strconv"
	"image/draw"
	"image"
	"image/color"
	"image/png"
	"bytes"
)

type MySQLCredentials struct {
	Hostname 	string
	Port		int
	Name		string
	Username	string
	Password	string
}

type MySQLProperties struct {
	Name 		string
	Label		string
	Plan		string
	Credentials	MySQLCredentials
}

type VcapServices struct {
	Pmysql 	[]MySQLProperties `json:"p-mysql"`
}

type Clue struct {
	Id   string
	Kind string
}

type Context struct {
	Clues       []Clue
	LastUpdated time.Time
	PuzzleId    string
}

type Puzzles map[string]*Context

var P = make(Puzzles)
var NewestPuzzleId string
var DBCreds MySQLCredentials
var DB *sql.DB

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

func (C *Context) GetCluesJson(kind string) string {
	clues := C.GetClues(kind)
	cluesJson, err := json.Marshal(clues)
	if err != nil {
		log.Print("Error encoding json: %s\n", err)
	}
	return string(cluesJson)
}

func (C *Context) toString() string {
	theStr := ""
	for _, c := range C.Clues {
		theStr = theStr + c.Id + ":" + c.Kind + ","
	}
	return theStr
}

func contextFromString(theStr string) *Context {
	newContext := new(Context)
	for _, c := range strings.Split(theStr, ",") {
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

func getNumOfEachKind(C Context) map[string]int {
	kinds := make(map[string]int)
	for _,c := range C.Clues {
		kinds[c.Kind]++
	}
	return kinds
}

func (C *Context) outputContextAsPNG(w http.ResponseWriter) {
	kinds := getNumOfEachKind(*C)
	maxHeight := len(kinds)
	maxWidth := 0
	for _,k := range kinds {
		if k > maxWidth {
			maxWidth = k
		}
	}
	maxHeight = maxHeight * 322
	maxWidth = maxWidth * 322

	m := image.NewRGBA(image.Rect(0, 0, maxWidth, maxHeight))
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, m); err != nil {
		log.Print("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Print("unable to write image.")
	}
}

func create(w http.ResponseWriter, r *http.Request) {
	puzzleId := getNewId()
	NewestPuzzleId = puzzleId
	log.Print("New PuzzleId %s\n", puzzleId)
	C := new(Context)
	C.PuzzleId = puzzleId
	P[puzzleId] = C
	t, _ := template.ParseFiles("create.html")
	err := t.Execute(w, C)
	if err != nil {
		log.Print("Error %v\n", err)
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
	C := P[puzzleId]
	C.Clues = append(C.Clues, Clue{Id: clueId, Kind: clueKind})
	C.LastUpdated = time.Now()

	// debug
	log.Print("PuzzleId %s\n", puzzleId)
	log.Print("Primary %v\n", C.GetClues("0"))
	log.Print("Secondary %v\n", C.GetClues("1"))
	log.Print("Tertiary %v\n\n", C.GetClues("2"))

	// output response
	t, _ := template.ParseFiles("view.html")
	err := t.Execute(w, C)
	if err != nil {
		log.Print("Error %v\n", err)
	}
}

func popItem(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")
	if P[puzzleId] == nil {
		// Error
		log.Print("Error: No such puzzle Id\n")
	}

	C := P[puzzleId]
	C.Clues = C.Clues[:len(C.Clues)-1]
	log.Print("Primary %v\n", C.GetClues("0"))
	log.Print("Secondary %v\n", C.GetClues("1"))
	log.Print("Tertiary %v\n\n", C.GetClues("2"))

	// output response
	t, _ := template.ParseFiles("view.html")
	err := t.Execute(w, C)
	if err != nil {
		log.Print("Error %v\n", err)
	}
}

func getConcept(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")
	clueKind := r.FormValue("clueKind")

	C := P[puzzleId]
	clues := C.GetCluesJson(clueKind)

	fmt.Fprintf(w, clues)
}

func clear(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")
	P[puzzleId] = new(Context)

	// output response
	t, _ := template.ParseFiles("view.html")
	err := t.Execute(w, P[puzzleId])
	if err != nil {
		log.Print("Error %v\n", err)
	}
}

func watch(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")

	// get Context
	if P[puzzleId] == nil {
		// Error
		log.Print("Error: No such puzzle Id\n")
	}
	C := P[puzzleId]

	// output response
	t, _ := template.ParseFiles("watch.html")
	err := t.Execute(w, C)
	if err != nil {
		log.Print("Error %v\n", err)
	}
}

func view(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")
	asPng := r.FormValue("asPng")

	// get Context
	if P[puzzleId] == nil {
		// Error
		log.Print("Error: No such puzzle Id\n")
	}
	C := P[puzzleId]

	if asPng != "" && asPng != "false" {
		C.outputContextAsPNG(w)
	} else {
		// output response
		t, _ := template.ParseFiles("view.html")
		err := t.Execute(w, C)
		if err != nil {
			log.Print("Error %v\n", err)
		}
	}
}

func watchRecent(w http.ResponseWriter, r *http.Request) {
	// get Context
	if P[NewestPuzzleId] == nil {
		// Error
		log.Print("Error: No such puzzle Id\n")
	}
	C := P[NewestPuzzleId]

	// output response
	t, _ := template.ParseFiles("watch.html")
	err := t.Execute(w, C)
	if err != nil {
		log.Print("Error %v\n", err)
	}
}

func save(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")

	// save to db
	str := P[puzzleId].toString()
	_, err := DB.Exec("INSERT INTO puzzles(ident, clues) VALUES(?, ?)", puzzleId, str)
	if err != nil {
		log.Fatal(err)
	}
}

func load(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")

	var clueStr string
	err := DB.QueryRow("select clues from puzzles where ident = ?", puzzleId).Scan(&clueStr)
	if err != nil {
		log.Fatal(err)
	}

	C := contextFromString(clueStr)
	C.PuzzleId = puzzleId
	P[puzzleId] = C

	// output response
	t, _ := template.ParseFiles("watch.html")
	err = t.Execute(w, C)
	if err != nil {
		log.Print("Error %v\n", err)
	}
}

func parseEnv() {
	var VcapServices VcapServices
	err := json.Unmarshal([]byte(os.Getenv("VCAP_SERVICES")), &VcapServices)
	if err != nil {
		log.Print("Error parsing VCAP_SERVICES: %s\n", err)
	}
	DBCreds = VcapServices.Pmysql[0].Credentials
}

func openDB() *sql.DB {
	cfg, err := mysql.ParseDSN("")
	if err != nil {
		log.Fatal("Error parsing null DSN: %s\n", err)
	}
	cfg.User = DBCreds.Username
	cfg.Passwd = DBCreds.Password
	cfg.Addr = DBCreds.Hostname + ":" + strconv.Itoa(DBCreds.Port)
	cfg.DBName = DBCreds.Name
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err, nil)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err, nil)
	}

	_, err = db.Exec("CREATE TABLE puzzles (ident TEXT, clues MEDIUMTEXT)")
	if err != nil && !strings.Contains(err.Error(), "Table 'puzzles' already exists"){
		log.Fatal(err, nil)
	}

	return db
}

func main() {
	parseEnv()
	DB = openDB()

	// create
	http.HandleFunc("/create", create)
	// pushItem?puzzleId=001&clueId=001&clueKind=0
	http.HandleFunc("/pushItem", pushItem)
	// popItem?puzzleId=001
	http.HandleFunc("/popItem", popItem)
	// pushItem?puzzleId=001&clueKind=0
	http.HandleFunc("/getConcept", getConcept)
	// clear?puzzleId=001
	http.HandleFunc("/clear", clear)
	// view?puzzleId=001
	http.HandleFunc("/view", view)
	// watch?puzzleId=001
	http.HandleFunc("/watch", watch)
	// watchRecent
	http.HandleFunc("/watchRecent", watchRecent)
	// save?puzzleId=001
	http.HandleFunc("/save", save)
	// load?puzzleId=001
	http.HandleFunc("/load", load)

	http.Handle("/", http.FileServer(http.Dir("assets")))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
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
	pokemon := []string{
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
