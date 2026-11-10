package config

import (
	"strings"
	"testing"
)

func TestLoadTrustedProxyHops(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    int
		wantErr string
	}{
		{name: "tidak diisi berarti nol", raw: "", want: 0},
		{name: "satu hop", raw: "1", want: 1},
		{name: "nol eksplisit", raw: "0", want: 0},
		{name: "delapan batas atas", raw: "8", want: 8},
		{name: "negatif ditolak", raw: "-1", wantErr: "TRUSTED_PROXY_HOPS"},
		{name: "nilai di atas batas ditolak", raw: "9", wantErr: "TRUSTED_PROXY_HOPS"},
		{name: "bukan angka ditolak", raw: "benar", wantErr: "TRUSTED_PROXY_HOPS"},
		{name: "angka plus spasi diterima", raw: " 1 ", want: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://cs7:sandi@127.0.0.1:5432/cs7")
			t.Setenv("APP_ORIGIN", "https://kelas.example.com")
			if tc.raw != "" || tc.wantErr != "" {
				t.Setenv("TRUSTED_PROXY_HOPS", tc.raw)
			}

			cfg, err := Load()
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("TRUSTED_PROXY_HOPS=%q harus gagal, hasil = %v (cfg %+v)", tc.raw, err, cfg)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if cfg.TrustedProxyHops != tc.want {
				t.Fatalf("TrustedProxyHops = %d, mau %d", cfg.TrustedProxyHops, tc.want)
			}
		})
	}
}
