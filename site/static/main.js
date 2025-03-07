import { autocomp } from "./lib.js";

let TAGS = []
if (!localStorage.tags) {
    fetch("/api/tags")
    .then(response => response.json())
    .then(data => {
        TAGS = data.data;
        localStorage.tags = TAGS.join("|");
    });
} else {
    TAGS = localStorage.tags.split("|");
}

const qInput = document.querySelector("input[data-autocomp-tags]");
const isTags = document.querySelector(".search input[name=field][value=tags]");
if (qInput) {
    autocomp(qInput, {
        onQuery: async (val) => {
            if (!isTags.checked) {
                return [];
            }
            const q = val.trim().toLowerCase();
            return TAGS.filter(s => s.includes(q)).slice(0, 10);
        },

        onSelect: (val) => {
            return val;
        }
    });
}

// Listen for ~ key and focus on the search bar.
document.addEventListener("keydown", function(event) {
  if (event.key === "`") {
    event.preventDefault();
    const q = document.querySelector("form.search input[name=q]");
    q.focus();
    q.select();
  }
});
