import { useSyncExternalStore } from "react";
import { toast } from "sonner";
import { track } from "../analytics";

// The event Chrome and Edge fire when the dashboard can be installed. It isn't
// in TypeScript's DOM types yet.
interface InstallPromptEvent extends Event {
  prompt(): Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

let installPrompt: InstallPromptEvent | null = null;
const listeners = new Set<() => void>();

function setInstallPrompt(event: InstallPromptEvent | null) {
  installPrompt = event;
  listeners.forEach((listener) => listener());
}

// listenForInstallPrompt keeps the browser's install prompt to show it from
// the menu later. Browsers fire it once, soon after the first page loads, so
// it has to listen before any page renders.
export function listenForInstallPrompt() {
  window.addEventListener("beforeinstallprompt", (event) => {
    event.preventDefault();
    setInstallPrompt(event as InstallPromptEvent);
  });
  window.addEventListener("appinstalled", () => setInstallPrompt(null));
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

type InstallMode = "prompt" | "share_menu" | null;

function installMode(): InstallMode {
  if (window.matchMedia("(display-mode: standalone)").matches) return null;
  if (installPrompt) return "prompt";
  // Only Safari on iPhone and iPad has navigator.standalone. It has no prompt
  // and installs from the share menu.
  const nav = navigator as Navigator & { standalone?: boolean };
  if (nav.standalone === false) return "share_menu";
  return null;
}

// useInstallApp is how the dashboard can be installed in this browser, null
// when it already runs installed or the browser can't install it.
export function useInstallApp() {
  return useSyncExternalStore(subscribe, installMode, () => null);
}

export async function installApp() {
  if (!installPrompt) {
    track("dashboard_install_clicked", { prompt: false });
    toast.info("Tap the share button, then Add to Home Screen.");
    return;
  }

  // A prompt can only be shown once, the browser fires a new one later.
  const prompt = installPrompt;
  setInstallPrompt(null);
  await prompt.prompt();
  const { outcome } = await prompt.userChoice;
  track("dashboard_install_clicked", { prompt: true, outcome });
}
