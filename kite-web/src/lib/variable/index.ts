export const variableTypes = [
  {
    name: "Text",
    value: "string",
    description: "A text variable can store a string of text.",
  },
  {
    name: "Number",
    value: "number",
    description: "A number variable can store any number, including decimals.",
  },
  {
    name: "Boolean",
    value: "boolean",
    description: "A boolean variable can store true or false.",
  },
];

export const variableScopes = [
  {
    name: "Global",
    value: "global",
    description: "A global variable stores one value across the whole app.",
  },
  {
    name: "Server",
    value: "guild",
    description: "A server variable stores one value per server.",
  },
  {
    name: "Channel",
    value: "channel",
    description: "A channel variable stores one value per channel.",
  },
  {
    name: "User",
    value: "user",
    description: "A user variable stores one value per user.",
  },
  {
    name: "Member",
    value: "member",
    description: "A member variable stores one value per member.",
  },
];
