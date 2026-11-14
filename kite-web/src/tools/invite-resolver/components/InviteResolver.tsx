import LookupForm from "@/tools/common/components/LookupForm";
import {
  formatCreated,
  LookupHeader,
  LookupLink,
  LookupRow,
  LookupRows,
} from "@/tools/common/components/LookupResult";
import userAvatar, {
  bannerUrl,
  guildIconUrl,
  guildSplashUrl,
  guildTagBadgeUrl,
} from "@/tools/common/utils/discordCdn";
import { decodeSnowflake } from "@/tools/snowflake-decoder/snowflake";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

interface InviteData {
  type: number;
  code: string;
  expires_at: string | null;
  guild?: {
    id: string;
    name: string;
    icon: string | null;
    banner: string | null;
    splash: string | null;
    description: string | null;
    features: string[];
    verification_level: number;
    vanity_url_code: string | null;
    nsfw_level: number;
    premium_subscription_count?: number;
    premium_tier?: number;
  };
  channel?: { id: string; type: number; name: string | null } | null;
  inviter?: {
    id: string;
    username: string;
    global_name: string | null;
    discriminator: string;
    avatar: string | null;
  };
  profile?: {
    tag: string | null;
    badge_hash: string | null;
  };
  approximate_member_count?: number;
  approximate_presence_count?: number;
}

type InviteResponse =
  | { success: true; invite: InviteData }
  | { success: false; error: string };

const inviteRegex =
  /^(?:https?:\/\/)?(?:www\.)?(?:(?:canary\.|ptb\.)?discord(?:app)?\.com\/invite|discord\.gg)\/([a-zA-Z0-9-]+)\/?$/;

// Finds the code in an invite link, or takes the code itself.
function parseInviteCode(value: string) {
  const match = value.match(inviteRegex);
  if (match) return match[1];
  if (/^[a-zA-Z0-9-]{2,}$/.test(value)) return value;
  return null;
}

const verificationLevels = ["None", "Low", "Medium", "High", "Highest"];

const featureBadges: [string, string][] = [
  ["VERIFIED", "Verified"],
  ["PARTNERED", "Partnered"],
  ["COMMUNITY", "Community"],
  ["DISCOVERABLE", "Discoverable"],
];

async function fetchInvite(code: string): Promise<InviteResponse> {
  const res = await fetch(
    `https://discord.com/api/v10/invites/${encodeURIComponent(
      code
    )}?with_counts=true&with_expiration=true`
  );
  if (res.status === 404) {
    return {
      success: false,
      error: "This invite doesn't exist. It may have expired or been deleted.",
    };
  }
  if (res.status === 429) {
    return {
      success: false,
      error: "Discord is limiting lookups right now. Try again in a minute.",
    };
  }
  if (!res.ok) {
    return {
      success: false,
      error: "Discord didn't answer. Try again in a moment.",
    };
  }
  return { success: true, invite: await res.json() };
}

export default function InviteResolver() {
  const [code, setCode] = useState<string | null>(null);
  const [inputError, setInputError] = useState<string>();

  const { data, isFetching } = useQuery({
    queryKey: ["lookup-invite", code],
    queryFn: () => fetchInvite(code!),
    enabled: !!code,
    staleTime: 5 * 60 * 1000,
  });

  function lookup(value: string) {
    const parsed = parseInviteCode(value);
    if (!parsed) {
      setInputError(
        "This isn't an invite. Paste a link like discord.gg/abc123 or just its code."
      );
      return;
    }
    setInputError(undefined);
    setCode(parsed);
  }

  const error = inputError ?? (data && !data.success ? data.error : undefined);

  return (
    <div className="space-y-6">
      <LookupForm
        id="invite"
        label="Invite link or code"
        placeholder="discord.gg/discord-developers"
        loading={isFetching}
        error={error}
        onLookup={lookup}
      />
      {data?.success && !inputError && <InviteInfo invite={data.invite} />}
    </div>
  );
}

