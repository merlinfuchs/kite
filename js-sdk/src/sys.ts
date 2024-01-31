declare var Kite: {
  call: (call: { type: string; data: any }) => void;
  describe: () => any;
  handle: (event: { type: string; data: any }) => { success: boolean };
};

const manifest = {
  events: [],
};

Kite.describe = function () {
  return manifest;
};

export type EventHandler = (e: any) => void;

const eventHandlers: Record<string, EventHandler[]> = {};

Kite.handle = function (event) {
  if (event.type in eventHandlers) {
    try {
      eventHandlers[event.type].forEach((handler) => handler(event.data));
    } catch (e) {
      console.error(e);
      return { success: false };
    }
  }

  return { success: true };
};

export function addEventHandler(event: string, handler: EventHandler) {
  if (!manifest.events.includes(event)) {
    manifest.events.push(event);
  }

  eventHandlers[event] = eventHandlers[event] || [];
  eventHandlers[event].push(handler);
}

export function makeCall(type: string, data: any) {
  Kite.call({ type, data });
}
