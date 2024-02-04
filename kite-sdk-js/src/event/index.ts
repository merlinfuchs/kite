import { EventHandler, addEventHandler } from "../sys";
import { EventType } from "./types";

export namespace event {
  export function on(event: EventType, handler: EventHandler) {
    addEventHandler(event, handler);
  }
}
