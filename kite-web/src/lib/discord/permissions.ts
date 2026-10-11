// From https://discord.com/developers/docs/topics/permissions. The categories
// group them for the permissions calculator.
export const permissionBits = [
  {
    label: "Create Instant Invite",
    bit: 0,
    category: "General",
  },
  {
    label: "Kick Members",
    bit: 1,
    category: "Members",
  },
  {
    label: "Ban Members",
    bit: 2,
    category: "Members",
  },
  {
    label: "Administrator",
    bit: 3,
    category: "Advanced",
  },
  {
    label: "Manage Channels",
    bit: 4,
    category: "General",
  },
  {
    label: "Manage Guild",
    bit: 5,
    category: "General",
  },
  {
    label: "Add Reactions",
    bit: 6,
    category: "Text",
  },
  {
    label: "View Audit Log",
    bit: 7,
    category: "General",
  },
  {
    label: "Priority Speaker",
    bit: 8,
    category: "Voice",
  },
  {
    label: "Stream",
    bit: 9,
    category: "Voice",
  },
  {
    label: "View Channel",
    bit: 10,
    category: "General",
  },
  {
    label: "Send Messages",
    bit: 11,
    category: "Text",
  },
  {
    label: "Send TTS Messages",
    bit: 12,
    category: "Text",
  },
  {
    label: "Manage Messages",
    bit: 13,
    category: "Text",
  },
  {
    label: "Embed Links",
    bit: 14,
    category: "Text",
  },
  {
    label: "Attach Files",
    bit: 15,
    category: "Text",
  },
  {
    label: "Read Message History",
    bit: 16,
    category: "Text",
  },
  {
    label: "Mention Everyone",
    bit: 17,
    category: "Text",
  },
  {
    label: "Use External Emojis",
    bit: 18,
    category: "Text",
  },
  {
    label: "View Guild Insights",
    bit: 19,
    category: "General",
  },
  {
    label: "Connect",
    bit: 20,
    category: "Voice",
  },
  {
    label: "Speak",
    bit: 21,
    category: "Voice",
  },
  {
    label: "Mute Members",
    bit: 22,
    category: "Voice",
  },
  {
    label: "Deafen Members",
    bit: 23,
    category: "Voice",
  },
  {
    label: "Move Members",
    bit: 24,
    category: "Voice",
  },
  {
    label: "Use VAD",
    bit: 25,
    category: "Voice",
  },
  {
    label: "Change Nickname",
    bit: 26,
    category: "Members",
  },
  {
    label: "Manage Nicknames",
    bit: 27,
    category: "Members",
  },
  {
    label: "Manage Roles",
    bit: 28,
    category: "General",
  },
  {
    label: "Manage Webhooks",
    bit: 29,
    category: "General",
  },
  {
    label: "Manage Expressions",
    bit: 30,
    category: "General",
  },
  {
    label: "Use Slash Commands",
    bit: 31,
    category: "Text",
  },
  {
    label: "Request to Speak",
    bit: 32,
    category: "Voice",
  },
  {
    label: "Manage Events",
    bit: 33,
    category: "General",
  },
  {
    label: "Manage Threads",
    bit: 34,
    category: "Text",
  },
  {
    label: "Create Public Threads",
    bit: 35,
    category: "Text",
  },
  {
    label: "Create Private Threads",
    bit: 36,
    category: "Text",
  },
  {
    label: "Use External Stickers",
    bit: 37,
    category: "Text",
  },
  {
    label: "Send Messages in Threads",
    bit: 38,
    category: "Text",
  },
  {
    label: "Use Embedded Activities",
    bit: 39,
    category: "Voice",
  },
  {
    label: "Moderate Members",
    bit: 40,
    category: "Members",
  },
  {
    label: "View Creator Monetization Analytics",
    bit: 41,
    category: "General",
  },
  {
    label: "Use Soundboard",
    bit: 42,
    category: "Voice",
  },
  {
    label: "Create Guild Expressions",
    bit: 43,
    category: "General",
  },
  {
    label: "Create Events",
    bit: 44,
    category: "General",
  },
  {
    label: "Use External Sounds",
    bit: 45,
    category: "Voice",
  },
  {
    label: "Send Voice Messages",
    bit: 46,
    category: "Text",
  },
  {
    label: "Set Voice Channel Status",
    bit: 48,
    category: "Voice",
  },
  {
    label: "Send Polls",
    bit: 49,
    category: "Text",
  },
  {
    label: "Use External Apps",
    bit: 50,
    category: "Text",
  },
  {
    label: "Pin Messages",
    bit: 51,
    category: "Text",
  },
  {
    label: "Bypass Slowmode",
    bit: 52,
    category: "Text",
  },
] as const;

export type Permission = (typeof permissionBits)[number];

export type PermissionCategory = Permission["category"];

export type PermissionBit = Permission["bit"];

export function decodePermissionsBitset(value: string): Permission[] {
  let bits = BigInt(0);
  try {
    bits = BigInt(value);
  } catch (e) {
    return [];
  }

  const permissions: Permission[] = [];

  for (const permission of permissionBits) {
    if (bits & (BigInt(1) << BigInt(permission.bit))) {
      permissions.push(permission);
    }
  }

  return permissions;
}

export function encodePermissionsBitset(permissions: number[]): string {
  let bits = BigInt(0);

  for (const permission of permissions) {
    bits |= BigInt(1) << BigInt(permission);
  }

  return bits.toString();
}
