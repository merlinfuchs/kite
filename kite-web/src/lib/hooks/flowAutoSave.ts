import { useCallback, useEffect, useState } from "react";

const storagePrefix = "kite:flow-auto-save:";

// Whether the flow editor saves on its own. It's set per flow, e.g.
// "command:<id>", and kept in the browser. Without a key it's only kept until
// the page is left.
export function useFlowAutoSave(
  key?: string
): [boolean, (enabled: boolean) => void] {
  const [enabled, setEnabled] = useState(false);

  // Read after mounting, the server render has no localStorage.
  useEffect(() => {
    if (!key) {
      setEnabled(false);
      return;
    }

    try {
      setEnabled(localStorage.getItem(storagePrefix + key) === "true");
    } catch {
      // Storage can be unavailable, e.g. when blocked by the browser.
      setEnabled(false);
    }
  }, [key]);

  const update = useCallback(
    (value: boolean) => {
      setEnabled(value);
      if (!key) return;

      try {
        if (value) {
          localStorage.setItem(storagePrefix + key, "true");
        } else {
          localStorage.removeItem(storagePrefix + key);
        }
      } catch {}
    },
    [key]
  );

  return [enabled, update];
}
