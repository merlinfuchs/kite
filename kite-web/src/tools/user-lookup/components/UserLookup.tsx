import { apiRequest } from "@/lib/api/client";
import { LookupUserGetResponse } from "@/lib/types/wire.gen";
import LookupForm from "@/tools/common/components/LookupForm";
import {
  formatCreated,
  LookupHeader,
  LookupLink,
  LookupRow,
  LookupRows,
} from "@/tools/common/components/LookupResult";
import userAvatar, {
  avatarDecorationUrl,
  avatarUrl,
  bannerUrl,
  guildTagBadgeUrl,
} from "@/tools/common/utils/discordCdn";
import { flagNames, parseLookupId } from "@/tools/common/utils/lookup";
import { decodeSnowflake } from "@/tools/snowflake-decoder/snowflake";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

// The badges Discord shows on profiles, from the public flags of the user.
const userBadges: [number, string][] = [
  [1 << 0, "Discord Staff"],
  [1 << 1, "Partnered Server Owner"],
  [1 << 2, "HypeSquad Events"],
  [1 << 3, "Bug Hunter"],
  [1 << 14, "Golden Bug Hunter"],
  [1 << 6, "HypeSquad Bravery"],
  [1 << 7, "HypeSquad Brilliance"],
  [1 << 8, "HypeSquad Balance"],
  [1 << 9, "Early Supporter"],
  [1 << 16, "Verified bot"],
  [1 << 17, "Early Verified Bot Developer"],
  [1 << 18, "Moderator Programs Alumni"],
  [1 << 22, "Active Developer"],
];

function colorHex(color: number) {
  return "#" + color.toString(16).padStart(6, "0");
}

export default function UserLookup() {
  const [userId, setUserId] = useState<string | null>(null);
  const [inputError, setInputError] = useState<string>();

  const { data, isFetching } = useQuery({
    queryKey: ["lookup-user", userId],
    queryFn: () =>
      apiRequest<LookupUserGetResponse>(`/v1/lookup/users/${userId}`),
    enabled: !!userId,
    staleTime: 10 * 60 * 1000,
  });

  function lookup(value: string) {
    const id = parseLookupId(value);
    if (!id) {
      setInputError(
        "This isn't a Discord user ID. IDs are numbers with 17 to 20 digits."
      );
      return;
    }
    setInputError(undefined);
    setUserId(id);
  }

  const user = data?.success ? data.data : null;
  const error =
    inputError ?? (data && !data.success ? data.error.message : undefined);

  return (
    <div className="space-y-6">
      <LookupForm
        id="user-id"
        label="User ID"
        placeholder="1083038298513887292"
        loading={isFetching}
        error={error}
        onLookup={lookup}
      />
      {user && !inputError && <UserInfo user={user} />}
    </div>
  );
}

function UserInfo({ user }: { user: LookupUserGetResponse }) {
  const created = decodeSnowflake(user.id);
  const badges = flagNames(user.public_flags, userBadges);
  if (user.bot && !badges.includes("Verified bot")) badges.unshift("Bot");
  if (user.system) badges.unshift("System");

  const avatar = user.avatar ? avatarUrl(user.id, user.avatar, 1024) : null;
  const banner = user.banner ? bannerUrl(user.id, user.banner, 2048) : null;
  const tag = user.primary_guild;
  const legacyTag = user.discriminator !== "0";

  const rows: LookupRow[] = [
    { label: "User ID", value: user.id, copy: user.id },
    {
      label: "Username",
      value: legacyTag
        ? `${user.username}#${user.discriminator}`
        : user.username,
      copy: user.username,
    },
    { label: "Display name", value: user.global_name ?? "-" },
    {
      label: "Mention",
      value: <span className="font-mono">{`<@${user.id}>`}</span>,
      copy: `<@${user.id}>`,
    },
  ];
  if (created) {
    rows.push({
      label: "Account created",
      value: formatCreated(created.timestamp),
    });
  }
  if (tag) {
    rows.push({
      label: "Server tag",
      value: (
        <span className="inline-flex items-center gap-2">
          <img
            src={guildTagBadgeUrl(tag.guild_id, tag.badge)}
            alt=""
            className="w-4 h-4"
          />
          {tag.tag}
          <span className="text-muted-foreground">
            from server {tag.guild_id}
          </span>
        </span>
      ),
    });
  }
  if (user.accent_color !== null) {
    rows.push({
      label: "Banner color",
      value: (
        <span className="inline-flex items-center gap-2 font-mono">
          <span
            className="w-4 h-4 rounded border"
            style={{ backgroundColor: colorHex(user.accent_color) }}
          />
          {colorHex(user.accent_color)}
        </span>
      ),
      copy: colorHex(user.accent_color),
    });
  }
  rows.push({
    label: "Avatar",
    value: avatar ? (
      <LookupLink href={avatar}>Open image</LookupLink>
    ) : (
      "Default avatar"
    ),
  });
  if (banner) {
    rows.push({
      label: "Banner",
      value: <LookupLink href={banner}>Open image</LookupLink>,
    });
  }
  if (user.avatar_decoration) {
    rows.push({
      label: "Avatar decoration",
      value: (
        <LookupLink href={avatarDecorationUrl(user.avatar_decoration)}>
          Open image
        </LookupLink>
      ),
    });
  }

  return (
    <div className="space-y-6">
      <LookupHeader
        bannerUrl={banner}
        bannerColor={
          user.accent_color !== null ? colorHex(user.accent_color) : null
        }
        iconUrl={avatar ?? userAvatar(user)}
        iconDecorationUrl={
          user.avatar_decoration
            ? avatarDecorationUrl(user.avatar_decoration)
            : null
        }
        title={user.global_name ?? user.username}
        subtitle={
          legacyTag ? `${user.username}#${user.discriminator}` : user.username
        }
        badges={badges}
      />
      <LookupRows rows={rows} />
    </div>
  );
}
