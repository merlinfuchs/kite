export function guildNameAbbreviation(name: string) {
  const words = name.split(" ");

  let res = "";
  for (const word of words) {
    if (word.length === 0) {
      continue;
    }

    res += word[0].toUpperCase();
  }

  return res.slice(0, 3);
}
