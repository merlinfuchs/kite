export const permissionBits = [
  {
    label: "Create Instant Invite",
    bit: 0,
  },
  {
    label: "Kick Members",
    bit: 1,
  },
  {
    label: "Ban Members",
    bit: 2,
  },
  {
    label: "Administrator",
    bit: 3,
  },
  {
    label: "Manage Channels",
    bit: 4,
  },
  {
    label: "Manage Guild",
    bit: 5,
  },
  {
    label: "Add Reactions",
    bit: 6,
  },
  {
    label: "View Audit Log",
    bit: 7,
  },
  {
    label: "Priority Speaker",
    bit: 8,
  },
  {
    label: "Stream",
    bit: 9,
  },
  {
    label: "View Channel",
    bit: 10,
  },
  {
    label: "Send Messages",
    bit: 11,
  },
  {
    label: "Send TTS Messages",
    bit: 12,
  },
  {
    label: "Manage Messages",
    bit: 13,
  },
  {
    label: "Embed Links",
    bit: 14,
  },
  {
    label: "Attach Files",
    bit: 15,
  },
  {
    label: "Read Message History",
    bit: 16,
  },
  {
    label: "Mention Everyone",
    bit: 17,
  },
  {
    label: "Use External Emojis",
    bit: 18,
  },
  {
    label: "View Guild Insights",
    bit: 19,
  },
  {
    label: "Connect",
    bit: 20,
  },
  {
    label: "Speak",
    bit: 21,
  },
  {
    label: "Mute Members",
    bit: 22,
  },
  {
    label: "Deafen Members",
    bit: 23,
  },
  {
    label: "Move Members",
    bit: 24,
  },
  {
    label: "Use VAD",
    bit: 25,
  },
  {
    label: "Change Nickname",
    bit: 26,
  },
  {
    label: "Manage Nicknames",
    bit: 27,
  },
  {
    label: "Manage Roles",
    bit: 28,
  },
  {
    label: "Manage Webhooks",
    bit: 29,
  },
  {
    label: "Manage Emojis",
    bit: 30,
  },
  {
    label: "Use Slash Commands",
    bit: 31,
  },
  {
    label: "Request to Speak",
    bit: 32,
  },
  {
    label: "Manage Threads",
    bit: 34,
  },
  {
    label: "Create Public Threads",
    bit: 35,
  },
  {
    label: "Create Private Threads",
    bit: 36,
  },
  {
    label: "Use External Stickers",
    bit: 37,
  },
  {
    label: "Send Messages in Threads",
    bit: 38,
  },
  {
    label: "Use Embedded Activities",
    bit: 39,
  },
  {
    label: "Moderate Members",
    bit: 40,
  },
  {
    label: "View Creator Monetization Analytics",
    bit: 41,
  },
  {
    label: "Use Soundboard",
    bit: 42,
  },
  {
    label: "Create Guild Expressions",
    bit: 43,
  },
  {
    label: "Create Events",
    bit: 44,
  },
  {
    label: "Use External Sounds",
    bit: 45,
  },
  {
    label: "Send Voice Messages",
    bit: 46,
  },
  {
    label: "Send Polls",
    bit: 49,
  },
  {
    label: "Use External Apps",
    bit: 50,
  },
] as const;

export type Permission = (typeof permissionBits)[number];

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
