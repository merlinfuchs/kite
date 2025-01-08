export function getAppInviteUrl(
  appId: string,
  type: "guild" | "user" = "guild"
) {
  if (type === "user") {
    return `https://discord.com/oauth2/authorize?client_id=${appId}&integration_type=1&scope=applications.commands`;
  }

  return `https://discord.com/oauth2/authorize?client_id=${appId}&integration_type=0&scope=bot%20applications.commands&permissions=8`;
}
