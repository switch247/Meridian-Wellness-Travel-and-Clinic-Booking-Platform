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

const DEFAULT_REGIONS = ["10001", "10002", "10003", "60601", "90001"];
let coverageSet = new Set<string>(DEFAULT_REGIONS);

export function setCoverageRegions(regions: string[]) {
  const normalized = regions.map((r) => r.trim()).filter((r) => r !== "");
  if (normalized.length === 0) {
    coverageSet = new Set(DEFAULT_REGIONS);
    return;
  }
  coverageSet = new Set(normalized);
}

export function inCoverage(postalCode: string): boolean {
  return coverageSet.has(postalCode.trim());
}
