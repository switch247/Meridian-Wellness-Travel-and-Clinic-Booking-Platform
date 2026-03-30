import { describe, expect, it } from 'vitest';
import { canAccess } from '../src/app/roleMatrix';

describe('role matrix', () => {
  it('allows matching role access', () => {
    expect(canAccess(['traveler'], ['traveler'])).toBe(true);
  });

  it('denies non-matching role access', () => {
    expect(canAccess(['coach'], ['admin'])).toBe(false);
  });
});
