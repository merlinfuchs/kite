import { EventHandler, addEventHandler, makeCall } from "./sys";

export function call(type: string, data: any) {
  makeCall(type, data);
}

export { discord } from "./discord/index";
export { event } from "./event/index";
