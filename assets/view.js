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
		clue_img.setAttribute("class","IconImg");
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

function fillMainView(puzzleId) {
	PUZZLEID = puzzleId
	conceptDiv = getConcept("0");
	conceptDiv = getConcept("1");
	conceptDiv = getConcept("2");
	conceptDiv = getConcept("3");
}