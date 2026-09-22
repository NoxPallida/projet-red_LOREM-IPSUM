# projet-red_LOREM-IPSUM
Projet Red Ynov 2026/2027
ASCII_ART = """

                                    ,--.
                                   /{    }
                                  K,   /j}
                                 /  ~Y`
                           ,   /   /
                       {_'-K.__ /
                        `/-.__L._
                            /  ' /`\\_}
                      /  ' /
             ____   /  ' /
      ,-'~~~~    ~~/  ' /_
    ,'             ``~~~  ',
   (                        Y
  {                         I
 {      -                    `,
 |       ',                   )
 |        |   ,..__      __. Y
 |    .,_./  Y ' // ^Y   J   )|
 \\           |' //   |   |   ||
  \\          L_//    . _ (_,.'(
   \\,   ,      ^^""' / |      )
     \\_  \\          /,L]     /
       '-_~-,       ` `   ./`
          `'{_            )
              ^^\\..___,.--`
"""

## Lancer / verifier

```bash
go run ./cmd/runa
go vet ./internal/tui/ ./internal/menu/
gofmt -l internal/ cmd/
```

## Fichiers (une ligne chacun)

```text
projet-red_LOREM-IPSUM/
├── .github/workflows/ci.yml      # CI : gofmt + vet + lint (sans tests ni build)
├── cmd/runa/main.go              # entree du jeu, appelle menu.Run
├── cmd/cine/main.go              # convertit video -> .cine, mode play pour tester
├── assets/embed.go               # embarque ost.mp3 dans le binaire
├── assets/audio/ost.mp3          # musique (mono 22kHz 24k, ~6 Mo)
├── internal/
│   ├── audio/audio.go            # mp3 memoire + decode + lecture boucle
│   ├── tui/                      # moteur graphique (pret)
│   │   ├── ansi.go               # alt-screen, curseur, Flush / FlushStyled
│   │   ├── box.go                # cadres, titres, remplissage, centrage
│   │   ├── canvas.go             # memoire video (lettre + couleur), rendu brut / style
│   │   ├── dialogue.go           # Dialogue() : boite en bas + typewriter + couleurs
│   │   ├── color.go              # palette 16 couleurs, degrades, balises /RED/, boite en degrade
│   │   ├── errors.go             # GameError : fichier auto + operation + cause
│   │   ├── input.go              # clavier : fleches, ENTREE, ECHAP, q, UTF-8
│   │   ├── layer.go              # calques superposes (le dernier gagne)
│   │   ├── loop.go               # boucle affichage -> touche -> affichage
│   │   ├── terminal.go           # taille ecran dynamique, detection TTY, mode brut
│   │   └── typewriter.go         # effet machine a ecrire, compatible texte colore
│   ├── menu/
│   │   ├── menu.go               # Run : demo erreur + boite dialogue + texte qui s'ecrit
│   │   ├── intro.go              # ShowIntro : cinematique + boite PLAY par-dessus
│   │   ├── errordemo.go          # showTestError : msgbox NewError X secondes (demo)
│   │   ├── character.go          # sous-menu personnage (stub)
│   │   ├── inventory.go          # sous-menu inventaire (stub)
│   │   ├── shop.go               # sous-menu marchand (stub)
│   │   └── forge.go              # sous-menu forgeron (stub)
│   ├── character/
│   │   └── character.go          # personnage, classes, PV, InitCharacter (en cours)
│   ├── item/
│   │   ├── item.go               # base commune des objets
│   │   ├── consumable.go         # potions, livres de sort
│   │   ├── equipment.go          # equipements + bonus PV
│   │   ├── loot.go               # materiaux (fourrure, peau, cuir, plume)
│   │   ├── weapon.go             # armes (bonus, hors cahier)
│   │   ├── item-effect.go        # effets generiques (poison, soin...)
│   ├── inventory/
│   │   └── inventory.go          # capacite, ajout, upgrades (a coder)
│   ├── cine/
│   │   ├── cine.go               # format .cine : Save + Load
│   │   └── play.go               # Play : joue un film (q pour passer)
│   ├── shop/shop.go              # marchand, catalogue, prix (a coder)
│   ├── forge/forge.go            # recettes forgeron (a coder)
│   └── spell/spell.go            # sorts, apprentissage (a coder)
├── docs/test.md                  # notes de test
├── docs/moteur.md                # doc moteur : boucle, couleurs, degrades, typewriter, boites
├── go.mod                        # module runa, Go 1.27
└── README.md                     # ce fichier
```
