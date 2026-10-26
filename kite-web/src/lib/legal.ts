// Date of the current Terms of Service and Privacy Policy. Users who signed up
// before it are told about the update in the dashboard.
export const legalUpdatedAt = new Date("2026-10-26T00:00:00Z");

export function formatLegalDate(date: Date) {
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
    timeZone: "UTC",
  });
}
