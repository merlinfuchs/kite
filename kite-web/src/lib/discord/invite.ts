import { AppInstallParams } from "../types/wire.gen";

// Administrator, which servers are asked for until the app's installation
// settings choose other permissions.
export const defaultGuildPermissions = "8";

// Used while the app's installation settings aren't known.
const defaultInstallParams: Record<"guild" | "user", AppInstallParams> = {
  guild: {
    scopes: ["bot", "applications.commands"],
    permissions: defaultGuildPermissions,
  },
  user: { scopes: ["applications.commands"], permissions: "0" },
};

export function getAppInviteUrl(
  appId: string,
  type: "guild" | "user" = "guild",
  params?: AppInstallParams | null
) {
  const install = params ?? defaultInstallParams[type];

  const query = new URLSearchParams({ client_id: appId });
  query.set("integration_type", type === "guild" ? "0" : "1");
  query.set("scope", install.scopes.join(" "));
  if (install.scopes.includes("bot")) {
    query.set("permissions", install.permissions || "0");
  }
  return `https://discord.com/oauth2/authorize?${query}`;
}
