# Moteur TUI

Le moteur est dans `internal/tui/`. Principe : on dessine tout sur un
`Canvas` (mémoire vidéo), puis on l'envoie d'un coup à l'écran ( c le meilleur compris que j'ai trouver )
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
- `ReadKey(in)` rend une touche (`KeyUp`, `KeyEnter`, `KeyRune`...)

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

// vert puiss normal, fond bleu :
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

## Typewriter

```go
tw := tui.NewTypewriter("Bienvenue dans RED...")
for !tw.Done() {
    tw.Tick(1)                        // affiche 1 lettre par 1 
    c.Clear()
    c.Write(x, y, tw.Visible())       // ce qui est visible pour l'instant
    tui.Flush(out, c.String())
    time.Sleep(25 * time.Millisecond) // le seul Sleep, (hors du moteur nous on peut se tick affichage )
}
```

Version **couleur** : les balises ne doivent pas compter comme lettres,
donc on sépare d'abord, puis on écrit les `n` premières lettres :

```go
plain, colors := tui.SplitTypewriter("test /RED/yo wassup/RED/ gg")
tw := tui.NewTypewriter(plain)
for !tw.Done() {
    tw.Tick(1)
    c.Clear()
    n := len([]rune(tw.Visible()))
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

## Dialogue cle en main

```go
quit := tui.Dialogue(c, out, in, "/RED/gg/RED/ test")
// boite en bas a la bonne taille + typewriter + couleurs,
// attend ENTREE (false) ; q/ECHAP rend true
```

Il n'y a pas d'objet "boite" à détruire : pour la fermer, on efface
(`Clear`) et on redessine sans elle

## Menu (boite de dialogue en bas)

`menu.Run(in, out)` fait exactement : boite en bas + typewriter dedans +
`[ENTREE]` pour fermer. Pour un nouveau menu, ( on modif plus tard ) :

```go
w, h, err := tui.Size()
if err != nil {	
    return err // pas de terminal lisible
}
c := tui.NewCanvas(w, h)
tui.EnterAltScreen(out)
defer tui.ExitAltScreen(out)
// ... boucle : Clear -> dessiner -> Flush -> ReadKey ...
```

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
