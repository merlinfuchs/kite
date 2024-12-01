export function getAppInviteUrl(
  appId: string,
  type: "guild" | "user" = "guild"
) {
  const integrationType = type === "user" ? 1 : 0;

  return `https://discord.com/oauth2/authorize?client_id=${appId}&integration_type=${integrationType}&scope=bot%20applications.commands&permissions=8&guild_id=${integrationType}`;
}
