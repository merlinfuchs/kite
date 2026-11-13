import queryClient, { apiRequest } from "@/lib/api/client";
import { itemPath, type ItemVersionTarget } from "@/lib/api/queries";
import type { EditorSessionUpdateResponse } from "@/lib/types/wire.gen";
import { useEffect } from "react";

// How long after an input the editor still counts as in use.
const idleAfterMs = 2 * 60_000;
const reportIntervalMs = 60_000;

// useEditorTimeReport reports how long the flow editor is in use. Time counts
// while the tab is visible and up to two minutes after each input. Nothing is
// reported until the first input, so opening a flow to look at it isn't a
// session. Saves are counted by the server.
export function useEditorTimeReport(
  target: ItemVersionTarget,
  sessionId: string
) {
  const { appId, type, id } = target;

  useEffect(() => {
    if (type === "message") return;
    const path = `${itemPath({
      appId,
      type,
      id,
    })}/editor-sessions/${sessionId}`;

    // The time of earlier stretches of input, and of the current one if any.
    let pastMs = 0;
    let activeSince: number | null = null;
    let activeUntil = 0;
    let started = false;
    let reportedSeconds = -1;

    function endStretch(now: number) {
      if (activeSince === null) return;
      pastMs += Math.min(now, activeUntil) - activeSince;
      activeSince = null;
    }

    function activeSeconds() {
      const now = performance.now();
      if (now >= activeUntil) endStretch(now);
      const currentMs = activeSince === null ? 0 : now - activeSince;
      return Math.floor((pastMs + currentMs) / 1000);
    }

    async function report(keepalive = false) {
      if (!started) return;
      const seconds = activeSeconds();
      if (seconds <= reportedSeconds) return;

      // The server keeps the highest time, so the order of the requests
      // doesn't matter.
      try {
        const res = await apiRequest<EditorSessionUpdateResponse>(path, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ active_seconds: seconds }),
          keepalive,
        });
        if (res.success) reportedSeconds = Math.max(reportedSeconds, seconds);
      } catch {
        // Tried again with the next report.
      }
    }

    function onInput() {
      const now = performance.now();
      if (now >= activeUntil) endStretch(now);
      activeSince ??= now;
      activeUntil = now + idleAfterMs;
      if (!started) {
        started = true;
        void report();
      }
    }

    function onVisibilityChange() {
      if (document.visibilityState === "hidden") {
        endStretch(performance.now());
        void report(true);
      }
    }

    const onPageHide = () => void report(true);
    const inputEvents = ["pointerdown", "keydown", "wheel"] as const;
    const listenerOptions = { capture: true, passive: true };

    for (const event of inputEvents) {
      window.addEventListener(event, onInput, listenerOptions);
    }
    document.addEventListener("visibilitychange", onVisibilityChange);
    window.addEventListener("pagehide", onPageHide);
    const interval = window.setInterval(() => void report(), reportIntervalMs);

    return () => {
      for (const event of inputEvents) {
        window.removeEventListener(event, onInput, listenerOptions);
      }
      document.removeEventListener("visibilitychange", onVisibilityChange);
      window.removeEventListener("pagehide", onPageHide);
      window.clearInterval(interval);
      if (!started) return;
      // So the lists show this session once it's reported.
      void report(true).then(() =>
        queryClient.invalidateQueries({
          queryKey: ["apps", appId, "editor-usage"],
        })
      );
    };
  }, [appId, type, id, sessionId]);
}
