// Settings that take a list hold one item per line, or a placeholder that
// holds a list, which the service reads like listOrLines.

// describeList says how many items a list setting has, or that a placeholder
// fills it in, which can't be counted before the flow runs.
export function describeList(value: string, noun: string): string {
  if (/^\{\{[^{}]*\}\}$/.test(value.trim())) {
    return "Filled in by the placeholder";
  }

  const count = value.split("\n").filter((line) => line.trim()).length;
  return `${count} ${noun}${count === 1 ? "" : "s"}`;
}
