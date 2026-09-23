package tui

// Clavier étendu : comprend en plus les séquences du protocole kitty
// (appui + RELÂCHEMENT des touches), quand le terminal le propose.
//
// Sans kitty : classique, une touche = un appui (délai de répétition
// de l'OS quand on reste appuyé, comme tous les jeux en terminal).
// Avec kitty : le jeu suit les touches TENUES -> répétition immédiate
// pilotée par le jeu, diagonales simultanées vraies.
//
// Détection auto avec repli : Windows (conhost), vieux terminaux,
// pipes : tout marche comme avant, sans rien changer au système.
// Zéro goroutine ici : que des sondages + petites siestes.

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// KeyNone : octets inconnus avalés, ni touche ni erreur.
const KeyNone Key = -1

const (
	kittyQueryWait = 150 * time.Millisecond // réponse maxi à "kitty ?"
	seqByteWait    = 30 * time.Millisecond  // suite d'une séquence d'échappement
	firstByteWait  = 15 * time.Millisecond  // premier octet (événements fantômes Windows)
	pollSleep      = 2 * time.Millisecond   // pause entre deux sondages
)

// errTimeout : délai dépassé, interne, jamais exposé tel quel.
var errTimeout = errors.New("tui: délai lecture")

// fdReader : tout fichier avec un descripteur (os.Stdin...).
type fdReader interface{ Fd() uintptr }

// readerAvail dit si lire ne bloquera (probablement) pas.
// Sans descripteur (tests) : toujours vrai, la lecture tranche.
func readerAvail(r io.Reader) bool {
	f, ok := r.(fdReader)
	if !ok {
		return true
	}
	return keyAvail(f.Fd())
}

// readByteWait lit UN octet, en attendant jusqu'à timeout, sans goroutine.
func readByteWait(r io.Reader, timeout time.Duration) (byte, error) {
	deadline := time.Now().Add(timeout)
	for {
		if readerAvail(r) {
			var buf [1]byte
			n, err := r.Read(buf[:])
			if err != nil {
				return 0, err // EOF / cassé
			}
			if n == 1 {
				return buf[0], nil
			}
		}
		if time.Now().After(deadline) {
			return 0, errTimeout
		}
		time.Sleep(pollSleep)
	}
}

// byteSoon pareil, mais faux si délai OU entrée fermée :
// une séquence incomplète est jetée, jamais une erreur.
func byteSoon(r io.Reader, timeout time.Duration) (byte, bool) {
	b, err := readByteWait(r, timeout)
	if err != nil {
		return 0, false
	}
	return b, true
}

// PollKey lit UNE touche sans jamais bloquer.
// have=false : rien à lire pour l'instant (rappeler plus tard).
// have=true : ev valide (KeyNone = bruit avalé, à ignorer).
// err != nil : entrée fermée/cassée -> quitter proprement.
// rel=true : RELÂCHEMENT (kitty seulement), la touche n'est plus tenue.
func PollKey(r io.Reader) (ev Event, rel bool, have bool, err error) {
	if !readerAvail(r) {
		return Event{}, false, false, nil
	}
	b, err := readByteWait(r, firstByteWait)
	if err != nil {
		if err == errTimeout {
			return Event{}, false, false, nil // fantôme : rien en fait
		}
		return Event{}, false, false, err
	}
	ev, rel, err = parseKeyByte(b, r)
	if err != nil {
		return Event{}, false, false, err
	}
	return ev, rel, true, nil
}

// parseKeyByte décode à partir du premier octet (déjà lu).
func parseKeyByte(b byte, r io.Reader) (Event, bool, error) {
	switch b {
	case '\r', '\n':
		return Event{K: KeyEnter}, false, nil
	case 0x7f, 0x08:
		return Event{K: KeyBackspace}, false, nil
	case 0x1b:
		return parseEscKey(r)
	}
	if b < 0x80 {
		return RuneEvent(rune(b)), false, nil
	}
	return parseUTF8Key(b, r)
}

// parseUTF8Key finit une lettre multi-octets (é, —, emoji...).
func parseUTF8Key(b byte, r io.Reader) (Event, bool, error) {
	n := utf8Size(b)
	var buf [4]byte
	buf[0] = b
	for i := 1; i < n; i++ {
		c, ok := byteSoon(r, seqByteWait)
		if !ok {
			return Event{K: KeyNone}, false, nil // coupée : jetée
		}
		buf[i] = c
	}
	ch, _ := utf8.DecodeRune(buf[:n])
	return RuneEvent(ch), false, nil
}

