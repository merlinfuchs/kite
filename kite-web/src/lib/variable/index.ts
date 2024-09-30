export const variableTypes = [
  {
    name: "Text",
    value: "string",
    description: "A text variable can store a string of text.",
  },
  {
    name: "Number",
    value: "float",
    description: "A number variable can store any number, including decimals.",
  },
  {
    name: "Whole Number",
    value: "integer",
    description: "A whole number variable can store whole numbers.",
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
    description:
      "A user variable stores one value per user across all servers.",
  },
  {
    name: "Member",
    value: "member",
    description: "A member variable stores one value per user and server.",
  },
];

export function getVariableTypeName(type: string) {
  return variableTypes.find((t) => t.value === type)?.name || "Unknown";
}

export function getVariableTypeDescription(type: string) {
  return variableTypes.find((t) => t.value === type)?.description || "Unknown";
}

export function getVariableScopeName(scope: string) {
  return variableScopes.find((t) => t.value === scope)?.name || "Unknown";
}

export function getVariableScopeDescription(scope: string) {
  return (
    variableScopes.find((t) => t.value === scope)?.description || "Unknown"
  );
}
