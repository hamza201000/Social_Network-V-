// formatDate turns a backend timestamp like "2026-11-12T19:23:55Z" into a
// readable local date such as "Nov 12, 2026, 7:23 PM".
export function formatDate(value?: string | null): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}
