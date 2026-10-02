import { getApiUrl } from "@/lib/api/client";
import { EventListener } from "@/lib/types/wire.gen";

// The URL that runs a webhook event listener, or null for other listeners.
export function getWebhookUrl(listener: EventListener): string | null {
  if (!listener.webhook_secret) return null;

  return absoluteUrl(
    getApiUrl(
      `/v1/apps/${listener.app_id}/webhooks/${listener.id}/${listener.webhook_secret}`
    )
  );
}

// The API base URL is empty when the service serves the web app itself, which
// makes API URLs relative. A webhook URL is entered into other services, so
// it needs the origin.
export function absoluteUrl(url: string, origin?: string): string {
  if (/^https?:\/\//.test(url)) return url;

  origin ??= typeof window !== "undefined" ? window.location.origin : "";
  return origin + url;
}
