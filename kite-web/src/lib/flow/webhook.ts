import { getApiUrl } from "@/lib/api/client";
import { EventListener } from "@/lib/types/wire.gen";

// The URL that runs a webhook event listener, or null for other listeners.
export function getWebhookUrl(listener: EventListener): string | null {
  if (!listener.webhook_secret) return null;

  // The API base URL is empty when the service serves the web app itself,
  // but the URL is entered into other services, so it needs the origin.
  return new URL(
    getApiUrl(
      `/v1/apps/${listener.app_id}/webhooks/${listener.id}/${listener.webhook_secret}`
    ),
    window.location.origin
  ).href;
}
