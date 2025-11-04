filterByTag("all")
function filterByTag(tag) {
  // Handle active button toggle
  var activeButtons = document.getElementsByClassName("btn-white");
  if (activeButtons.length > 0) {
    var b = document.getElementById(activeButtons[0].id);
    b.classList.toggle("btn-white");
    b.classList.toggle("btn-blue");
    if (b.id === tag) {
      tag = "all";
    }
  }

  // "all" means no filtering
  if (tag === "all") {
    tag = "";
  }

  var elements = document.getElementsByClassName("filterTag");

  for (var i = 0; i < elements.length; i++) {
    // Always start visible
    w3RemoveClass(elements[i], "hide");

    // Hide only if it doesn't match the tag
    if (tag !== "" && elements[i].className.indexOf(tag) === -1) {
      w3AddClass(elements[i], "hide");
    }
  }

  // Highlight active button
  if (tag !== "") {
    var b = document.getElementById(tag);
    if (b) {
      b.classList.toggle("btn-blue");
      b.classList.toggle("btn-white");
    }
  }
}

function w3AddClass(element, name) {
  var arr1 = element.className.split(" ");
  var arr2 = name.split(" ");
  for (var i = 0; i < arr2.length; i++) {
    if (arr1.indexOf(arr2[i]) == -1) {
      element.className += " " + arr2[i];
    }
  }
}

function w3RemoveClass(element, name) {
  var arr1 = element.className.split(" ");
  var arr2 = name.split(" ");
  for (var i = 0; i < arr2.length; i++) {
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
  x.classList.toggle("bg-myblueDarker");
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
      item.classList.add('bg-myblueDarker');
    }
  });
});

function copyCode(ahref, alertID) {
  const pre = ahref.closest('.relative').querySelector('pre');
  const code = pre.innerText;
  copyText(code, alertID);
}

async function copyText(text, alertID) {
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
  const tooltip = document.getElementById(alertID);
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


class Scroller {
  static updateElement(element, add) {
    if (add === true) {
      element.classList.add("bg-myblueDarker", "text-white", "hover:bg-myblueDarkest", "hover:!text-white")
    } else {
      element.classList.remove("bg-myblueDarker", "text-white", "hover:bg-myblueDarkest", "hover:!text-white");
    }
  }

  static updateHeaders(scrolling) {
    let activeIndex = this.headers.findIndex((header) => {
      return header.getBoundingClientRect().top > 180;
    });
    if (activeIndex == -1) {
      activeIndex = this.headers.length - 1;
    } else if (activeIndex > 0) {
      activeIndex--;
    }

    if (scrolling === true) {
      let active = this.headers[activeIndex];
      if (active !== this.activeHeader) {
        this.activeHeader = active;
        this.tocLinks.forEach(link => this.updateElement(link, false));
        this.updateElement(this.tocLinks[activeIndex], true);
      }
    } else {
      this.updateElement(this.tocLinks[activeIndex], true);
    }
  }

  static init() {
    if (document.querySelector('.tocLinks')) {
      this.tocLinks = document.querySelectorAll('.tocLinks a');
      this.tocLinks.forEach(link => link.classList.add('transition', 'duration-200'))
      this.headers = Array.from(this.tocLinks).map(link => {
        return document.querySelector(`#${link.href.split('#')[1]}`);
      })
      this.updateHeaders(false);
      this.ticking = false;
      window.addEventListener('scroll', (e) => {
        this.onScroll();
      })
    }
  }

  static onScroll() {
    if (!this.ticking) {
      requestAnimationFrame(this.update.bind(this));
      this.ticking = true;
    }
  }

  static update() {
    this.activeHeader ||= this.headers[0];
    this.updateHeaders(true);
    this.ticking = false;
  }
}

document.addEventListener('DOMContentLoaded', function (e) {
  Scroller.init();
})
