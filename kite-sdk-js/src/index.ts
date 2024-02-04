import { Call, CallConfig, CallType } from "./call/types";
import { makeCall } from "./sys";

export function call(type: CallType, data: Call["data"], config?: CallConfig) {
  makeCall(type, data);
}

export { discord } from "./discord/index";
export { event } from "./event/index";
