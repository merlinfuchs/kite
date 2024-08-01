export function getAppInviteUrl(appId: string) {
  return `https://discord.com/oauth2/authorize?client_id=${appId}&scope=bot%20applications.commands`;
}
