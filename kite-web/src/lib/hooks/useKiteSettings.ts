import { useCallback, useEffect, useState } from "react";
import { defaultKiteSettings, KiteSettings } from "@/lib/settings/types";

const STORAGE_KEY = "kite:settings";

export function useKiteSettings() {
  const [settings, setSettings] = useState<KiteSettings>(defaultKiteSettings);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        setSettings({ ...defaultKiteSettings, ...JSON.parse(stored) });
      }
    } catch {}
    setIsLoaded(true);

    const onStorage = (e: StorageEvent) => {
      if (e.key === STORAGE_KEY && e.newValue) {
        try {
          setSettings({ ...defaultKiteSettings, ...JSON.parse(e.newValue) });
        } catch {}
      }
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);

  const updateSettings = useCallback((newSettings: Partial<KiteSettings>) => {
    setSettings((prev) => {
      const updated = { ...prev, ...newSettings };
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
        window.dispatchEvent(
          new StorageEvent("storage", {
            key: STORAGE_KEY,
            newValue: JSON.stringify(updated),
          })
        );
      } catch {}
      return updated;
    });
  }, []);

  const resetSettings = useCallback(() => {
    setSettings(defaultKiteSettings);
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(defaultKiteSettings));
      window.dispatchEvent(
        new StorageEvent("storage", {
          key: STORAGE_KEY,
          newValue: JSON.stringify(defaultKiteSettings),
        })
      );
    } catch {}
  }, []);

  return {
    settings,
    updateSettings,
    resetSettings,
    isLoaded,
  };
}
