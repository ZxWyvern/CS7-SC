import { describe, expect, it } from 'vitest';
import { hasErrors, safeNextPath, validateLogin } from './login';

describe('validateLogin', () => {
  it('rejects empty fields', () => {
    expect(validateLogin('', '')).toEqual({
      email: 'Isi email.',
      password: 'Isi kata sandi.',
    });
  });

  it('accepts valid credentials', () => {
    expect(validateLogin('siswa@contoh.id', 'rahasia')).toEqual({});
  });

  it('rejects malformed emails', () => {
    expect(validateLogin('bukan-email', 'rahasia')).toEqual({
      email: 'Masukkan email yang valid.',
    });
  });

  it('rejects overlong passwords', () => {
    expect(validateLogin('siswa@contoh.id', 'a'.repeat(129))).toEqual({
      password: 'Kata sandi terlalu panjang.',
    });
  });
});

describe('hasErrors', () => {
  it('reports whether any error exists', () => {
    expect(hasErrors({})).toBe(false);
    expect(hasErrors({ password: 'Isi kata sandi.' })).toBe(true);
  });
});

describe('safeNextPath', () => {
  it('accepts local paths', () => {
    expect(safeNextPath('/akun')).toBe('/akun');
  });

  it('falls back for external or missing paths', () => {
    expect(safeNextPath(null)).toBe('/akun');
    expect(safeNextPath('https://situs-jahat.example')).toBe('/akun');
    expect(safeNextPath('//situs-jahat.example')).toBe('/akun');
    expect(safeNextPath('/\\situs-jahat.example')).toBe('/akun');
  });
});
