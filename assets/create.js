var KIND = "0"
var CONCEPTS = {}
var PUZZLEID = ""

function getConcept(kind) {
	var xmlHttp = new XMLHttpRequest();
	xmlHttp.onreadystatechange = function() {
		if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
			updateConcept(JSON.parse(xmlHttp.responseText), kind);
	}
	var url = "http://concept-game.cfapps.pez.pivotal.io/getConcept?puzzleId="+PUZZLEID+"&clueKind="+kind
	xmlHttp.open("GET", url, true); // true for asynchronous 
	xmlHttp.send(null);
}

function updateConcept(clues, kind) {
	var mainView = document.getElementById("MainView");
	var old_outer_div = document.getElementById("ConceptDiv"+kind);
	var outer_div = document.createElement("div");
	outer_div.setAttribute("id","ConceptDiv"+kind);
	for(var i = 0, len = clues.length; i < len; i++) {
		var inner_div = document.createElement("div");
		inner_div.setAttribute("class","IconDiv");

		var clue_img = document.createElement("img");
		clue_img.setAttribute("class","ClueImg");
		clue_img.setAttribute("src",clues[i].Id+".png");
		var concept_img = document.createElement('img');
		concept_img.setAttribute("class","ConceptIcon");
		concept_img.setAttribute('src', 'ico-square-'+kind+'.png');

		inner_div.appendChild(clue_img);
		inner_div.appendChild(concept_img);
		outer_div.appendChild(inner_div);
	}
	for (var key in CONCEPTS) {
		if (CONCEPTS.hasOwnProperty(key)) {
			mainView.removeChild(CONCEPTS[key])
		}
	}
	CONCEPTS[kind] = outer_div
	for (var key in CONCEPTS) {
		if (CONCEPTS.hasOwnProperty(key)) {
			mainView.appendChild(CONCEPTS[key])
		}
	}
}

function PushItem(id) {
	id = pad(id)
	var url = "http://concept-game.cfapps.pez.pivotal.io/pushItem?puzzleId="+PUZZLEID+"&clueId="+id+"&clueKind="+KIND
	var xhr = new XMLHttpRequest();
	xhr.open("GET", url, true);
	xhr.send()
}

function popItem() {
	id = pad(id)
	var url = "http://concept-game.cfapps.pez.pivotal.io/popItem?puzzleId="+PUZZLEID
	var xhr = new XMLHttpRequest();
	xhr.open("GET", url, true);
	xhr.send()
}

function pad(n) {
	p = '0';
	n = n + '';
	return n.length >= 3 ? n : new Array(3 - n.length + 1).join(p) + n;
}

function changeKind(n) {
	KIND = n
}

function icons(i) {
	id = pad(i)
	
	var icon_img = document.createElement("img");
	icon_img.setAttribute("class","IconImg");
	icon_img.setAttribute("src",id+".png");
	icon_img.setAttribute("onclick", "PushItem('"+id+"')");

	var icons_div = document.getElementById("Icons")
	icons_div.appendChild(icon_img)
}

function renderCreatePage(puzzleId) {
	PUZZLEID = puzzleId
	for (var i=1; i<=118; i++) {
		icons(i);
	}
	setInterval(function(){
		getConcept("0");
		getConcept("1");
		getConcept("2");
		getConcept("3");
	}, 700);
}

function renderWatchPage(puzzleId) {
	PUZZLEID = puzzleId
	setInterval(function(){
		getConcept("0");
		getConcept("1");
		getConcept("2");
		getConcept("3");
	}, 700);
}
