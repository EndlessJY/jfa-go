import assert from "node:assert/strict";

const modulePath = process.argv[2];
if (!modulePath) {
  throw new Error("Usage: node scripts/tests/common-auth-refresh.mjs <bundled-common-module>");
}

class FakeXMLHttpRequest {
  static responses = [];
  static requests = [];

  method = "";
  url = "";
  async = true;
  headers = {};
  readyState = 0;
  status = 0;
  responseType = "";
  response = null;
  onreadystatechange = null;

  open(method, url, async = true) {
    this.method = method;
    this.url = url;
    this.async = async;
  }

  setRequestHeader(key, value) {
    this.headers[key] = value;
  }

  send(body) {
    FakeXMLHttpRequest.requests.push({
      method: this.method,
      url: this.url,
      headers: { ...this.headers },
      body,
    });

    const next = FakeXMLHttpRequest.responses.shift();
    if (!next) {
      throw new Error(`No mocked response for ${this.method} ${this.url}`);
    }

    setTimeout(() => {
      this.status = next.status;
      this.response = next.response;
      this.readyState = 4;
      this.onreadystatechange?.();
    }, 0);
  }
}

globalThis.XMLHttpRequest = FakeXMLHttpRequest;
globalThis.window = {
  token: "expired-token",
  pages: {
    Base: "",
  },
  lang: {
    notif: (key) => key,
  },
  notifications: {
    errors: [],
    connectionError() {
      this.errors.push("connection");
    },
    customError(type, message) {
      this.errors.push({ type, message });
    },
  },
};

const { _get } = await import(modulePath);

FakeXMLHttpRequest.responses.push(
  { status: 401, response: { error: "Unauthorized" } },
  { status: 200, response: { token: "fresh-token" } },
  { status: 200, response: { ok: true } },
);

const seen = [];
_get("/my/details", null, (req) => {
  if (req.readyState === 4) {
    seen.push({ status: req.status, response: req.response });
  }
});

await new Promise((resolve) => setTimeout(resolve, 30));

assert.deepEqual(
  FakeXMLHttpRequest.requests.map((req) => [req.method, req.url, req.headers.Authorization]),
  [
    ["GET", "/my/details", "Bearer expired-token"],
    ["GET", "/my/token/refresh", undefined],
    ["GET", "/my/details", "Bearer fresh-token"],
  ],
);
assert.equal(globalThis.window.token, "fresh-token");
assert.deepEqual(seen, [{ status: 200, response: { ok: true } }]);
assert.deepEqual(globalThis.window.notifications.errors, []);
