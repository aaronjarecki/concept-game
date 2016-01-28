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

function makeIcons(puzzleId) {
	for (var i=1; i<=118; i+=2) {
		iconPair(puzzleId,i);
	}
}
