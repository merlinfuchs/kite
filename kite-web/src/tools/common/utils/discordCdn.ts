export default function userAvatar({
  id,
  avatar,
}: {
  id: string;
  avatar?: string | null;
}) {
  if (avatar) {
    return `https://cdn.discordapp.com/avatars/${id}/${avatar}.png`;
  }
  return `https://cdn.discordapp.com/embed/avatars/${
    (BigInt(id) >> BigInt(22)) % BigInt(6)
  }.png`;
}

export function discordAvatarUrl(url: string) {
  return url.replace(
    /^(https:\/\/cdn\.discordapp\.com\/avatars\/[^/]+\/a_[^/.]+)\.(?:png|webp|jpe?g)(?=[?#]|$)/i,
    "$1.gif"
  );
}

export function discordEmojiUrl(id: string, animated?: boolean) {
  return `https://cdn.discordapp.com/emojis/${id}.${animated ? "gif" : "webp"}`;
}

const cdn = "https://cdn.discordapp.com";

// Animated images have hashes that start with a_.
function imageUrl(path: string, hash: string, size: number) {
  const ext = hash.startsWith("a_") ? "gif" : "png";
  return `${cdn}/${path}/${hash}.${ext}?size=${size}`;
}

export function avatarUrl(userId: string, hash: string, size = 256) {
  return imageUrl(`avatars/${userId}`, hash, size);
}

export function bannerUrl(id: string, hash: string, size = 1024) {
  return imageUrl(`banners/${id}`, hash, size);
}

export function guildIconUrl(guildId: string, hash: string, size = 256) {
  return imageUrl(`icons/${guildId}`, hash, size);
}

export function guildSplashUrl(guildId: string, hash: string, size = 1024) {
  return `${cdn}/splashes/${guildId}/${hash}.png?size=${size}`;
}

export function appIconUrl(appId: string, hash: string, size = 256) {
  return `${cdn}/app-icons/${appId}/${hash}.png?size=${size}`;
}

export function avatarDecorationUrl(asset: string) {
  return `${cdn}/avatar-decoration-presets/${asset}.png?size=240&passthrough=true`;
}

export function guildTagBadgeUrl(guildId: string, badge: string) {
  return `${cdn}/guild-tag-badges/${guildId}/${badge}.png?size=32`;
}
