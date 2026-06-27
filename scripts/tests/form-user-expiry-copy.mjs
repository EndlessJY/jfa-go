import assert from "node:assert/strict";

const modulePath = process.argv[2];
if (!modulePath) {
  throw new Error("Usage: node scripts/tests/form-user-expiry-copy.mjs <bundled-form-expiry-module>");
}

const { userExpiryDisplayForInviteMode } = await import(modulePath);

const baseOptions = {
  months: 0,
  days: 7,
  hours: 0,
  minutes: 0,
  now: new Date("2026-06-27T00:00:00Z"),
  formatDate: (date) => date.toISOString().slice(0, 10),
  userExpiryMessage: "Your account will be valid until {date}.",
  userExpiryRegisterMessage: "Your new account will be valid until {date}.",
  userExpiryRenewalMessage: "This invite will extend the existing account after renewal.",
};

assert.deepEqual(userExpiryDisplayForInviteMode(null, baseOptions), {
  visible: false,
  text: "",
});

assert.deepEqual(userExpiryDisplayForInviteMode("register", baseOptions), {
  visible: true,
  text: "Your new account will be valid until 2026-07-04.",
});

assert.deepEqual(userExpiryDisplayForInviteMode("register", {
  ...baseOptions,
  userExpiryRegisterMessage: "",
}), {
  visible: true,
  text: "Your account will be valid until 2026-07-04.",
});

assert.deepEqual(userExpiryDisplayForInviteMode("renew", baseOptions), {
  visible: true,
  text: "This invite will extend the existing account after renewal.",
});
