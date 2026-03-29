import { describe, expect, it } from 'vitest';
import { inCoverage, normalizeAddressInput } from './address';

describe('address utils', () => {
  it('normalizes common street abbreviations', () => {
    const v = normalizeAddressInput('123 Main St.', 'New York', 'NY', '10001');
    expect(v).toContain('street');
  });

  it('checks coverage by postal code', () => {
    expect(inCoverage('10001')).toBe(true);
    expect(inCoverage('99999')).toBe(false);
  });
});
