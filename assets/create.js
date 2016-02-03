function getConcept(puzzleId, kind) {
	var xmlHttp = new XMLHttpRequest();
	xmlHttp.onreadystatechange = function() {
		if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
			updateConcept(JSON.parse(xmlHttp.responseText), kind);
	}
	var url = "http://localhost:8888/getConcept?puzzleId="+puzzleId+"&clueKind="+kind
	xmlHttp.open("GET", url, true); // true for asynchronous 
	xmlHttp.send(null);
}

function updateConcept(clues, kind) {
	var mainView = document.getElementById("MainViewSticky");
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
	if (old_outer_div != null) {
		mainView.replaceChild(outer_div,old_outer_div);	
	}
	mainView.appendChild(outer_div);
}

function PushItem(puzzleId,id) {
	var kind = 0
	id = pad(id)
	var url = "http://localhost:8888/pushItem?puzzleId="+puzzleId+"&clueId="+id+"&clueKind="+kind
	var xhr = new XMLHttpRequest();
	xhr.open("GET", url, true);
	xhr.send()
}

function pad(n) {
	p = '0';
	n = n + '';
	return n.length >= 3 ? n : new Array(3 - n.length + 1).join(p) + n;
}

function iconPair(puzzleId,i) {
	id1 = pad(i)
	id2 = pad(i+1)
	var outer_div = document.createElement("div");
	outer_div.setAttribute("class","IconPairContainer");

	var inner_div = document.createElement("div");
	inner_div.setAttribute("class","IconContainer");

	var icon_img = document.createElement("img");
	icon_img.setAttribute("class","Icon");
	icon_img.setAttribute("src",id1+".png");
	icon_img.setAttribute("onclick", "PushItem('"+puzzleId+"','"+id1+"')");

	inner_div.appendChild(icon_img)
	outer_div.appendChild(inner_div)

	inner_div = document.createElement("div");
	inner_div.setAttribute("class","IconContainer");

	icon_img = document.createElement("img");
	icon_img.setAttribute("class","Icon");
	icon_img.setAttribute("src",id2+".png");
	icon_img.setAttribute("onclick", "PushItem('"+puzzleId+"','"+id2+"')");

	inner_div.appendChild(icon_img)
	outer_div.appendChild(inner_div)
	var icons_div = document.getElementById("Icons")
	icons_div.appendChild(outer_div)
}

function renderPage(puzzleId) {
	for (var i=1; i<=118; i+=2) {
		iconPair(puzzleId,i);
	}
	setInterval(function(){
		conceptDiv = getConcept(puzzleId, "0");
		conceptDiv = getConcept(puzzleId, "1");
		conceptDiv = getConcept(puzzleId, "2");
	}, 700);
}
