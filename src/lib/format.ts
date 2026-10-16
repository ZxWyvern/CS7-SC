// Pure TS, no backend, no DOM. Untuk unit test CI.
export function formatNama(nama: string): string {
  return nama.trim().replace(/\s+/g, ' ');
}

export function isHariSekolah(day: number): boolean {
  // 0=Minggu, 6=Sabtu libur
  return day >= 1 && day <= 5;
}
