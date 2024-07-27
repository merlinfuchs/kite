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
