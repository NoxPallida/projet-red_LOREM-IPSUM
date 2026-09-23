package audio

// Musique de fond : OST compressee (mono 22kHz 24k), embarquee
// dans le binaire. Au lancement on la charge en memoire (jamais
// de fichier temporaire) et on joue en streaming direct depuis
// la memoire.

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/hajimehoshi/go-mp3"

	"runa/assets"
)

// MP3 embarque via assets/embed.go (embed interdit les "..").
var ostMP3 = assets.OST

// Track c'est la piste prete a jouer : le MP3 en memoire.
type Track struct {
	mp3data []byte
	dec     *mp3.Decoder
	mu      sync.Mutex
}

// Load charge le MP3 embarque en memoire et prepare le decodeur.
// Rien n'est jamais ecrit sur disque.
func Load() (*Track, error) {
	if len(ostMP3) < 4 {
		return nil, fmt.Errorf("audio: donnees embarquees vides")
	}
	dec, err := mp3.NewDecoder(bytes.NewReader(ostMP3))
	if err != nil {
		return nil, fmt.Errorf("audio: mp3: %w", err)
	}
	return &Track{mp3data: ostMP3, dec: dec}, nil
}

// SampleRate rend le taux de la piste (22050 pour OST).
func (t *Track) SampleRate() int {
	return t.dec.SampleRate()
}

// read remplit p depuis le decodeur, en bouclant la piste.
// Appele depuis le thread audio : mutex obligatoire.
func (t *Track) read(p []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	off := 0
	for off < len(p) {
		n, err := t.dec.Read(p[off:])
		off += n
		if err != nil {
			dec, derr := mp3.NewDecoder(bytes.NewReader(t.mp3data))
			if derr != nil {
				for i := off; i < len(p); i++ {
					p[i] = 0
				}
				return
			}
			t.dec = dec
		}
		if n == 0 && err == nil {
			break
		}
	}
}

// Player c'est la lecture en cours. Stop coupe le son et libere tout.
type Player struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device
}

// Play demarre la lecture en boucle. Sans carte son (CI, container)
// ca renvoie une erreur propre : le jeu continue sans musique.
func (t *Track) Play() (*Player, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(string) {})
	if err != nil {
		return nil, fmt.Errorf("audio: pas de sortie son (%v)", err)
	}
	cfg := malgo.DefaultDeviceConfig(malgo.Playback)
	cfg.Playback.Format = malgo.FormatS16
	cfg.Playback.Channels = 2
	cfg.SampleRate = uint32(t.SampleRate())
	cfg.Alsa.NoMMap = 1
	dev, err := malgo.InitDevice(ctx.Context, cfg, malgo.DeviceCallbacks{
		Data: func(out, _ []byte, _ uint32) { t.read(out) },
	})
	if err != nil {
		ctx.Uninit()
		ctx.Free()
		return nil, fmt.Errorf("audio: device (%v)", err)
	}
	if err := dev.Start(); err != nil {
		dev.Uninit()
		ctx.Uninit()
		ctx.Free()
		return nil, fmt.Errorf("audio: start (%v)", err)
	}
	return &Player{ctx: ctx, device: dev}, nil
}

// Stop coupe et libere. Appels multiples sans danger.
func (p *Player) Stop() {
	if p == nil || p.device == nil {
		return
	}
	p.device.Uninit()
	p.device = nil
	if p.ctx != nil {
		p.ctx.Uninit()
		p.ctx.Free()
		p.ctx = nil
	}
}

// StartOST charge + joue l'OST. Ne plante jamais le jeu : en cas
// d'echec (pas de son) ca previent sur stderr et rend un stop vide.
// Usage : stop := audio.StartOST(); defer stop()
func StartOST() func() {
	noop := func() {}
	track, err := Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "audio:", err)
		return noop
	}
	player, err := track.Play()
	if err != nil {
		fmt.Fprintln(os.Stderr, "audio:", err)
		return noop
	}
	return player.Stop
}
