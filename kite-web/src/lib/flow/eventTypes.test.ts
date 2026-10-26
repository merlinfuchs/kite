import { describe, expect, it } from "vitest";
import { discordEntryEvent } from "../blocks/discordEntryEvent";
import {
  discordEventTypes,
  eventSourceName,
  eventSourceOfFlow,
} from "./eventTypes";

describe("discordEventTypes", () => {
  it("lists every Discord event of the entry block once", () => {
    const field = discordEntryEvent.fields?.find(
      (f) => f.name === "event_type"
    );
    const options = (field?.schema as unknown as { options: string[] }).options;
    const discordEvents = options.filter(
      (o) => o !== "cron" && o !== "webhook"
    );

    const values = discordEventTypes.map((t) => t.value);
    expect(new Set(values).size).toBe(values.length);
    expect([...values].sort()).toEqual([...discordEvents].sort());
  });

  it("links each source to its own page", () => {
    expect(eventSourceName("discord").path).toBe("/apps/[appId]/events");
    expect(eventSourceName("schedule").path).toBe("/apps/[appId]/schedules");
    expect(eventSourceName("webhook").path).toBe("/apps/[appId]/webhooks");
  });

  it("names each source", () => {
    expect(eventSourceName("discord").title).toBe("Event Listener");
    expect(eventSourceName("schedule").singular).toBe("scheduled task");
    expect(eventSourceName("webhook").title).toBe("Webhook");
  });

  it("derives the source from the entry block like the service", () => {
    const flow = (eventType: string) => ({
      nodes: [{ type: "entry_event", data: { event_type: eventType } }],
    });
    expect(eventSourceOfFlow(flow("cron"))).toBe("schedule");
    expect(eventSourceOfFlow(flow("webhook"))).toBe("webhook");
    expect(eventSourceOfFlow(flow("message_create"))).toBe("discord");
  });
});
