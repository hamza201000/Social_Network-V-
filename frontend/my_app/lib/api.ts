export const API_BASE_URL =
  (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080").replace(/\/+$/, "");

export function imageUrl(filename?: string | null): string {
  if (!filename) return "";
  return `${API_BASE_URL}/api/images/${encodeURIComponent(filename)}`;
}
