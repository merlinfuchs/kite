import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  Table as TanstackTable,
  useReactTable,
} from "@tanstack/react-table";
import {
  CheckIcon,
  CopyIcon,
  EyeOffIcon,
  FolderIcon,
  HashIcon,
  ImageIcon,
  LucideIcon,
  MegaphoneIcon,
  MessageSquareTextIcon,
  MessagesSquareIcon,
  RadioIcon,
  Volume2Icon,
  XIcon,
} from "lucide-react";
import { useRouter } from "next/router";
import { ReactNode, useCallback, useMemo, useState } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  useAppStateGuildChannelsQuery,
  useAppStateGuildMembersQuery,
  useAppStateGuildRolesQuery,
  useAppStateGuildsQuery,
} from "@/lib/api/queries";
import { useApp, useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import {
  Channel,
  Guild,
  MaxStateGuildMembers,
  Member,
  Role,
} from "@/lib/types/wire.gen";

export type GuildExplorerView = "channels" | "members" | "roles";

const channelTypes: Record<number, { label: string; icon: LucideIcon }> = {
  0: { label: "Text", icon: HashIcon },
  2: { label: "Voice", icon: Volume2Icon },
  4: { label: "Category", icon: FolderIcon },
  5: { label: "Announcement", icon: MegaphoneIcon },
  10: { label: "Announcement Thread", icon: MessageSquareTextIcon },
  11: { label: "Thread", icon: MessageSquareTextIcon },
  12: { label: "Private Thread", icon: MessageSquareTextIcon },
  13: { label: "Stage", icon: RadioIcon },
  15: { label: "Forum", icon: MessagesSquareIcon },
  16: { label: "Media", icon: ImageIcon },
};

function channelType(channel: Channel) {
  return channelTypes[channel.type] ?? { label: "Channel", icon: HashIcon };
}

function isCategory(channel: Channel) {
  return channel.type === 4;
}

function isThread(channel: Channel) {
  return channel.type === 10 || channel.type === 11 || channel.type === 12;
}

// Voice and stage channels come after the other channels of their category,
// like in Discord.
function isVoiceChannel(channel: Channel) {
  return channel.type === 2 || channel.type === 13;
}

function compareChannels(a: Channel, b: Channel) {
  const voice = Number(isVoiceChannel(a)) - Number(isVoiceChannel(b));
  if (voice !== 0) return voice;
  return (a.position ?? 0) - (b.position ?? 0);
}

// Discord permission bits, as BigInt because some don't fit in a number.
const permissionBits = {
  addReactions: 6,
  viewChannel: 10,
  sendMessages: 11,
  embedLinks: 14,
  attachFiles: 15,
  readMessageHistory: 16,
  connect: 20,
  speak: 21,
  sendMessagesInThreads: 38,
} as const;

type PermissionName = keyof typeof permissionBits;

const permissionLabels: Record<PermissionName, string> = {
  viewChannel: "View Channel",
  sendMessages: "Send Messages",
  sendMessagesInThreads: "Send Messages in Threads",
  readMessageHistory: "Read Message History",
  embedLinks: "Embed Links",
  attachFiles: "Attach Files",
  addReactions: "Add Reactions",
  connect: "Connect",
  speak: "Speak",
};

// The permissions that matter for each kind of channel, in display order.
function relevantPermissions(channel: Channel): PermissionName[] {
  if (isVoiceChannel(channel)) {
    return ["viewChannel", "connect", "speak", "sendMessages"];
  }
  if (isCategory(channel)) {
    return ["viewChannel"];
  }
  return [
    "viewChannel",
    isThread(channel) ? "sendMessagesInThreads" : "sendMessages",
    "readMessageHistory",
    "embedLinks",
    "attachFiles",
    "addReactions",
  ];
}

// Returns null when the service couldn't work out the bot's permissions.
function botPermissions(
  channel: Channel
): Record<PermissionName, boolean> | null {
  if (!channel.bot_permissions) return null;

  let bits: bigint;
  try {
    bits = BigInt(channel.bot_permissions);
  } catch {
    return null;
  }

  const res = {} as Record<PermissionName, boolean>;
  for (const [name, bit] of Object.entries(permissionBits)) {
    res[name as PermissionName] =
      (bits & (BigInt(1) << BigInt(bit))) !== BigInt(0);
  }
  return res;
}

// Whether the bot can see the channel. Channels with unknown permissions
// count as visible, so nothing is hidden by mistake.
function canView(channel: Channel) {
  return botPermissions(channel)?.viewChannel ?? true;
}

// The most important thing the bot can't do in a channel it can see.
function accessIssue(channel: Channel): string | null {
  const perms = botPermissions(channel);
  if (!perms || !perms.viewChannel) return null;

  if (isVoiceChannel(channel)) {
    if (!perms.connect) return "Can't connect";
    if (!perms.speak) return "Can't speak";
    return null;
  }
  if (isThread(channel)) {
    return perms.sendMessagesInThreads ? null : "Can't send";
  }
  if (!isCategory(channel) && !perms.sendMessages) return "Can't send";
  return null;
}

function permissionList(channel: Channel) {
  const perms = botPermissions(channel);
  if (!perms) return null;

  return relevantPermissions(channel).map((name) => ({
    label: permissionLabels[name],
    allowed: perms[name],
  }));
}

function copyId(id: string) {
  navigator.clipboard.writeText(id);
  toast.success("ID copied to clipboard");
}

function CopyIdButton({ id }: { id: string }) {
  return (
    <Button
      variant="ghost"
      className="h-8 w-8 p-0"
      onClick={() => copyId(id)}
      title="Copy ID"
    >
      <span className="sr-only">Copy ID</span>
      <CopyIcon className="h-4 w-4" />
    </Button>
  );
}

interface Props {
  view: GuildExplorerView;
}

// Lists the channels or members of one server the app is in. The server is
// picked next to the filter and kept in the URL, so the server list can link
// to it.
export default function AppStateGuildExplorer({ view }: Props) {
  const router = useRouter();
  const appId = useAppId();
  const guildsQuery = useAppStateGuildsQuery(appId);
  const guilds = useResponseData(guildsQuery);

  const guildList = useMemo(
    () =>
      (guilds ?? [])
        .filter((g): g is Guild => !!g)
        .sort((a, b) => a.name.localeCompare(b.name)),
    [guilds]
  );

  const queryGuildId =
    typeof router.query.guild === "string" ? router.query.guild : null;
  const guild =
    guildList.find((g) => g.id === queryGuildId) ?? guildList[0] ?? null;

  const setGuild = useCallback(
    (guildId: string) => {
      router.replace(
        {
          pathname: router.pathname,
          query: { ...router.query, guild: guildId },
        },
        undefined,
        { shallow: true }
      );
    },
    [router]
  );

  const guildSelect = (
    <Select value={guild?.id ?? ""} onValueChange={setGuild} disabled={!guild}>
      <SelectTrigger className="w-64">
        <SelectValue placeholder="No servers" />
      </SelectTrigger>
      <SelectContent>
        {guildList.map((g) => (
          <SelectItem key={g.id} value={g.id}>
            <div className="flex items-center gap-2">
              {g.icon_url ? (
                <img
                  src={g.icon_url}
                  alt=""
                  className="w-5 h-5 rounded-full flex-none"
                />
              ) : (
                <div className="w-5 h-5 rounded-full bg-muted flex-none" />
              )}
              <span className="truncate">{g.name}</span>
            </div>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );

  const key = guild?.id;
  const tableProps = {
    guildId: guild?.id ?? null,
    guildsLoading: guildsQuery.isLoading,
    guildSelect,
  };

  if (view === "channels") {
    return <GuildChannelTable key={key} {...tableProps} />;
  }
  if (view === "members") {
    return <GuildMemberTable key={key} {...tableProps} />;
  }
  return <GuildRoleTable key={key} {...tableProps} />;
}

function ExplorerTable<T>({
  table,
  columnCount,
  empty,
  toolbar,
  footer,
}: {
  table: TanstackTable<T>;
  columnCount: number;
  empty: ReactNode;
  toolbar: ReactNode;
  footer: ReactNode;
}) {
  return (
    <div className="w-full">
      <div className="flex flex-wrap items-center py-4 gap-3">{toolbar}</div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columnCount} className="h-24 text-center">
                  {empty}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm text-muted-foreground">{footer}</div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}

interface ChannelRow {
  channel: Channel;
  parent: Channel | null;
  category: Channel | null;
}

const channelColumns: ColumnDef<ChannelRow>[] = [
  {
    id: "icon",
    cell: ({ row }) => {
      const { channel } = row.original;
      const Icon = channelType(channel).icon;
      return (
        <div
          className={`flex items-center ${isThread(channel) ? "pl-4" : ""} ${
            canView(channel) ? "" : "opacity-50"
          }`}
        >
          <Icon className="h-4 w-4 text-muted-foreground" />
        </div>
      );
    },
  },
  {
    id: "name",
    header: "Name",
    cell: ({ row }) => {
      const { channel, parent } = row.original;
      return (
        <div
          className={`min-w-0 ${canView(channel) ? "" : "opacity-50"} ${
            isCategory(channel) ? "text-muted-foreground font-medium" : ""
          }`}
        >
          <div className="truncate max-w-72">{channel.name}</div>
          {isThread(channel) ? (
            <div className="text-xs text-muted-foreground truncate max-w-72">
              in #{parent?.name ?? "unknown channel"}
            </div>
          ) : (
            channel.topic && (
              <div className="text-xs text-muted-foreground truncate max-w-72">
                {channel.topic}
              </div>
            )
          )}
        </div>
      );
    },
  },
  {
    id: "type",
    header: "Type",
    cell: ({ row }) => <div>{channelType(row.original.channel).label}</div>,
  },
  {
    id: "category",
    header: "Category",
    cell: ({ row }) => (
      <div className="text-muted-foreground">
        {row.original.category?.name ?? "—"}
      </div>
    ),
  },
  {
    id: "access",
    header: "Bot Access",
    cell: ({ row }) => {
      const { channel } = row.original;
      const permissions = permissionList(channel);
      if (!permissions) {
        return <div className="text-muted-foreground">—</div>;
      }

      const issue = accessIssue(channel);
      const badge = !canView(channel) ? (
        <Badge variant="outline" className="font-normal gap-1">
          <EyeOffIcon className="h-3 w-3" />
          No access
        </Badge>
      ) : issue ? (
        <Badge
          variant="outline"
          className="font-normal border-amber-500/50 text-amber-500"
        >
          {issue}
        </Badge>
      ) : (
        <Badge
          variant="outline"
          className="font-normal border-green-500/50 text-green-500"
        >
          OK
        </Badge>
      );

      return (
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="w-fit">{badge}</div>
          </TooltipTrigger>
          <TooltipContent side="right">
            <div className="space-y-1">
              {permissions.map(({ label, allowed }) => (
                <div key={label} className="flex items-center gap-2">
                  {allowed ? (
                    <CheckIcon className="h-3.5 w-3.5 text-green-500" />
                  ) : (
                    <XIcon className="h-3.5 w-3.5 text-red-500" />
                  )}
                  <span>{label}</span>
                </div>
              ))}
            </div>
          </TooltipContent>
        </Tooltip>
      );
    },
  },
  {
    id: "id",
    header: "ID",
    cell: ({ row }) => <div>{row.original.channel.id}</div>,
  },
  {
    id: "actions",
    cell: ({ row }) => <CopyIdButton id={row.original.channel.id} />,
  },
];

// Puts channels in Discord's order: channels without a category, then each
// category followed by its channels. Threads come right after their parent.
function orderChannels(all: Channel[]): ChannelRow[] {
  const byId = new Map(all.map((c) => [c.id, c]));
  const categories = all
    .filter(isCategory)
    .sort((a, b) => (a.position ?? 0) - (b.position ?? 0));
  const channels = all.filter((c) => !isCategory(c) && !isThread(c));
  const threads = all
    .filter(isThread)
    .sort((a, b) => a.name.localeCompare(b.name));

  const res: ChannelRow[] = [];
  const placedThreads = new Set<string>();

  const pushChannel = (channel: Channel, category: Channel | null) => {
    res.push({ channel, parent: null, category });
    for (const thread of threads) {
      if (thread.parent_id === channel.id) {
        res.push({ channel: thread, parent: channel, category });
        placedThreads.add(thread.id);
      }
    }
  };

  const categoryIds = new Set(categories.map((c) => c.id));
  channels
    .filter((c) => !c.parent_id || !categoryIds.has(c.parent_id))
    .sort(compareChannels)
    .forEach((c) => pushChannel(c, null));

  for (const category of categories) {
    res.push({ channel: category, parent: null, category: null });
    channels
      .filter((c) => c.parent_id === category.id)
      .sort(compareChannels)
      .forEach((c) => pushChannel(c, category));
  }

  // Threads whose parent isn't listed.
  for (const thread of threads) {
    if (!placedThreads.has(thread.id)) {
      const parent = thread.parent_id ? byId.get(thread.parent_id) : undefined;
      res.push({ channel: thread, parent: parent ?? null, category: null });
    }
  }

  return res;
}

function GuildChannelTable({
  guildId,
  guildsLoading,
  guildSelect,
}: {
  guildId: string | null;
  guildsLoading: boolean;
  guildSelect: ReactNode;
}) {
  const query = useAppStateGuildChannelsQuery(useAppId(), guildId);
  const channels = useResponseData(query);
  const [search, setSearch] = useState("");
  const [showHidden, setShowHidden] = useState(false);

  const all = useMemo(
    () => (channels ?? []).filter((c): c is Channel => !!c),
    [channels]
  );
  const ordered = useMemo(() => orderChannels(all), [all]);

  const tableData = useMemo(() => {
    const term = search.trim().toLowerCase();
    return ordered.filter(
      ({ channel }) =>
        (showHidden || canView(channel)) &&
        (!term ||
          channel.name.toLowerCase().includes(term) ||
          channel.id.includes(term))
    );
  }, [ordered, search, showHidden]);

  const table = useReactTable({
    data: tableData,
    columns: channelColumns,
    getRowId: (row) => row.channel.id,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    autoResetPageIndex: true,
  });

  // Categories are rows in the table too, so they count like any channel.
  const total = all.length;
  const hiddenCount = all.filter((c) => !canView(c)).length;

  let empty: ReactNode = "No results.";
  if (guildsLoading || (guildId && query.isLoading)) {
    empty = "Loading channels...";
  } else if (!guildId) {
    empty = "Your app isn't in any servers. Start it to see its channels.";
  } else if (query.data && !query.data.success) {
    empty = `Failed to load channels: ${query.data.error.message}`;
  } else if (!search && hiddenCount > 0 && hiddenCount === total) {
    empty = "The bot can't see any channels in this server.";
  }

  return (
    <ExplorerTable
      table={table}
      columnCount={channelColumns.length}
      empty={empty}
      toolbar={
        <>
          <Input
            placeholder="Filter channels..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-sm"
          />
          {guildSelect}
          {hiddenCount > 0 && (
            <label className="flex items-center gap-2 text-sm text-muted-foreground cursor-pointer ml-auto">
              Show {hiddenCount} channel{hiddenCount === 1 ? "" : "s"} the bot
              can&apos;t see
              <Switch checked={showHidden} onCheckedChange={setShowHidden} />
            </label>
          )}
        </>
      }
      footer={
        <>
          Showing {table.getRowModel().rows.length} of {tableData.length} total
          channels.
          {total > 0 &&
            all.some((c) => !!c.bot_permissions) &&
            ` The bot can see ${total - hiddenCount} of ${total}.`}
        </>
      }
    />
  );
}

interface MemberRow {
  member: Member;
  roles: string[];
}

const memberColumns: ColumnDef<MemberRow>[] = [
  {
    id: "avatar",
    cell: ({ row }) => {
      const url = row.original.member.avatar_url;
      if (!url) {
        return <div className="w-8 h-8 rounded-full bg-muted"></div>;
      }
      return <img src={url} alt="" className="w-8 h-8 rounded-full" />;
    },
  },
  {
    id: "name",
    header: "Name",
    cell: ({ row }) => {
      const { member } = row.original;
      return (
        <div className="flex items-center gap-2">
          <span className="truncate max-w-56">{member.display_name}</span>
          {member.bot && (
            <Badge
              variant="secondary"
              className="px-1.5 py-0 text-[11px] leading-4 font-normal"
            >
              Bot
            </Badge>
          )}
        </div>
      );
    },
  },
  {
    id: "username",
    header: "Username",
    cell: ({ row }) => (
      <div className="text-muted-foreground">
        @{row.original.member.username}
      </div>
    ),
  },
  {
    id: "roles",
    header: "Roles",
    cell: ({ row }) => {
      const { roles } = row.original;
      if (!roles.length) {
        return <div className="text-muted-foreground">—</div>;
      }

      const shown = roles.slice(0, 2);
      // A member can have a lot of roles, so the tooltip doesn't list all.
      const listed = roles.slice(0, 20);
      return (
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="flex items-center gap-1 w-fit">
              {shown.map((name, i) => (
                <Badge
                  key={i}
                  variant="outline"
                  className="font-normal max-w-40"
                >
                  <span className="truncate">{name}</span>
                </Badge>
              ))}
              {roles.length > shown.length && (
                <span className="text-xs text-muted-foreground">
                  +{roles.length - shown.length}
                </span>
              )}
            </div>
          </TooltipTrigger>
          <TooltipContent side="right">
            <div className="space-y-1">
              {listed.map((name, i) => (
                <div key={i}>{name}</div>
              ))}
              {roles.length > listed.length && (
                <div className="text-muted-foreground">
                  and {roles.length - listed.length} more
                </div>
              )}
            </div>
          </TooltipContent>
        </Tooltip>
      );
    },
  },
  {
    id: "joined_at",
    header: "Joined At",
    cell: ({ row }) => {
      const joinedAt = row.original.member.joined_at;
      return <div>{joinedAt ? new Date(joinedAt).toLocaleString() : "—"}</div>;
    },
  },
  {
    id: "id",
    header: "ID",
    cell: ({ row }) => <div>{row.original.member.id}</div>,
  },
  {
    id: "actions",
    cell: ({ row }) => <CopyIdButton id={row.original.member.id} />,
  },
];

function GuildMemberTable({
  guildId,
  guildsLoading,
  guildSelect,
}: {
  guildId: string | null;
  guildsLoading: boolean;
  guildSelect: ReactNode;
}) {
  const appId = useAppId();
  const app = useApp();
  const query = useAppStateGuildMembersQuery(appId, guildId);
  const members = useResponseData(query);
  const roles = useResponseData(useAppStateGuildRolesQuery(appId, guildId));
  const [search, setSearch] = useState("");

  const all = useMemo(() => {
    const roleById = new Map(
      (roles ?? []).filter((r) => !!r).map((r) => [r!.id, r!])
    );

    return (members ?? [])
      .filter((m): m is Member => !!m)
      .sort((a, b) => a.display_name.localeCompare(b.display_name))
      .map((member) => ({
        member,
        // Highest role first, like in Discord.
        roles: member.role_ids
          .map((id) => roleById.get(id))
          .filter((r) => !!r)
          .sort((a, b) => b!.position - a!.position)
          .map((r) => r!.name),
      }));
  }, [members, roles]);

  const tableData = useMemo(() => {
    const term = search.trim().toLowerCase();
    if (!term) return all;
    return all.filter(
      ({ member }) =>
        member.display_name.toLowerCase().includes(term) ||
        member.username.toLowerCase().includes(term) ||
        member.id.includes(term)
    );
  }, [all, search]);

  const table = useReactTable({
    data: tableData,
    columns: memberColumns,
    getRowId: (row) => row.member.id,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    autoResetPageIndex: true,
  });

  const error = query.data && !query.data.success ? query.data.error : null;

  let empty: ReactNode = "No results.";
  if (guildsLoading || (guildId && query.isLoading)) {
    empty = "Loading members...";
  } else if (!guildId) {
    empty = "Your app isn't in any servers. Start it to see its members.";
  } else if (error) {
    empty = (
      <div className="flex flex-col items-center gap-3">
        <div>{error.message}</div>
        {error.code === "missing_members_intent" && app && (
          <Button variant="outline" size="sm" asChild>
            <a
              href={`https://discord.com/developers/applications/${app.discord_id}/bot`}
              target="_blank"
              rel="noopener noreferrer"
            >
              Open Developer Portal
            </a>
          </Button>
        )}
      </div>
    );
  }

  return (
    <ExplorerTable
      table={table}
      columnCount={memberColumns.length}
      empty={empty}
      toolbar={
        <>
          <Input
            placeholder="Filter members..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-sm"
          />
          {guildSelect}
        </>
      }
      footer={
        <>
          Showing {table.getRowModel().rows.length} of {tableData.length} total
          members.
          {all.length >= MaxStateGuildMembers &&
            ` Only the first ${MaxStateGuildMembers} members are listed.`}
        </>
      }
    />
  );
}

// Every Discord permission by bit, named like in Discord's role settings.
const allPermissions: [number, string][] = [
  [3, "Administrator"],
  [10, "View Channels"],
  [4, "Manage Channels"],
  [28, "Manage Roles"],
  [43, "Create Expressions"],
  [30, "Manage Expressions"],
  [7, "View Audit Log"],
  [19, "View Server Insights"],
  [29, "Manage Webhooks"],
  [5, "Manage Server"],
  [0, "Create Invite"],
  [26, "Change Nickname"],
  [27, "Manage Nicknames"],
  [1, "Kick Members"],
  [2, "Ban Members"],
  [40, "Timeout Members"],
  [11, "Send Messages"],
  [38, "Send Messages in Threads"],
  [35, "Create Public Threads"],
  [36, "Create Private Threads"],
  [14, "Embed Links"],
  [15, "Attach Files"],
  [6, "Add Reactions"],
  [18, "Use External Emojis"],
  [37, "Use External Stickers"],
  [17, "Mention @everyone, @here and All Roles"],
  [13, "Manage Messages"],
  [51, "Pin Messages"],
  [52, "Bypass Slowmode"],
  [34, "Manage Threads"],
  [16, "Read Message History"],
  [12, "Send Text-to-Speech Messages"],
  [46, "Send Voice Messages"],
  [49, "Create Polls"],
  [20, "Connect"],
  [21, "Speak"],
  [9, "Video"],
  [42, "Use Soundboard"],
  [45, "Use External Sounds"],
  [25, "Use Voice Activity"],
  [8, "Priority Speaker"],
  [22, "Mute Members"],
  [23, "Deafen Members"],
  [24, "Move Members"],
  [32, "Request to Speak"],
  [31, "Use Application Commands"],
  [39, "Use Activities"],
  [50, "Use External Apps"],
  [44, "Create Events"],
  [33, "Manage Events"],
  [41, "View Creator Monetization Analytics"],
];

function rolePermissionNames(role: Role): string[] {
  let bits: bigint;
  try {
    bits = BigInt(role.permissions);
  } catch {
    return [];
  }

  return allPermissions
    .filter(([bit]) => (bits & (BigInt(1) << BigInt(bit))) !== BigInt(0))
    .map(([, name]) => name);
}

function roleColorHex(color: number) {
  return `#${color.toString(16).padStart(6, "0").toUpperCase()}`;
}

// @everyone has the same ID as the server.
function isEveryoneRole(role: Role, guildId: string | null) {
  return role.id === guildId;
}

interface RoleRow {
  role: Role;
  permissions: string[];
  everyone: boolean;
}

const roleColumns: ColumnDef<RoleRow>[] = [
  {
    id: "icon",
    cell: ({ row }) => {
      const { role } = row.original;
      if (role.icon_url) {
        return <img src={role.icon_url} alt="" className="w-6 h-6" />;
      }
      if (role.unicode_emoji) {
        return (
          <div className="w-6 text-lg leading-6">{role.unicode_emoji}</div>
        );
      }
      // Without an icon, show the role's color like Discord's role list.
      return (
        <div
          className="w-3 h-3 rounded-full ml-1.5 bg-muted-foreground"
          style={
            role.color ? { backgroundColor: roleColorHex(role.color) } : {}
          }
        />
      );
    },
  },
  {
    id: "name",
    header: "Name",
    cell: ({ row }) => {
      const { role, everyone } = row.original;
      return (
        <div className="flex items-center gap-2">
          <span
            className="truncate max-w-56 font-medium"
            style={role.color ? { color: roleColorHex(role.color) } : {}}
          >
            {everyone ? "@everyone" : role.name}
          </span>
          {role.managed && (
            <Badge
              variant="secondary"
              className="px-1.5 py-0 text-[11px] leading-4 font-normal"
            >
              Managed
            </Badge>
          )}
        </div>
      );
    },
  },
  {
    id: "color",
    header: "Color",
    cell: ({ row }) => {
      const { color } = row.original.role;
      if (!color) {
        return <div className="text-muted-foreground">Default</div>;
      }
      return (
        <div className="flex items-center gap-2">
          <div
            className="w-4 h-4 rounded border"
            style={{ backgroundColor: roleColorHex(color) }}
          />
          <span className="font-mono text-xs">{roleColorHex(color)}</span>
        </div>
      );
    },
  },
  {
    id: "permissions",
    header: "Permissions",
    cell: ({ row }) => {
      const { permissions } = row.original;
      if (!permissions.length) {
        return <div className="text-muted-foreground">None</div>;
      }

      const admin = permissions.includes("Administrator");
      const listed = permissions.slice(0, 24);

      return (
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="w-fit">
              {admin ? (
                <Badge
                  variant="outline"
                  className="font-normal border-red-500/50 text-red-500"
                >
                  Administrator
                </Badge>
              ) : (
                <Badge variant="outline" className="font-normal">
                  {permissions.length} permission
                  {permissions.length === 1 ? "" : "s"}
                </Badge>
              )}
            </div>
          </TooltipTrigger>
          <TooltipContent side="right" className="max-w-md">
            {admin && (
              <div className="mb-2 text-muted-foreground">
                Administrator grants every permission.
              </div>
            )}
            <div className="grid grid-cols-2 gap-x-4 gap-y-1">
              {listed.map((name) => (
                <div key={name} className="flex items-center gap-2">
                  <CheckIcon className="h-3.5 w-3.5 flex-none text-green-500" />
                  <span>{name}</span>
                </div>
              ))}
            </div>
            {permissions.length > listed.length && (
              <div className="mt-1 text-muted-foreground">
                and {permissions.length - listed.length} more
              </div>
            )}
          </TooltipContent>
        </Tooltip>
      );
    },
  },
  {
    id: "created_at",
    header: "Created At",
    cell: ({ row }) => (
      <div>{new Date(row.original.role.created_at).toLocaleString()}</div>
    ),
  },
  {
    id: "id",
    header: "ID",
    cell: ({ row }) => <div>{row.original.role.id}</div>,
  },
  {
    id: "actions",
    cell: ({ row }) => <CopyIdButton id={row.original.role.id} />,
  },
];

function GuildRoleTable({
  guildId,
  guildsLoading,
  guildSelect,
}: {
  guildId: string | null;
  guildsLoading: boolean;
  guildSelect: ReactNode;
}) {
  const query = useAppStateGuildRolesQuery(useAppId(), guildId);
  const roles = useResponseData(query);
  const [search, setSearch] = useState("");

  // Highest role first, like in Discord.
  const all = useMemo(
    () =>
      (roles ?? [])
        .filter((r): r is Role => !!r)
        .sort((a, b) => b.position - a.position)
        .map((role) => ({
          role,
          permissions: rolePermissionNames(role),
          everyone: isEveryoneRole(role, guildId),
        })),
    [roles, guildId]
  );

  const tableData = useMemo(() => {
    const term = search.trim().toLowerCase();
    if (!term) return all;
    return all.filter(
      ({ role, everyone }) =>
        (everyone ? "@everyone" : role.name).toLowerCase().includes(term) ||
        role.id.includes(term)
    );
  }, [all, search]);

  const table = useReactTable({
    data: tableData,
    columns: roleColumns,
    getRowId: (row) => row.role.id,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    autoResetPageIndex: true,
  });

  let empty: ReactNode = "No results.";
  if (guildsLoading || (guildId && query.isLoading)) {
    empty = "Loading roles...";
  } else if (!guildId) {
    empty = "Your app isn't in any servers. Start it to see its roles.";
  } else if (query.data && !query.data.success) {
    empty = `Failed to load roles: ${query.data.error.message}`;
  }

  return (
    <ExplorerTable
      table={table}
      columnCount={roleColumns.length}
      empty={empty}
      toolbar={
        <>
          <Input
            placeholder="Filter roles..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-sm"
          />
          {guildSelect}
        </>
      }
      footer={
        <>
          Showing {table.getRowModel().rows.length} of {tableData.length} total
          roles.
        </>
      }
    />
  );
}
