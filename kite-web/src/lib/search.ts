// Case-insensitive search used by the dashboard lists. Every whitespace
// separated term has to appear in at least one of the given fields, so
// "ban user" matches a command named "user-ban".
export function matchesSearch(
  query: string,
  fields: (string | null | undefined)[]
): boolean {
  const terms = query.toLowerCase().split(/\s+/).filter(Boolean);
  if (terms.length === 0) return true;

  const haystack = fields
    .filter((f): f is string => !!f)
    .join("\n")
    .toLowerCase();

  return terms.every((term) => haystack.includes(term));
}
