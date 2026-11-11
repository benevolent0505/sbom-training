const formatter = new Intl.DateTimeFormat("ja-JP", {
  dateStyle: "long",
  timeStyle: "short",
  timeZone: "Asia/Tokyo",
});

export function formatDate(iso: string): string {
  return formatter.format(new Date(iso));
}
