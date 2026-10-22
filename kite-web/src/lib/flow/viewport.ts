import { Viewport } from "@xyflow/react";

const storageKey = "kite-flow-viewports";
const maxViewports = 30;

function readViewports(): Record<string, Viewport> {
  try {
    return JSON.parse(localStorage.getItem(storageKey) || "{}");
  } catch {
    return {};
  }
}

export function getSavedViewport(key: string): Viewport | undefined {
  return readViewports()[key];
}

// Keeps the viewports of the most recently moved flows.
export function saveViewport(key: string, viewport: Viewport) {
  const viewports = readViewports();
  delete viewports[key];
  viewports[key] = viewport;

  const keys = Object.keys(viewports);
  for (const k of keys.slice(0, keys.length - maxViewports)) {
    delete viewports[k];
  }

  try {
    localStorage.setItem(storageKey, JSON.stringify(viewports));
  } catch {
    // Storage can be full or disabled.
  }
}
