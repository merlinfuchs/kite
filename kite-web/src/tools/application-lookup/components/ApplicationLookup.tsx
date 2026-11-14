import LookupForm from "@/tools/common/components/LookupForm";
import {
  formatCreated,
  LookupHeader,
  LookupLink,
  LookupRow,
  LookupRows,
} from "@/tools/common/components/LookupResult";
import { appIconUrl } from "@/tools/common/utils/discordCdn";
import { useLookup } from "@/tools/common/hooks/useLookup";
import {
  discordUnreachable,
  fetchDiscord,
  parseLookupId,
} from "@/tools/common/utils/lookup";
import { decodeSnowflake } from "@/tools/snowflake-decoder/snowflake";

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

// Each privileged intent has a flag for verified apps and one for apps in
// fewer than 100 servers, which don't need to be approved for it.
const intents: [string, number, number][] = [
  ["Presence", 1 << 12, 1 << 13],
  ["Server Members", 1 << 14, 1 << 15],
  ["Message Content", 1 << 18, 1 << 19],
];

function fetchApplication(id: string) {
  return fetchDiscord<ApplicationData>(
    `/applications/${id}/rpc`,
    "There is no Discord app with this ID."
  );
}

export default function ApplicationLookup() {
  const { result, error, loading, initialValue, lookup } = useLookup({
    name: "application",
    param: "id",
    parse: parseLookupId,
    invalid:
      "This isn't a Discord app ID. IDs are numbers with 17 to 20 digits.",
    fetch: fetchApplication,
    unreachable: discordUnreachable,
    staleTime: 10 * 60 * 1000,
  });

  return (
    <div className="space-y-6">
      <LookupForm
        id="app-id"
        label="Application ID"
        placeholder="235148962103951360"
        initialValue={initialValue}
        key={initialValue}
        loading={loading}
        error={error}
        onLookup={lookup}
      />
      {result?.success && <ApplicationInfo app={result.data} />}
    </div>
  );
}

function ApplicationInfo({ app }: { app: ApplicationData }) {
  const created = decodeSnowflake(app.id);

  const badges: string[] = [];
  if (app.is_verified) badges.push("Verified");
  if (app.is_discoverable) badges.push("In the App Directory");
  if (app.is_monetized) badges.push("Has premium");

  // Without install settings, Discord needs the scopes in the link.
  const inviteUrl =
    app.custom_install_url ??
    `https://discord.com/oauth2/authorize?client_id=${app.id}` +
      (app.install_params ? "" : "&scope=bot%20applications.commands");

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
