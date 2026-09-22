# Moteur TUI

Le moteur est dans `internal/tui/`. Principe : on dessine tout sur un
`Canvas` (mémoire vidéo), puis on l'envoie d'un coup à l'écran.
Boucle typique d'une frame :

```go
c := tui.NewCanvas(w, h)
c.Clear()                    // 1 : effacer
tui.DrawBox(c, 2, 2, 20, 7)  // 2 : dessiner (boites, texte, calques...)
tui.Flush(out, c.String())   // 3 : envoyer (ou FlushStyled si couleurs)
```

A RETENIR :

- `w, h <= 0` = taille lue toute seule (`tui.Size` : vrai terminal,
  sinon `COLUMNS`/`LINES`)
- Dernier dessin gagne, case par case : une boite dessinée en dernier
  cache ce qui est dessous
- `String()` = texte brut (pour les tests), `RenderStyled()` = avec
  couleurs (à envoyer avec `FlushStyled`)
- `ReadKey(in)` rend une touche bloquante (`KeyUp`, `KeyEnter`...) ;
  `PollKey(in)` rend une touche sans bloquer (`have=false` si rien)

## Clavier : ReadKey vs PollKey

- `ReadKey` : bloque jusqu'à la prochaine touche (dialogues, menus).
- `PollKey` : jamais bloquant, rend `(ev, rel, have, err)`.
  `rel=true` = relâchement (protocole kitty seulement) :
  le jeu suit les touches TENUES pour la répétition immédiate.
- Kittty auto-détecté (`QueryKitty`) : avec, `z` tenu = ça avance
  tout seul ; sans, 1 touche = 1 pas. `KeyNone` = bruit avalé, ignorer.
- Touches jeu : ZQSD + flèches (+ diagonales `aecw`), `e` = inventaire,
  `ESPACE`/`ENTREE` = valider/interagir, `ECHAP` = quitter.

## Texte multicolor

Couleur simple, sans balise :

```go
c.WriteStyled(2, 1, "Bonjour", tui.FGRed, "") // texte rouge
```

Balises `/COULEUR/` (16 couleurs : `BLACK RED GREEN YELLOW BLUE MAGENTA
CYAN WHITE GRAY LIGHTRED LIGHTGREEN LIGHTYELLOW LIGHTBLUE LIGHTMAGENTA
LIGHTCYAN BRIGHTWHITE`). La balise **bascule** : 1re fois elle allume,
2e fois identique elle éteint. `/RESET/` (ou `//`) éteint toujours :

```go
// "yo wassup" en rouge :
tui.WriteMarkup(c, 2, 1, "test /RED/yo wassup/RED/ gg bg", "")

// vert puis normal, fond bleu :
tui.WriteMarkup(c, 2, 2, "vie /GREEN/+25/RESET/ pv", tui.BGBlue)
```

Fond coloré sur un rectangle (ex : bandeau) :

```go
tui.FillStyled(c, 0, 0, w, 3, ' ', tui.FGWhite, tui.BGLightRed)
```

## Dégradés

Les couleurs tournent en boucle sur les lettres. Dégradés prêts :
`Gradient("FIRE")`, `OCEAN`, `FOREST`, `SUNSET`, `SHADOW`
```go

// Titre en feu :
tui.WriteGradient(c, 2, 1, "DRAGON ROUGE", "", tui.Gradient("FIRE"))

// Boite au cadre dégradé océan, fond teinté, titre blanc :
tui.DrawBoxGradient(c, x, y, boxW, boxH, "DIALOGUE", tui.Gradient("OCEAN"), tui.BGBlue)
```

## Typewriter (fonctions, pas méthodes)

```go
tw := tui.NewTypewriter("Bienvenue dans RED...")
for !tui.IsDone(tw) {
    tui.Tick(tw, 1)                   // affiche 1 lettre
    c.Clear()
    c.Write(x, y, tui.VisibleText(tw)) // ce qui est visible pour l'instant
    tui.Flush(out, c.String())
    time.Sleep(25 * time.Millisecond)
}
// tui.Skip(tw) = tout d'un coup ; tui.VisibleLen(tw) = nb lettres visibles
```

Version **couleur** : les balises ne doivent pas compter comme lettres,
donc on sépare d'abord, puis on écrit les `n` premières lettres :

```go
plain, colors := tui.SplitTypewriter("test /RED/yo wassup/RED/ gg")
tw := tui.NewTypewriter(plain)
for !tui.IsDone(tw) {
    tui.Tick(tw, 1)
    c.Clear()
    n := tui.VisibleLen(tw)
    tui.WriteColoredRunes(c, x, y, plain, colors, n, "")
    tui.FlushStyled(out, c)
    time.Sleep(25 * time.Millisecond)
}
```

## Boites

```go
tui.DrawBox(c, x, y, w, h)                          // cadre simple
tui.DrawBoxWithTitle(c, x, y, w, h, "MARCHAND")     // cadre + titre
tui.FillRect(c, x, y, w, h, ' ')                    // remplir / effacer une zone
x, y := tui.CenteredBox(c.W, c.H, w, h)             // centrer une boite w*h
```

Toutes les boites sont **creuses** : seul le cadre est dessine,
l'interieur n'est jamais touche (on voit le fond a travers).

## Dialogue + saisie cle en main

