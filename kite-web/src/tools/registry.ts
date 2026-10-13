import {
  ClockIcon,
  FileSearchIcon,
  FolderSearchIcon,
  ImageIcon,
  LucideIcon,
  MailPlusIcon,
  PaletteIcon,
  ShieldCheckIcon,
  SnowflakeIcon,
  UserSearchIcon,
  WebhookIcon,
} from "lucide-react";

export interface Tool {
  id: string;
  title: string;
  description: string;
  icon: LucideIcon;
  // Classes for the background and icon color of the icon tile.
  color: string;
  href: string;
  // Tools on other sites, which open in a new tab.
  external?: boolean;
}

// Tools that are part of kite.onl.
export const kiteTools: Tool[] = [
  {
    id: "permissions-calculator",
    title: "Permissions Calculator",
    description:
      "Get the permissions integer and an invite link for your Discord bot.",
    icon: ShieldCheckIcon,
    color: "bg-orange-500/15 text-orange-600 dark:text-orange-400",
    href: "/tools/permissions-calculator",
  },
  {
    id: "timestamp-generator",
    title: "Timestamp Generator",
    description:
      "Create timestamps that show everyone the time in their own time zone.",
    icon: ClockIcon,
    color: "bg-blue-500/15 text-blue-600 dark:text-blue-400",
    href: "/tools/timestamp-generator",
  },
  {
    id: "snowflake-decoder",
    title: "Snowflake Decoder",
    description: "Find out when a Discord user, server or message was created.",
    icon: SnowflakeIcon,
    color: "bg-cyan-500/15 text-cyan-600 dark:text-cyan-400",
    href: "/tools/snowflake-decoder",
  },
  {
    id: "webhook-info",
    title: "Webhook Info",
    description: "Get the name, avatar, channel and server of a webhook.",
    icon: WebhookIcon,
    color: "bg-violet-500/15 text-violet-600 dark:text-violet-400",
    href: "/tools/webhook-info",
  },
];

export const externalTools: Tool[] = [
  {
    id: "message-creator",
    title: "Message Creator",
    description:
      "Create Discord messages with embeds and send them through webhooks.",
    icon: MailPlusIcon,
    color: "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400",
    href: "https://message.style/app",
    external: true,
  },
  {
    id: "colored-text",
    title: "Colored Text",
    description: "Generate colored text for your Discord messages.",
    icon: PaletteIcon,
    color: "bg-pink-500/15 text-pink-600 dark:text-pink-400",
    href: "https://message.style/app/tools/colored-text",
    external: true,
  },
  {
    id: "embed-links",
    title: "Embed Links",
    description: "Create links with a custom title, description and image.",
    icon: ImageIcon,
    color: "bg-amber-500/15 text-amber-600 dark:text-amber-400",
    href: "https://message.style/app/tools/embed-links",
    external: true,
  },
  {
    id: "user-lookup",
    title: "User Lookup",
    description: "Get information about a Discord user from their ID.",
    icon: UserSearchIcon,
    color: "bg-slate-500/15 text-slate-600 dark:text-slate-400",
    href: "https://dis.wtf/lookup/user",
    external: true,
  },
  {
    id: "server-lookup",
    title: "Server Lookup",
    description: "Get information about a Discord server from its ID.",
    icon: FolderSearchIcon,
    color: "bg-slate-500/15 text-slate-600 dark:text-slate-400",
    href: "https://dis.wtf/lookup/guild",
    external: true,
  },
  {
    id: "invite-resolver",
    title: "Invite Resolver",
    description: "Get information about a Discord server from an invite code.",
    icon: FolderSearchIcon,
    color: "bg-slate-500/15 text-slate-600 dark:text-slate-400",
    href: "https://dis.wtf/lookup/invite",
    external: true,
  },
  {
    id: "application-lookup",
    title: "Application Lookup",
    description: "Get information about a Discord application from its ID.",
    icon: FileSearchIcon,
    color: "bg-slate-500/15 text-slate-600 dark:text-slate-400",
    href: "https://dis.wtf/lookup/app",
    external: true,
  },
];
