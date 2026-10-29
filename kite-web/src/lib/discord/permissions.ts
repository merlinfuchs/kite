// From https://discord.com/developers/docs/topics/permissions. The flag is
// Discord's name for the permission and the label the one its server settings
// show. The categories group them for the permissions calculator.
export const permissionBits = [
  {
    flag: "CREATE_INSTANT_INVITE",
    label: "Create Invite",
    bit: 0,
    category: "General",
  },
  {
    flag: "KICK_MEMBERS",
    label: "Kick Members",
    bit: 1,
    category: "Members",
  },
  {
    flag: "BAN_MEMBERS",
    label: "Ban Members",
    bit: 2,
    category: "Members",
  },
  {
    flag: "ADMINISTRATOR",
    label: "Administrator",
    bit: 3,
    category: "Advanced",
  },
  {
    flag: "MANAGE_CHANNELS",
    label: "Manage Channels",
    bit: 4,
    category: "General",
  },
  {
    flag: "MANAGE_GUILD",
    label: "Manage Server",
    bit: 5,
    category: "General",
  },
  {
    flag: "ADD_REACTIONS",
    label: "Add Reactions",
    bit: 6,
    category: "Text",
  },
  {
    flag: "VIEW_AUDIT_LOG",
    label: "View Audit Log",
    bit: 7,
    category: "General",
  },
  {
    flag: "PRIORITY_SPEAKER",
    label: "Priority Speaker",
    bit: 8,
    category: "Voice",
  },
  {
    flag: "STREAM",
    label: "Video",
    bit: 9,
    category: "Voice",
  },
  {
    flag: "VIEW_CHANNEL",
    label: "View Channel",
    bit: 10,
    category: "General",
  },
  {
    flag: "SEND_MESSAGES",
    label: "Send Messages",
    bit: 11,
    category: "Text",
  },
  {
    flag: "SEND_TTS_MESSAGES",
    label: "Send Text-to-Speech Messages",
    bit: 12,
    category: "Text",
  },
  {
    flag: "MANAGE_MESSAGES",
    label: "Manage Messages",
    bit: 13,
    category: "Text",
  },
  {
    flag: "EMBED_LINKS",
    label: "Embed Links",
    bit: 14,
    category: "Text",
  },
  {
    flag: "ATTACH_FILES",
    label: "Attach Files",
    bit: 15,
    category: "Text",
  },
  {
    flag: "READ_MESSAGE_HISTORY",
    label: "Read Message History",
    bit: 16,
    category: "Text",
  },
  {
    flag: "MENTION_EVERYONE",
    label: "Mention Everyone",
    bit: 17,
    category: "Text",
  },
  {
    flag: "USE_EXTERNAL_EMOJIS",
    label: "Use External Emojis",
    bit: 18,
    category: "Text",
  },
  {
    flag: "VIEW_GUILD_INSIGHTS",
    label: "View Server Insights",
    bit: 19,
    category: "General",
  },
  {
    flag: "CONNECT",
    label: "Connect",
    bit: 20,
    category: "Voice",
  },
  {
    flag: "SPEAK",
    label: "Speak",
    bit: 21,
    category: "Voice",
  },
  {
    flag: "MUTE_MEMBERS",
    label: "Mute Members",
    bit: 22,
    category: "Voice",
  },
  {
    flag: "DEAFEN_MEMBERS",
    label: "Deafen Members",
    bit: 23,
    category: "Voice",
  },
  {
    flag: "MOVE_MEMBERS",
    label: "Move Members",
    bit: 24,
    category: "Voice",
  },
  {
    flag: "USE_VAD",
    label: "Use Voice Activity",
    bit: 25,
    category: "Voice",
  },
  {
    flag: "CHANGE_NICKNAME",
    label: "Change Nickname",
    bit: 26,
    category: "Members",
  },
  {
    flag: "MANAGE_NICKNAMES",
    label: "Manage Nicknames",
    bit: 27,
    category: "Members",
  },
  {
    flag: "MANAGE_ROLES",
    label: "Manage Roles",
    bit: 28,
    category: "General",
  },
  {
    flag: "MANAGE_WEBHOOKS",
    label: "Manage Webhooks",
    bit: 29,
    category: "General",
  },
  {
    flag: "MANAGE_GUILD_EXPRESSIONS",
    label: "Manage Expressions",
    bit: 30,
    category: "General",
  },
  {
    flag: "USE_APPLICATION_COMMANDS",
    label: "Use Application Commands",
    bit: 31,
    category: "Text",
  },
  {
    flag: "REQUEST_TO_SPEAK",
    label: "Request to Speak",
    bit: 32,
    category: "Voice",
  },
  {
    flag: "MANAGE_EVENTS",
    label: "Manage Events",
    bit: 33,
    category: "General",
  },
  {
    flag: "MANAGE_THREADS",
    label: "Manage Threads",
    bit: 34,
    category: "Text",
  },
  {
    flag: "CREATE_PUBLIC_THREADS",
    label: "Create Public Threads",
    bit: 35,
    category: "Text",
  },
  {
    flag: "CREATE_PRIVATE_THREADS",
    label: "Create Private Threads",
    bit: 36,
    category: "Text",
  },
  {
    flag: "USE_EXTERNAL_STICKERS",
    label: "Use External Stickers",
    bit: 37,
    category: "Text",
  },
  {
    flag: "SEND_MESSAGES_IN_THREADS",
    label: "Send Messages in Threads",
    bit: 38,
    category: "Text",
  },
  {
    flag: "USE_EMBEDDED_ACTIVITIES",
    label: "Use Activities",
    bit: 39,
    category: "Voice",
  },
  {
    flag: "MODERATE_MEMBERS",
    label: "Timeout Members",
    bit: 40,
    category: "Members",
  },
  {
    flag: "VIEW_CREATOR_MONETIZATION_ANALYTICS",
    label: "View Creator Monetization Analytics",
    bit: 41,
    category: "General",
  },
  {
    flag: "USE_SOUNDBOARD",
    label: "Use Soundboard",
    bit: 42,
    category: "Voice",
  },
  {
    flag: "CREATE_GUILD_EXPRESSIONS",
    label: "Create Expressions",
    bit: 43,
    category: "General",
  },
  {
    flag: "CREATE_EVENTS",
    label: "Create Events",
    bit: 44,
    category: "General",
  },
  {
    flag: "USE_EXTERNAL_SOUNDS",
    label: "Use External Sounds",
    bit: 45,
    category: "Voice",
  },
  {
    flag: "SEND_VOICE_MESSAGES",
    label: "Send Voice Messages",
    bit: 46,
    category: "Text",
  },
  {
    flag: "SET_VOICE_CHANNEL_STATUS",
    label: "Set Voice Channel Status",
    bit: 48,
    category: "Voice",
  },
  {
    flag: "SEND_POLLS",
    label: "Send Polls",
    bit: 49,
    category: "Text",
  },
  {
    flag: "USE_EXTERNAL_APPS",
    label: "Use External Apps",
    bit: 50,
    category: "Text",
  },
  {
    flag: "PIN_MESSAGES",
    label: "Pin Messages",
    bit: 51,
    category: "Text",
  },
  {
    flag: "BYPASS_SLOWMODE",
    label: "Bypass Slowmode",
    bit: 52,
    category: "Text",
  },
] as const;

export type Permission = (typeof permissionBits)[number];

export type PermissionFlag = Permission["flag"];

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
