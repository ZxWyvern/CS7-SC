import { describe, expect, it } from 'vitest';
import { formatNama, isHariSekolah } from './format';

describe('formatNama', () => {
  it('rapikan spasi', () => {
    expect(formatNama('  budi   santoso ')).toBe('budi santoso');
  });
});

describe('isHariSekolah', () => {
  it('senin-jumat true, sabtu-minggu false', () => {
    expect(isHariSekolah(1)).toBe(true);
    expect(isHariSekolah(0)).toBe(false);
    expect(isHariSekolah(6)).toBe(false);
  });
});
