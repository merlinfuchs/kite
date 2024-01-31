import { EventHandler, addEventHandler } from "../sys";

export namespace event {
  export function on(event: string, handler: EventHandler) {
    addEventHandler(event, handler);
  }
}
