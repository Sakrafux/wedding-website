import "@testing-library/jest-dom/vitest";

import { cleanup } from "@testing-library/react";
import { afterEach, expect, vi } from "vitest";

// jsdom has no layout, so it cannot scroll. TanStack Router restores scroll on
// navigation, and the unimplemented-method warnings it produces would bury a real
// failure in the output.
window.scrollTo = () => {};

// Radix's popover and radio-group measure their trigger, and jsdom ships no
// ResizeObserver. A no-op is enough: nothing here asserts on position, and the
// alternative is every component that uses a Radix primitive failing with
// "ResizeObserver is not defined" instead of with whatever it is actually about.
globalThis.ResizeObserver ??= class {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// React Testing Library does not unmount between tests on its own when globals are
// enabled through a setup file, and a left-over tree makes the next test's queries
// ambiguous in a way that reads like a component bug.
afterEach(() => {
  cleanup();
  // The fetch stub, and the confirmed-households list the confirmation screen
  // persists. Both outlive a test otherwise, and the symptom — a screen that does
  // not appear because a previous test confirmed the same household — is a long
  // way from the cause.
  vi.unstubAllGlobals();
  window.localStorage.clear();
});

// Keeps the matcher types attached when a test file imports nothing from vitest.
export { expect };
