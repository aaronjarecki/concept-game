function getConcept(puzzleId, kind) {
	var xmlHttp = new XMLHttpRequest();
	xmlHttp.onreadystatechange = function() {
		if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
			updateConcept(JSON.parse(xmlHttp.responseText), kind);
	}
	var url = "http://concept-game.cfapps.pez.pivotal.io/getConcept?puzzleId="+puzzleId+"&clueKind="+kind
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
		clue_img.setAttribute("src",clues[i].Id+".png");
		var concept_img = document.createElement('img');
		concept_img.setAttribute("class","ConceptIcon");
		concept_img.setAttribute('src', 'ico-square-'+kind+'.png');

		inner_div.appendChild(clue_img);
		inner_div.appendChild(concept_img);
		outer_div.appendChild(inner_div);
	}
	if (old_outer_div != null) {
		mainView.replaceChild(outer_div,old_outer_div);	
	}
	mainView.appendChild(outer_div);
}

function fillMainView(puzzleId) {
	setInterval(function(){
		conceptDiv = getConcept(puzzleId, "0");
		conceptDiv = getConcept(puzzleId, "1");
		conceptDiv = getConcept(puzzleId, "2");
	}, 700);
}