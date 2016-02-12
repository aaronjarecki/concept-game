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
	"image/png"
	"bytes"
	"bufio"
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

func getCluesOfEachKind(C Context) map[string][]Clue {
	kinds := make(map[string][]Clue)
	for _,c := range C.Clues {
		kinds[c.Kind] = append(kinds[c.Kind], c)
	}
	return kinds
}

func getClueImage(id string) image.Image {
	f, err := os.Open("assets/"+id+".png")
	if err != nil {
	    log.Print(err)
	}
	defer f.Close()
	img, _, err := image.Decode(bufio.NewReader(f))
	if err != nil {
		log.Print(err)
	}
	return img
}

func getConImage(kind string) image.Image {
	f, err := os.Open("assets/ico-square-"+kind+".png")
	if err != nil {
	    log.Print(err)
	}
	defer f.Close()
	img, _, err := image.Decode(bufio.NewReader(f))
	if err != nil {
		log.Print(err)
	}
	return img
}

func getClueImageWithCon(id string, kind string) image.Image {
	clueImg := getClueImage(id)
	conImg := getConImage(kind)
	clueMaxPt := clueImg.Bounds().Max
	conImgOffset := clueMaxPt.Sub(conImg.Bounds().Max)
	base := image.NewRGBA(clueImg.Bounds())
	draw.Draw(base, clueImg.Bounds(), clueImg, image.ZP, draw.Src)
	draw.Draw(base, conImg.Bounds().Add(conImgOffset), conImg, image.ZP, draw.Src)
	return base
}

