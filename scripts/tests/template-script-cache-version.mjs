import assert from "node:assert/strict";
import { readFileSync } from "node:fs";

const expected = [
  ["html/form-base.html", "js/form.js?v={{ .cacheVersion }}"],
  ["html/form-base.html", "js/pwr.js?v={{ .cacheVersion }}"],
  ["html/admin.html", "/js/admin.js?v={{ .cacheVersion }}"],
  ["html/user.html", "/js/user.js?v={{ .cacheVersion }}"],
  ["html/password-reset.html", "/js/pwr-pin.js?v={{ .cacheVersion }}"],
  ["html/setup.html", "js/setup.js?v={{ .cacheVersion }}"],
];

for (const [file, needle] of expected) {
  const content = readFileSync(file, "utf8");
  assert(
    content.includes(needle),
    `${file} should version entry script with ${needle}`,
  );
}
