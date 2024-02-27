export function userAvatarUrl(
  user: { id: string; discriminator: string | null; avatar: string | null },
  size: number = 128
) {
  if (user.avatar) {
    return `https://cdn.discordapp.com/avatars/${user.id}/${user.avatar}.png?size=${size}`;
  } else {
    let defaultAvatar: number | BigInt =
      parseInt(user.discriminator || "0") % 5;
    if (!user.discriminator || user.discriminator === "0") {
      defaultAvatar = (BigInt(user.id) >> BigInt(22)) % BigInt(6);
    }

    return `https://cdn.discordapp.com/embed/avatars/${defaultAvatar}.png?size=${size}`;
  }
}

export function guildIconUrl(
  guild: { id: string; icon: string | null },
  size: number = 128
) {
  if (guild.icon) {
    return `https://cdn.discordapp.com/icons/${guild.id}/${guild.icon}.png?size=${size}`;
  } else {
    return null;
  }
}
