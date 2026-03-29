const STREET_MAP: Record<string, string> = {
  st: "street",
  rd: "road",
  ave: "avenue",
  blvd: "boulevard",
  dr: "drive",
  ln: "lane",
};

export function normalizeAddressInput(line1: string, city: string, state: string, postalCode: string): string {
  const normalizedLine = line1
    .toLowerCase()
    .replace(/[.,]/g, " ")
    .split(/\s+/)
    .filter(Boolean)
    .map((part) => STREET_MAP[part] || part)
    .join(" ");
  return `${normalizedLine}|${city.trim().toLowerCase()}|${state.trim().toLowerCase()}|${postalCode.trim()}`;
}

export function inCoverage(postalCode: string): boolean {
  return ["10001", "10002", "10003", "60601", "90001"].includes(postalCode.trim());
}
