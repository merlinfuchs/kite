import LookupForm from "@/tools/common/components/LookupForm";
import {
  formatCreated,
  LookupHeader,
  LookupLink,
  LookupRow,
  LookupRows,
} from "@/tools/common/components/LookupResult";
import { appIconUrl } from "@/tools/common/utils/discordCdn";
import { parseLookupId } from "@/tools/common/utils/lookup";
import { decodeSnowflake } from "@/tools/snowflake-decoder/snowflake";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

interface ApplicationData {
  id: string;
  name: string;
  icon: string | null;
  description: string;
  flags: number;
  guild_id?: string;
  bot_public?: boolean;
  bot_require_code_grant?: boolean;
  is_verified?: boolean;
  is_monetized?: boolean;
  is_discoverable?: boolean;
  terms_of_service_url?: string;
  privacy_policy_url?: string;
  custom_install_url?: string;
  install_params?: { scopes: string[]; permissions: string };
  tags?: string[];
  bot?: { id: string; username: string; discriminator: string };
}

type ApplicationResponse =
  | { success: true; app: ApplicationData }
  | { success: false; error: string };

// Each privileged intent has a flag for verified apps and one for apps in
// fewer than 100 servers, which don't need to be approved for it.
const intents: [string, number, number][] = [
  ["Presence", 1 << 12, 1 << 13],
  ["Server Members", 1 << 14, 1 << 15],
  ["Message Content", 1 << 18, 1 << 19],
];

async function fetchApplication(id: string): Promise<ApplicationResponse> {
  const res = await fetch(`https://discord.com/api/v10/applications/${id}/rpc`);
  if (res.status === 404) {
    return {
      success: false,
      error: "There is no Discord app with this ID.",
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
  return { success: true, app: await res.json() };
}

export default function ApplicationLookup() {
  const [appId, setAppId] = useState<string | null>(null);
  const [inputError, setInputError] = useState<string>();

  const { data, isFetching } = useQuery({
    queryKey: ["lookup-application", appId],
    queryFn: () => fetchApplication(appId!),
    enabled: !!appId,
    staleTime: 10 * 60 * 1000,
  });

  function lookup(value: string) {
    const id = parseLookupId(value);
    if (!id) {
      setInputError(
        "This isn't a Discord app ID. IDs are numbers with 17 to 20 digits."
      );
      return;
    }
    setInputError(undefined);
    setAppId(id);
  }

  const error = inputError ?? (data && !data.success ? data.error : undefined);

  return (
    <div className="space-y-6">
      <LookupForm
        id="app-id"
        label="Application ID"
        placeholder="235148962103951360"
        loading={isFetching}
        error={error}
        onLookup={lookup}
      />
      {data?.success && !inputError && <ApplicationInfo app={data.app} />}
    </div>
  );
}

function ApplicationInfo({ app }: { app: ApplicationData }) {
  const created = decodeSnowflake(app.id);

  const badges: string[] = [];
  if (app.is_verified) badges.push("Verified");
  if (app.is_discoverable) badges.push("In the App Directory");
  if (app.is_monetized) badges.push("Has premium");

  const inviteUrl =
    app.custom_install_url ??
    `https://discord.com/oauth2/authorize?client_id=${app.id}`;

  const rows: LookupRow[] = [
    { label: "Application ID", value: app.id, copy: app.id },
  ];
  if (created) {
    rows.push({ label: "Created", value: formatCreated(created.timestamp) });
  }
  if (app.bot) {
    rows.push({
      label: "Bot",
      value:
        app.bot.discriminator !== "0"
          ? `${app.bot.username}#${app.bot.discriminator}`
          : app.bot.username,
    });
  }
  rows.push({
    label: "Public bot",
    value: app.bot_public
      ? "Yes, anyone can add it"
      : "No, only the owner can add it",
  });
  if (app.bot_require_code_grant) {
    rows.push({ label: "Requires OAuth2 code grant", value: "Yes" });
  }
  rows.push({
    label: "Privileged intents",
    value:
      intents
        .filter(([, verified, limited]) => app.flags & (verified | limited))
        .map(([name]) => name)
        .join(", ") || "None",
  });
  if (app.install_params) {
    rows.push({
      label: "Install scopes",
      value: (
        <span className="font-mono">
          {app.install_params.scopes.join(", ")}
        </span>
      ),
    });
    rows.push({
      label: "Install permissions",
      value: (
        <span className="font-mono">{app.install_params.permissions}</span>
      ),
      copy: app.install_params.permissions,
    });
  }
  if (app.tags?.length) {
    rows.push({ label: "Tags", value: app.tags.join(", ") });
  }
  if (app.guild_id) {
    rows.push({
      label: "Support server ID",
      value: app.guild_id,
      copy: app.guild_id,
    });
  }
  if (app.terms_of_service_url) {
    rows.push({
      label: "Terms of service",
      value: (
        <LookupLink href={app.terms_of_service_url}>
          {app.terms_of_service_url}
        </LookupLink>
      ),
    });
  }
  if (app.privacy_policy_url) {
    rows.push({
      label: "Privacy policy",
      value: (
        <LookupLink href={app.privacy_policy_url}>
          {app.privacy_policy_url}
        </LookupLink>
      ),
    });
  }
  if (app.bot_public) {
    rows.push({
      label: "Invite link",
      value: <LookupLink href={inviteUrl}>{inviteUrl}</LookupLink>,
      copy: inviteUrl,
    });
  }

  return (
    <div className="space-y-6">
      <LookupHeader
        iconUrl={app.icon ? appIconUrl(app.id, app.icon) : null}
        iconShape="rounded"
        title={app.name}
        subtitle={
          app.description ? (
            <span className="whitespace-pre-line">{app.description}</span>
          ) : null
        }
        badges={badges}
      />
      <LookupRows rows={rows} />
    </div>
  );
}
