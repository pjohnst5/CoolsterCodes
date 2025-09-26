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

function animate_blue(x) {
  x.classList.toggle("bg-myblue");
}

document.addEventListener('DOMContentLoaded', function () {
  document.querySelectorAll('.nav_item').forEach(item => {
    const href = item.getAttribute('href');
    if (href === window.location.pathname) {
      item.classList.add('active');
    }
  });

  document.querySelectorAll('.mobile_nav').forEach(item => {
    const link = item.querySelector('a');
    const href = link.getAttribute('href');
    if (href === window.location.pathname) {
      item.classList.add('bg-myblue');
    }
  });
});


async function copyText(text) {
  try {
    if (navigator.clipboard && document.hasFocus()) {
      // Modern API
      await navigator.clipboard.writeText(text);
    } else {
      // Fallback for older browsers or unfocused document
      fallbackCopyText(text);
    }
  } catch (err) {
    console.warn('Clipboard API failed, using fallback:', err);
    fallbackCopyText(text);
  }
  // Show tooltip
  const tooltip = document.getElementById('copyalert');
  tooltip.classList.remove('opacity-0');
  tooltip.classList.remove('hidden');
  tooltip.classList.add('opacity-100');

  // Hide after 2 seconds
  setTimeout(() => {
    tooltip.classList.add('opacity-0');
    tooltip.classList.add('hidden');
    tooltip.classList.remove('opacity-100');
  }, 2000);
}

function fallbackCopyText(text) {
  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed'; // Avoid scrolling
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  document.execCommand('copy');
  document.body.removeChild(textarea);
}
