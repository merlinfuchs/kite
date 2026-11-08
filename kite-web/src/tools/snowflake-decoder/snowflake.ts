// Milliseconds of the first second of 2015, where Discord IDs start counting.
const discordEpoch = BigInt(1420070400000);

export function decodeSnowflake(value: string) {
  if (!/^\d{17,20}$/.test(value)) return null;

  const id = BigInt(value);
  const timestamp = Number((id >> BigInt(22)) + discordEpoch);
  // An ID from the future isn't a real Discord ID.
  if (timestamp > Date.now() + 24 * 60 * 60 * 1000) return null;

  return {
    timestamp,
    workerId: Number((id >> BigInt(17)) & BigInt(0x1f)),
    processId: Number((id >> BigInt(12)) & BigInt(0x1f)),
    increment: Number(id & BigInt(0xfff)),
  };
}