```go
quit := tui.Dialogue(c, out, in, "/RED/gg/RED/ test", "suite")
// boite en bas a la bonne taille + typewriter + couleurs.
// ESPACE pendant l'ecriture = skip ; ESPACE/ENTREE apres = texte
// suivant (false a la fin) ; q/ECHAP = quitter (true).
// ZQSD restent des lettres normales : Q = aller a l'ouest, PAS quitter.
```

```go
// Pompe partagee (une seule par clavier !) + boite de saisie :
keys, stop := tui.PumpKeys(in)
defer stop()
tui.LayoutBottomBox(c, "TITRE", texte)        // calcule la boite
tui.DrawTextBox(c, box, colors, n, "indice")  // la redessine
name, ok := tui.AskKeys(c, out, keys, "QUI ES-TU ?", "Ton nom ?", 12, fond)
```

Il n'y a pas d'objet "boite" à détruire : pour la fermer, on efface
(`Clear`) et on redessine sans elle

## Flow du jeu (main -> intro -> jeu)

`cmd/runa/main.go` : mode raw + `audio.StartOST()` (musique en fond,
jeu continue sans carte son) + intro + `menu.Run` :

1. `menu.ShowIntro` : cinématique `.cine` + boite PLAY (skip : q/ESPACE/ENTREE).
2. `menu.Setup` -> `ShowStory` (soul.cine en boucle) : histoire typewriter,
   inputbox pseudo (`AskKeys`), choix race (Humain/Elfe/Nain) + sous-classe,
   **machine a sous du mana** (`runSlotMachine` : roulette qui ralentit et
   tombe sur le mana tiré), puis perso créé avec race + mana.
   Sans video : perso "Traveler" Humain.
3. `menu.Run` -> `StartGame` : monde `map.txt` (165x197, chunks 16x16),
   spawn auto, 30 monstres, boucle `RunGame`.

## Monde,interieurs, PNJ

- `world/` : tiles colorées (`█` par nature : herbe, eau, sable...),
  chunks 16x16 chargés autour du joueur (`EnsureLoaded`), `TileAt`.
- Portes `E` (adjacentes = même porte) + `ZoneAt` : entrer = intérieur
  généré (`world.BuildInterior`), sol bois jaune, murs blancs, porte sud.
- Intérieurs 27x15, 1 PNJ chacun : Marchand (vert), Forgeron (rouge),
  Aieule (magenta), Maitre de guilde (cyan). Nom affiché au-dessus,
  case infranchissable. Sortie = remarcher sur la porte.
- `spawner` : 30 mobs (rats/loups/sangliers/trolls selon rang),
  respawn à proximité du kill. **Gobelin fixe** au terrain
  d'entraînement (pas de spawn aléatoire dessus) pour grind.

## Combat, mort, quetes, boutiques

- Combat (`menu/combat.go` + `internal/combat`) : menus attaque / objet /
  fuite, sorts (coût mana), consommables, effets (poison...), fuite,
  regen mana 5 %. Victoire = XP + respawn du mob + quête guilde.
- Mort : écran `DEAD` rouge sur noir (`menu/death.go`), ESPACE =
  respawn au spawn avec 50 % des PV.
- Guilde : rangs F->S, quêtes "tuer N monstres", `RegisterKill`,
  promotion + récompenses (`TurnInQuest`). HUD : rang + XP + Lvl.
- Marchand (`shop/`) : catalogue + prix, achat/vente, potion de soin
  gratuite la 1re fois. Forgeron (`forge/`) : recettes + fabrication.
  Inventaire (`e` en jeu) : voir/jeter, potions, livres de sort.
- Perso (`character/`) : race (PV : Humain 100 / Elfe 80 / Nain 120,
  mana/speed/force par race), XP/niveaux, sorts connus, arme, or.

## Erreurs

`GameError` remplace `fmt.Errorf` !

```go
return tui.NewError("impossible de lire la taille", err)
// message final : loop.go: impossible de lire la taille: <cause>
tui.ErrorText(e)  // phrase
e.Err             // cause (ou nil)
```

## Cinematiques (.cine)

Convertir une video (demande `ffmpeg`, une fois par video) :

```bash
go run ./cmd/cine video.mp4 [-o film.cine] [-w 80] [-h 22] [-fps 10] [-maxf 300] [-sat 2.0]
```

`-sat` = saturation des couleurs (1 = video d'origine, 2 = deux fois
plus vif). Monte-le (3-4) si ta video est terne et ne sort qu'en
gris/jaune : les lettres (niveaux de gris) ne changent pas, seules
les couleurs sont avivees.

Jouer un `.cine` pour tester (`q` / `ECHAP` / `ENTREE` pour passer) :

```bash
go run ./cmd/cine play film.cine
```

Dans le jeu : `cine.Load(path)` puis `cine.Play(out, in, movie)`.
Format unique `CINE2` couleur : `----COLOR----` + lignes hexa 0-9A-F
= index `tui.FGPalette`. Le convertisseur ne sort que du `CINE2`
(voir `internal/cine/cine.go`).

## Musique

`internal/audio` : `ost.mp3` (mono 44kHz 32k) embarqué dans le binaire
(`assets/embed.go`), décodé en mémoire au lancement, joué en boucle
(`StartOST`, jamais fatal). Deps : `go-mp3` (pur Go) + `malgo`
(C vendu avec, juste `gcc`, pas d'ALSA système requis).