func (C *Context) outputContextAsPNG(w http.ResponseWriter) {
	kinds := getCluesOfEachKind(*C)

	// Make base image of correct size
	maxHeight := len(kinds)
	maxWidth := 0
	for _,clues := range kinds {
		if len(clues) > maxWidth {
			maxWidth = len(clues)
		}
	}
	maxHeight = maxHeight * 322
	maxWidth = maxWidth * 322
	if maxHeight == 0 || maxWidth == 0 {
		maxHeight = 1
		maxWidth = 1
	}
	base := image.NewRGBA(image.Rect(0, 0, maxWidth, maxHeight))
	draw.Draw(base, base.Bounds(), image.Transparent, image.ZP, draw.Src)

	// Add concepts to image
	yPt := 2
	for kind := 0; kind < 4; kind++ {
		kindstr := strconv.Itoa(kind)
		for i, c := range kinds[kindstr] {
			xPt := i*320 + 2
			offset := image.Pt(xPt, yPt)
			log.Print(offset.String())
			clueImg := getClueImageWithCon(c.Id, kindstr)
			draw.Draw(base, clueImg.Bounds().Add(offset), clueImg, image.ZP, draw.Src)
		}
		if len(kinds[kindstr]) > 0 {
			yPt += 322
		}
	}

	// Encode image as PNG
	pngEncoding := new(bytes.Buffer)
	if err := png.Encode(pngEncoding, base); err != nil {
		log.Print("unable to encode image.")
	}

	// Output png encoding
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(pngEncoding.Bytes())))
	if _, err := w.Write(pngEncoding.Bytes()); err != nil {
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
	var C *Context
	C, ok := P[puzzleId]
	if !ok {
		log.Print("Request for %s\nGoing to load from DB\n", puzzleId)
		var author, solution, clueStr string
		if err := loadFromDB(puzzleId, &author, &solution, &clueStr); err != nil {
			log.Print(err)
		} else {
			log.Print("Got Clue String %s\n", clueStr)
			C = contextFromString(clueStr)
			C.PuzzleId = puzzleId
			P[puzzleId] = C
		}
	}

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

func saveToDB(puzzleId string, author string, solution string) error {
	// save to db
	str := P[puzzleId].toString()
	_, err := DB.Exec("INSERT INTO puzzles(ident, author, solution, clues) VALUES(?, ?, ?, ?)", puzzleId, author, solution, str)
	return err
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")
	solution := r.FormValue("solution")
	author := r.FormValue("author")
	if solution == "" {
		solution = "none"
	}
	if author == "" {
		author = "anonymous"
	}

	log.Print("PuzzleId: %s\nSolution: %s\nAuthor: %s\n", puzzleId, solution, author)

	// save to db
	if err := saveToDB(puzzleId, author, solution); err != nil {
		log.Print(err)
	}
	log.Print("Saved puzzle %s:\nAuthor: %s\nSolution:%s\n", puzzleId, author, solution)
}

func loadFromDB(puzzleId string, author *string, solution *string, clueStr *string) error {
	return DB.QueryRow("select author, solution, clues from puzzles where ident = ?", puzzleId).Scan(author,solution,clueStr)
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")

	var author, solution, clueStr string
	if err := loadFromDB(puzzleId, &author, &solution, &clueStr); err != nil {
		log.Print(err)
	}

	C := contextFromString(clueStr)
	C.PuzzleId = puzzleId
	P[puzzleId] = C

	// output response
	log.Print("Loaded puzzle %s:\nAuthor: %s\nSolution:%s\n", puzzleId, author, solution)
	t, _ := template.ParseFiles("watch.html")
	if err := t.Execute(w, C); err != nil {
		log.Print("Error %v\n", err)
	}
}

func dbBrowse(w http.ResponseWriter, r *http.Request) {
	puzzles := make([]map[string]string, 0, 100000)
	var puzzleId, author, solution string
	rows, err := DB.Query("select ident, author, solution from puzzles")
	if err != nil {
		log.Print(err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&puzzleId, &author, &solution); err == nil {
			puzzles = append(puzzles, map[string]string{"puzzleId":puzzleId,"author":author,"solution":solution})
		} else {
			log.Print(err)
		}
	}
	if err = rows.Err(); err != nil {
		log.Print(err)
	}

	// output response
	log.Print("Loaded %s puzzles\n", len(puzzles))
	t, _ := template.ParseFiles("dbBrowse.html")
	err = t.Execute(w, puzzles)
	if err != nil {
		log.Print("Error %v\n", err)
	}
}

func dbClear(w http.ResponseWriter, r *http.Request) {
	_, err := DB.Exec("truncate puzzles")
	if err != nil {
		log.Print(err)
	}
	log.Print("Removed all records from puzzle table\n")
}

func deletePuzzle(w http.ResponseWriter, r *http.Request) {
	// parse args
	puzzleId := r.FormValue("puzzleId")

	_, err := DB.Exec("delete from puzzles where ident = ?", puzzleId)
	if err != nil {
		log.Print(err)
	}
	log.Print("Removed %s from DB\n")
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
		log.Print("Error parsing null DSN: %s\n", err)
	}
	cfg.User = DBCreds.Username
	cfg.Passwd = DBCreds.Password
	cfg.Addr = DBCreds.Hostname + ":" + strconv.Itoa(DBCreds.Port)
	cfg.DBName = DBCreds.Name
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Print(err, nil)
	}
	if err = db.Ping(); err != nil {
		log.Print(err, nil)
	}
	_, err = db.Exec("DROP TABLE puzzles")
	_, err = db.Exec("CREATE TABLE puzzles (ident VARCHAR(50) UNIQUE NOT NULL, author TEXT, solution TEXT, clues MEDIUMTEXT)")
	if err != nil && !strings.Contains(err.Error(), "Table 'puzzles' already exists"){
		log.Print(err, nil)
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
	http.HandleFunc("/save", saveHandler)
	// load?puzzleId=001
	http.HandleFunc("/load", loadHandler)
	// dbBrowse
	http.HandleFunc("/dbBrowse", dbBrowse)
	// dbBrowse
	http.HandleFunc("/dbClear", dbClear)
	// dbBrowse
	http.HandleFunc("/deletePuzzle", deletePuzzle)

	http.Handle("/", http.FileServer(http.Dir("assets")))
	log.Print(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
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
