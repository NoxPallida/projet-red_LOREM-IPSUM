package tui

import (
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

// Key dit quelle touche logique a ete appuyee
// On ne renvoie jamais l'octet brut : l'appelant teste juste K
type Key int

const (
	KeyRune      Key = iota // une lettre / chiffre / symbole, lisible dans R
	KeyUp                   // fleche haut (ESC [ A)
	KeyDown                 // fleche bas (ESC [ B)
	KeyLeft                 // fleche gauche (ESC [ D)
	KeyRight                // fleche droite (ESC [ C)
	KeyEnter                // ENTREE (\r en raw, \n en cooked)
	KeyEsc                  // ECHAP seul
	KeyBackspace            // retour arriere (DEL 0x7f ou BS 0x08)
)

// Event c'est UNE touche decodee, prete a l'emploi
// On lit K pour savoir quoi faire R n'a un sens que si K == KeyRune
// Exemple d'usage dans un menu :
//
//	ev, _ := ReadKey(in)
//	if ev.K == KeyUp { ... }
//	if ev.K == KeyRune { fmt.Println(ev.R) }
type Event struct {
	K Key
	R rune
}

// RuneEvent c'est juste un raccourci pour fabriquer un Event lettre
// Au lieu d'ecrire Event{K: KeyRune, R: 'a'} partout,
// on ecrit RuneEvent('a') Meme resultat, une ligne au lieu de deux
func RuneEvent(r rune) Event { return Event{K: KeyRune, R: r} }

// escDelay c'est le temps qu'on attend les octets suivants apres ESC
// Une fleche envoie ses 3 octets d'un coup, donc ils sont deja la
// Un ECHAP seul n'envoie rien d'autre : au bout du delai on rend KeyEsc
// au lieu de rester bloque a attendre pour toujours
const escDelay = 25 * time.Millisecond

// ReadKey lit UNE touche depuis r et la decode en Event
// Touches reconnues : \r ou \n -> Enter, ESPACE -> Rune ' ',
// 0x7f / 0x08 -> Backspace, ESC seul -> Esc,
// ESC [ A/B/C/D -> fleches, tout le reste -> Rune (UTF-8 gere)
//
// Regles a connaitre :
// - '\r' et '\n' donnent chacun Enter. Les terminaux n'envoient
// qu'un seul des deux par appui, donc pas de double ENTREE en vrai.
// (Un fichier Windows "\r\n" donnerait deux Enter : cas rare, assume.)
// - AUCUNE touche ne veut dire "quitter" ici : 'q' rend RuneEvent('q')
// comme une lettre normale, c'est le jeu qui decide (ex : Q = aller
// a l'ouest). ESPACE rend RuneEvent(' ') : c'est le dialogue qui
// s'en sert pour passer, pas le moteur.
// - Les fleches arrivent en 3 octets (ESC [ A...) : ZQSD, ce sont de
// simples lettres, elles ne passent jamais par readEsc.
// - Pas de bufio ici : on lit octet par octet avec io.ReadFull,
// donc aucun octet lu en trop, rien a "remettre" avec UnreadByte
func ReadKey(r io.Reader) (Event, error) {
	b, err := readByte(r)
	if err != nil {
		return Event{}, err
	}
	switch b {
	case '\r', '\n':
		return Event{K: KeyEnter}, nil
	case 0x7f, 0x08:
		return Event{K: KeyBackspace}, nil
	case 0x1b:
		return readEsc(r), nil
	}
	if b < 0x80 {
		// Lettre ASCII sur un seul octet, cas le plus courant
		return RuneEvent(rune(b)), nil
	}
	// Lettre UTF-8 sur plusieurs octets (ex : 'é' = 2 octets)
	// On lit la suite d'un coup : les octets arrivent ensemble
	return readRune(b, r)
}

// readEsc decode ce qui suit un octet ESC.
// Soit c'est une fleche (ESC [ A/B/C/D), soit c'etait juste ECHAP,
// soit une sequence CSI/SS3 etendue (pavedown, numpad...).
// Une sequence non reconnue est avalee (KeyNone), JAMAIS prise pour ECHAP.
func readEsc(r io.Reader) Event {
	b2, ok := readByteSoon(r)
	if !ok {
		return Event{K: KeyEsc}
	}
	if b2 == 'O' {
		b3, ok := readByteSoon(r)
		if !ok {
			return Event{K: KeyNone}
		}
		if b3 >= 'p' && b3 <= 'y' {
			return RuneEvent(rune('0' + (b3 - 'p')))
		}
		if b3 == 'M' {
			return Event{K: KeyEnter}
		}
		return Event{K: KeyNone}
	}
	if b2 != '[' {
		return Event{K: KeyNone}
	}
	var body []byte
	for len(body) < 24 {
		b, ok := readByteSoon(r)
		if !ok {
			return Event{K: KeyNone}
		}
		if (b >= '0' && b <= '9') || b == ';' || b == ':' || b == '?' {
			body = append(body, b)
			continue
		}
		if len(body) == 0 {
			switch b {
			case 'A':
				return Event{K: KeyUp}
			case 'B':
				return Event{K: KeyDown}
			case 'C':
				return Event{K: KeyRight}
			case 'D':
				return Event{K: KeyLeft}
			}
		} else {
			s := string(body)
			if s == "1" || strings.HasPrefix(s, "1;") {
				switch b {
				case 'A':
					return Event{K: KeyUp}
				case 'B':
					return Event{K: KeyDown}
				case 'C':
					return Event{K: KeyRight}
				case 'D':
					return Event{K: KeyLeft}
				}
			}
			if b == '~' {
				if s == "3" {
					return Event{K: KeyBackspace}
				}
				return Event{K: KeyNone}
			}
			if b == 'u' || b == 'U' {
				ev, _, _ := parseKittyU(body)
				return ev
			}
		}
		return Event{K: KeyNone}
	}
	return Event{K: KeyNone}
}

// readByte lit UN seul octet, en bloquant jusqu'a l'avoir
// io.ReadFull ne lit jamais en trop, contrairement a bufio
// qui remplit son tampon d'avance (d'ou l'ancien UnreadByte)
func readByte(r io.Reader) (byte, error) {
	var buf [1]byte
	_, err := io.ReadFull(r, buf[:])
	return buf[0], err
}

// readByteSoon lit un octet mais abandonne apres escDelay
// true = octet recu, false = delai depasse ou erreur de lecture
// La goroutine en attente utilise un canal taille 1 : elle finit
// toute seule a la prochaine touche, rien a nettoyer
func readByteSoon(r io.Reader) (byte, bool) {
	type res struct {
		b   byte
		err error
	}
	ch := make(chan res, 1)
	go func() {
		var buf [1]byte
		_, err := io.ReadFull(r, buf[:])
		ch <- res{buf[0], err}
	}()
	select {
	case o := <-ch:
		if o.err != nil {
			return 0, false
		}
		return o.b, true
	case <-time.After(escDelay):
		return 0, false
	}
}

// readRune finit de lire une lettre UTF-8 dont b est le premier octet
// utf8Size dit combien d'octets la lettre prend au total
func readRune(b byte, r io.Reader) (Event, error) {
	n := utf8Size(b)
	var buf [4]byte
	buf[0] = b
	if n > 1 {
		if _, err := io.ReadFull(r, buf[1:n]); err != nil {
			return Event{}, err
		}
	}
	ch, _ := utf8.DecodeRune(buf[:n])
	return RuneEvent(ch), nil
}

// utf8Size donne la taille en octets d'une lettre UTF-8
// d'apres son premier octet. 1 par defaut si octet invalide
func utf8Size(b byte) int {
	switch {
	case b>>5 == 0x6: // 110xxxxx : 2 octets (é, è, ...)
		return 2
	case b>>4 == 0xE: // 1110xxxx : 3 octets (—, …)
		return 3
	case b>>3 == 0x1E: // 11110xxx : 4 octets (emoji)
		return 4
	default:
		return 1
	}
}
