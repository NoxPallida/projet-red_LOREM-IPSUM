package tui

import (
	"io"
	"strings"
	"testing"
)

// chunkReader rejoue les octets par petits bouts (séquences coupées).
type chunkReader struct {
	data []byte
	n    int
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if c.n >= len(c.data) {
		return 0, io.EOF
	}
	m := 2
	if m > len(p) {
		m = len(p)
	}
	if m > len(c.data)-c.n {
		m = len(c.data) - c.n
	}
	copy(p, c.data[c.n:c.n+m])
	c.n += m
	return m, nil
}

func TestPollKeyLegacy(t *testing.T) {
	cases := []struct {
		in  string
		k   Key
		r   rune
		rel bool
	}{
		{"z", KeyRune, 'z', false},
		{" ", KeyRune, ' ', false},
		{"\r", KeyEnter, 0, false},
		{"\x1b[A", KeyUp, 0, false},
		{"\x1b[B", KeyDown, 0, false},
		{"\x1b[C", KeyRight, 0, false},
		{"\x1b[D", KeyLeft, 0, false},
		{"\x1b[1;5A", KeyUp, 0, false},
		{"\x1b", KeyEsc, 0, false},
		{"é", KeyRune, 'é', false},
	}
	for _, tc := range cases {
		ev, rel, have, err := PollKey(strings.NewReader(tc.in))
		if err != nil || !have {
			t.Errorf("PollKey(%q): err=%v have=%v", tc.in, err, have)
			continue
		}
		if ev.K != tc.k || ev.R != tc.r || rel != tc.rel {
			t.Errorf("PollKey(%q) = %+v rel=%v, want K=%v R=%q", tc.in, ev, rel, tc.k, tc.r)
		}
	}
}

func TestPollKeyKitty(t *testing.T) {
	cases := []struct {
		in  string
		k   Key
		r   rune
		rel bool
	}{
		{"\x1b[122u", KeyRune, 'z', false},     // appui forme-u
		{"\x1b[122;1:1u", KeyRune, 'z', false}, // appui explicite
		{"\x1b[122;1:2u", KeyRune, 'z', false}, // répétition = appui
		{"\x1b[122;1:3u", KeyRune, 'z', true},  // RELÂCHEMENT
		{"\x1b[90;5:1u", KeyRune, 'Z', false},  // majuscule + modifieurs
		{"\x1b[32;1:3u", KeyRune, ' ', true},   // espace relâché
		{"\x1b[999;1:3u", KeyNone, 0, false},   // numéro inconnu : avalé
		{"\x1b[?2u", KeyNone, 0, false},        // réponse parasite : avalée
		{"\x1b[X", KeyNone, 0, false},          // final inconnu : avalé
	}
	for _, tc := range cases {
		r := &chunkReader{data: []byte(tc.in)} // en morceaux : comme un vrai terminal
		ev, rel, have, err := PollKey(r)
		if err != nil || !have {
			t.Errorf("PollKey(%q): err=%v have=%v", tc.in, err, have)
			continue
		}
		if ev.K != tc.k || ev.R != tc.r || rel != tc.rel {
			t.Errorf("PollKey(%q) = %+v rel=%v, want K=%v R=%q rel=%v", tc.in, ev, rel, tc.k, tc.r, tc.rel)
		}
	}
}

func TestQueryKitty(t *testing.T) {
	var out strings.Builder
	if !QueryKitty(strings.NewReader("\x1b[?2u"), &out) {
		t.Error("QueryKitty devrait dire oui pour flags=2")
	}
	if out.String() != "\x1b[?u" {
		t.Errorf("requête = %q, want CSI ? u", out.String())
	}
	var out2 strings.Builder
	if QueryKitty(strings.NewReader("\x1b[?0u"), &out2) {
		t.Error("QueryKitty devrait dire non pour flags=0")
	}
	var out3 strings.Builder
	if QueryKitty(strings.NewReader(""), &out3) {
		t.Error("QueryKitty devrait dire non sur EOF")
	}
}