function InviteInfo({ invite }: { invite: InviteData }) {
  const guild = invite.guild;
  const inviter = invite.inviter;
  const rows: LookupRow[] = [];

  if (guild) {
    const created = decodeSnowflake(guild.id);
    rows.push({ label: "Server ID", value: guild.id, copy: guild.id });
    if (created) {
      rows.push({
        label: "Server created",
        value: formatCreated(created.timestamp),
      });
    }
    if (invite.approximate_member_count !== undefined) {
      rows.push({
        label: "Members",
        value: invite.approximate_member_count.toLocaleString(),
      });
    }
    if (invite.approximate_presence_count !== undefined) {
      rows.push({
        label: "Online",
        value: invite.approximate_presence_count.toLocaleString(),
      });
    }
    if (guild.premium_subscription_count !== undefined) {
      rows.push({
        label: "Boosts",
        value: guild.premium_tier
          ? `${guild.premium_subscription_count} (level ${guild.premium_tier})`
          : String(guild.premium_subscription_count),
      });
    }
    rows.push({
      label: "Verification level",
      value: verificationLevels[guild.verification_level] ?? "Unknown",
    });
    if (guild.vanity_url_code) {
      rows.push({
        label: "Vanity link",
        value: `discord.gg/${guild.vanity_url_code}`,
        copy: `https://discord.gg/${guild.vanity_url_code}`,
      });
    }
    if (invite.profile?.tag) {
      rows.push({
        label: "Server tag",
        value: (
          <span className="inline-flex items-center gap-2">
            {invite.profile.badge_hash && (
              <img
                src={guildTagBadgeUrl(guild.id, invite.profile.badge_hash)}
                alt=""
                className="w-4 h-4"
              />
            )}
            {invite.profile.tag}
          </span>
        ),
      });
    }
  }

  if (invite.channel) {
    rows.push({
      label: "Channel",
      value: invite.channel.name
        ? `${guild ? "#" : ""}${invite.channel.name} (${invite.channel.id})`
        : invite.channel.id,
      copy: invite.channel.id,
    });
  }

  if (inviter) {
    rows.push({
      label: "Created by",
      value: (
        <span className="inline-flex items-center gap-2">
          <img
            src={userAvatar(inviter)}
            alt=""
            className="w-5 h-5 rounded-full"
          />
          {inviter.global_name ?? inviter.username}
          <span className="text-muted-foreground">({inviter.id})</span>
        </span>
      ),
      copy: inviter.id,
    });
  }

  rows.push({
    label: "Expires",
    value: invite.expires_at
      ? formatCreated(new Date(invite.expires_at).getTime())
      : "Never",
  });

  if (guild?.icon) {
    rows.push({
      label: "Icon",
      value: (
        <LookupLink href={guildIconUrl(guild.id, guild.icon, 1024)}>
          Open image
        </LookupLink>
      ),
    });
  }
  if (guild?.banner) {
    rows.push({
      label: "Banner",
      value: (
        <LookupLink href={bannerUrl(guild.id, guild.banner, 2048)}>
          Open image
        </LookupLink>
      ),
    });
  }
  if (guild?.splash) {
    rows.push({
      label: "Invite background",
      value: (
        <LookupLink href={guildSplashUrl(guild.id, guild.splash, 2048)}>
          Open image
        </LookupLink>
      ),
    });
  }

  const header = guild ? (
    <LookupHeader
      bannerUrl={
        guild.banner
          ? bannerUrl(guild.id, guild.banner)
          : guild.splash
          ? guildSplashUrl(guild.id, guild.splash)
          : null
      }
      iconUrl={guild.icon ? guildIconUrl(guild.id, guild.icon) : null}
      iconShape="rounded"
      title={guild.name}
      subtitle={guild.description}
      badges={featureBadges
        .filter(([feature]) => guild.features.includes(feature))
        .map(([, name]) => name)}
    />
  ) : inviter ? (
    <LookupHeader
      iconUrl={userAvatar(inviter)}
      title={
        invite.type === 1
          ? invite.channel?.name ?? "Group DM"
          : inviter.global_name ?? inviter.username
      }
      subtitle={invite.type === 1 ? "Group DM invite" : "Friend invite"}
    />
  ) : null;

  return (
    <div className="space-y-6">
      {header}
      <LookupRows rows={rows} />
    </div>
  );
}