// parseEscKey décode après un octet ESC : vrai ECHAP seul, flèche,
// ou séquence kitty (bouton relâché...). Inconnu = avalé (KeyNone).
func parseEscKey(r io.Reader) (Event, bool, error) {
	b2, ok := byteSoon(r, seqByteWait)
	if !ok {
		return Event{K: KeyEsc}, false, nil // vrai ECHAP seul (ou entrée finie)
	}
	if b2 != '[' {
		return Event{K: KeyNone}, false, nil // Alt+lettre ou bruit : avalé
	}
	var body []byte
	for len(body) < 24 {
		b, ok := byteSoon(r, seqByteWait)
		if !ok {
			return Event{K: KeyNone}, false, nil // coupée : jetée
		}
		if b >= '0' && b <= '9' || b == ';' || b == ':' || b == '?' {
			body = append(body, b)
			continue
		}
		return finishEscKey(b, body)
	}
	return Event{K: KeyNone}, false, nil
}

// finishEscKey : l'octet final décide (u = kitty, A/B/C/D = flèches).
func finishEscKey(final byte, body []byte) (Event, bool, error) {
	if final == 'u' || final == 'U' {
		return parseKittyU(body)
	}
	if final == 'A' || final == 'B' || final == 'C' || final == 'D' {
		// Flèches : "" (classique) ou "1[;mods]" (précisé).
		s := string(body)
		if s == "" || s == "1" || strings.HasPrefix(s, "1;") {
			switch final {
			case 'A':
				return Event{K: KeyUp}, false, nil
			case 'B':
				return Event{K: KeyDown}, false, nil
			case 'C':
				return Event{K: KeyRight}, false, nil
			case 'D':
				return Event{K: KeyLeft}, false, nil
			}
		}
	}
	return Event{K: KeyNone}, false, nil
}

// parseKittyU : "ESC [ <num> [;mods[:event]] u".
// num = code de la touche (lettre = son code ASCII), event 3 = relâchement.
// Le reste (flèches numérotées...) est avalé : leurs appuis classiques
// suffisent, et l'expiration du jeu évite les touches coincées.
func parseKittyU(body []byte) (Event, bool, error) {
	s := strings.TrimPrefix(string(body), "?")
	event := 1
	if i := strings.IndexByte(s, ':'); i >= 0 {
		if n, err := strconv.Atoi(s[i+1:]); err == nil {
			event = n
		}
		s = s[:i]
	}
	if i := strings.IndexByte(s, ';'); i >= 0 {
		s = s[:i]
	}
	num, err := strconv.Atoi(s)
	if err != nil {
		return Event{K: KeyNone}, false, nil
	}
	rel := event == 3
	switch {
	case num >= 'a' && num <= 'z', num >= 'A' && num <= 'Z', num == ' ':
		return RuneEvent(rune(num)), rel, nil
	}
	return Event{K: KeyNone}, false, nil
}

// QueryKitty demande au terminal s'il envoie appui+relâchement.
// true = oui (bit 2) ; false/rien/EOF = non ou pipe -> classique.
// Les touches tapées pendant la question (150ms maxi) sont ignorées.
func QueryKitty(r io.Reader, w io.Writer) bool {
	if _, err := io.WriteString(w, "\x1b[?u"); err != nil {
		return false
	}
	deadline := time.Now().Add(kittyQueryWait)
	var buf []byte
	for {
		if readerAvail(r) {
			var tmp [32]byte
			n, err := r.Read(tmp[:])
			if err != nil {
				return false
			}
			buf = append(buf, tmp[:n]...)
			if i := strings.Index(string(buf), "\x1b[?"); i >= 0 {
				rest := string(buf)[i+3:]
				if j := strings.IndexByte(rest, 'u'); j >= 0 {
					if n, err := strconv.Atoi(rest[:j]); err == nil {
						return n&2 != 0
					}
					return false
				}
			}
			if len(buf) > 64 {
				return false
			}
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(pollSleep)
	}
}

// PushKitty demande l'envoi des types d'événements (appui/répétition/relâchement).
func PushKitty(w io.Writer) {
	_, _ = io.WriteString(w, "\x1b[>2u")
}

// PopKitty rend la main (retour au mode d'avant).
func PopKitty(w io.Writer) {
	_, _ = io.WriteString(w, "\x1b[<u")
}
