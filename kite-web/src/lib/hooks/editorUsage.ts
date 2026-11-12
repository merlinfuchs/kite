import { apiRequest } from "@/lib/api/client";
import type { ItemVersionTarget } from "@/lib/api/queries";
import { useEffect, useRef } from "react";

export function useEditorUsageReport(
  target: ItemVersionTarget,
  sessionId: string,
  saveCount: number
) {
  const { appId, type, id } = target;
  const saveCountRef = useRef(saveCount);
  const previousSaveCount = useRef(saveCount);
  const flushRef = useRef<(() => void) | null>(null);
  saveCountRef.current = saveCount;

  useEffect(() => {
    if (!appId || type === "message") return;

    let activeMilliseconds = 0;
    let checkpointAt = performance.now();
    let wasVisible = document.visibilityState === "visible";
    let sending = false;
    let sendAgain = false;
    let sendAgainKeepalive = false;

    function checkpoint() {
      const now = performance.now();
      if (wasVisible) activeMilliseconds += now - checkpointAt;
      checkpointAt = now;
      wasVisible = document.visibilityState === "visible";
    }

    async function send(keepalive = false) {
      checkpoint();
      if (sending) {
        sendAgain = true;
        sendAgainKeepalive ||= keepalive;
        return;
      }

      sending = true;
      try {
        const response = await apiRequest<{ saved: boolean }>(
          `/v1/apps/${encodeURIComponent(
            appId
          )}/editor-sessions/${encodeURIComponent(sessionId)}`,
          {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              target_type: type,
              target_id: id,
              active_seconds: Math.floor(activeMilliseconds / 1000),
              save_count: saveCountRef.current,
            }),
            keepalive,
          }
        );
        if (!response.success) {
          console.error(
            "Failed to record editor usage:",
            response.error.message
          );
        }
      } catch (error) {
        console.error("Failed to record editor usage:", error);
      } finally {
        sending = false;
        if (sendAgain) {
          sendAgain = false;
          const keepaliveAgain = sendAgainKeepalive;
          sendAgainKeepalive = false;
          void send(keepaliveAgain);
        }
      }
    }

    const flush = () => void send();
    const onVisibilityChange = () => {
      checkpoint();
      void send(document.visibilityState === "hidden");
    };
    const onPageHide = () => void send(true);

    flushRef.current = flush;
    document.addEventListener("visibilitychange", onVisibilityChange);
    window.addEventListener("pagehide", onPageHide);
    const interval = window.setInterval(() => {
      if (document.visibilityState === "visible") void send();
    }, 60_000);
    void send();

    return () => {
      window.clearInterval(interval);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      window.removeEventListener("pagehide", onPageHide);
      if (flushRef.current === flush) flushRef.current = null;
      void send(true);
    };
  }, [appId, id, sessionId, type]);

  useEffect(() => {
    if (previousSaveCount.current === saveCount) return;
    previousSaveCount.current = saveCount;
    flushRef.current?.();
  }, [saveCount]);
}
