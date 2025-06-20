filterByTag("all")
function filterByTag(tag) {
  // If a current button is white, you KNOW you have to make it blue again (for the functionality we want)
  var b = document.getElementsByClassName("btn-white");
  if (b.length > 0) {
    // Get it by id though, so it doesn't disappear from "btn-white" list when class is removed
    b = document.getElementById(b[0].id)
    b.classList.toggle("btn-white");
    b.classList.toggle("btn-blue");
    // If incoming button clicked was white button before, "show all" functionality
    if (b.id == tag) {
      tag = "all";
    }
    // Otherwise, we just let it fall through, so it is filtered accordingly
  }

  // If tag is "all", then match empty string to class name (always matches)
  // This is only relevant for the first time around
  if (tag == "all") {
    tag = "";
  }

  // Get all elements with the "filterTag" class
  var x = document.getElementsByClassName("filterTag");

  // Go through each element
  for (var i = 0; i < x.length; i++) {
    // We hide all elements first (by removing "show")
    w3RemoveClass(x[i], "show");
    // Then, only add "show" in if it has the tag
    if (x[i].className.indexOf(tag) > -1) {
      w3AddClass(x[i], "show");
    }
  }

  // If tag is not empty (from "all"), then make that button white lols
  if (tag != "") {
    var b = document.getElementById(tag);
    b.classList.toggle("btn-blue");
    b.classList.toggle("btn-white");
  }
}

function w3AddClass(element, name) {
  var i, arr1, arr2;
    arr1 = element.className.split(" ");
    arr2 = name.split(" ");
    for (i = 0; i < arr2.length; i++) {
    if (arr1.indexOf(arr2[i]) == -1) {element.className += " " + arr2[i];}
  }
}

function w3RemoveClass(element, name) {
  var i, arr1, arr2;
  arr1 = element.className.split(" ");
  arr2 = name.split(" ");
  for (i = 0; i < arr2.length; i++) {
    while (arr1.indexOf(arr2[i]) > -1) {
        arr1.splice(arr1.indexOf(arr2[i]), 1);
    }
  }
  element.className = arr1.join(" ");
}

function toggle(x) {
  x.classList.toggle("change");
  var dropdown = document.getElementById("drop-down");
  dropdown.classList.toggle("hidden");
}
