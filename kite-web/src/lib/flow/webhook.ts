import { getApiUrl } from "@/lib/api/client";
import { EventListener } from "@/lib/types/wire.gen";

// The URL that runs a webhook event listener, or null for other listeners.
export function getWebhookUrl(listener: EventListener): string | null {
  const url = getWebhookApiUrl(listener);
  if (!url) return null;

  // The API base URL is empty when the service serves the web app itself,
  // but the URL is entered into other services, so it needs the origin.
  return new URL(url, window.location.origin).href;
}

// Like getWebhookUrl, but without the origin when the API base URL is empty,
// for where there is no window, like the MCP server.
export function getWebhookApiUrl(listener: EventListener): string | null {
  if (!listener.webhook_secret) return null;
  return getApiUrl(
    `/v1/webhooks/${listener.app_id}/${listener.id}/${listener.webhook_secret}`
  );
}
