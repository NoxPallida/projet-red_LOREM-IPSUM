# projet-red_runa
Projet red_runa Ynov 2026/2027
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
go vet ./...
gofmt -l internal/ cmd/
go test ./...
```

## Fichiers (une ligne chacun)

```text
projet-red_runa/
├── .github/workflows/ci.yml      # CI : gofmt + vet + lint + tests
├── cmd/runa/main.go              # entree du jeu : raw + musique + intro + menu + jeu
├── cmd/cine/main.go              # convertit video -> .cine, mode play pour tester
├── assets/embed.go               # embarque ost.mp3 dans le binaire
├── assets/audio/ost.mp3          # musique (mono 44kHz 32k, ~8 Mo)
├── assets/cinematic/             # soul.cine (intro) + videoplayback.cine
├── internal/
│   ├── audio/audio.go            # mp3 memoire + decode + lecture boucle
│   ├── tui/                      # moteur graphique (pret)
│   │   ├── ansi.go               # alt-screen, curseur, Flush / FlushStyled
│   │   ├── box.go                # cadres, titres, remplissage, centrage
│   │   ├── canvas.go             # memoire video (lettre + couleur), rendu brut / style
│   │   ├── dialogue.go           # Dialogue() : boite en bas + typewriter + couleurs
│   │   ├── prompt.go             # PumpKeys, IsQuit/IsConfirm, TextBox, Ask/AskKeys
│   │   ├── color.go              # palette 16 couleurs, degrades, balises /RED/, boite en degrade
│   │   ├── errors.go             # GameError : fichier auto + operation + cause
│   │   ├── input.go              # ReadKey bloquant : fleches, ENTREE, ECHAP, UTF-8
│   │   ├── keyboard.go           # PollKey non-bloquant + protocole kitty (tenues)
│   │   ├── keyavail_unix.go      # detecte touche dispo sans bloquer (linux)
│   │   ├── keyavail_windows.go   # detecte touche dispo sans bloquer (windows)
│   │   ├── layer.go              # calques superposes (le dernier gagne)
│   │   ├── loop.go               # boucle affichage -> touche -> affichage
│   │   ├── terminal.go           # taille ecran dynamique, detection TTY, mode brut
│   │   └── typewriter.go         # effet machine a ecrire (fonctions, pas methodes)
│   ├── menu/                     # ecrans du jeu
│   │   ├── menu.go               # Setup (pseudo) + Run (perso -> StartGame)
│   │   ├── intro.go              # ShowIntro, histoire, choix race/classe, machine a sous
│   │   ├── game.go               # boucle jeu : ZQSD, E inventaire, ESPACE interagir
│   │   ├── world.go              # fond carte + ennemis + HUD + rencontre
│   │   ├── combat.go             # ecran combat (menus attaque/objet/fuite)
│   │   ├── death.go              # ecran DEAD + respawn 50 % PV
│   │   ├── interior.go           # dedans/dehors batiments, PNJ, porte sortie
│   │   ├── shop.go               # menu marchand (catalogue, acheter/vendre)
│   │   ├── forge.go              # menu forgeron (recettes, fabriquer)
│   │   ├── guild.go              # menu guilde (quetes, rangs, recompenses)
│   │   ├── inventory.go          # menu inventaire (voir, jeter, potions, livres)
│   │   ├── character.go          # fiche perso (stats, sorts connus, arme)
│   │   └── errordemo.go          # showTestError : msgbox NewError X secondes (demo)
│   ├── character/
│   │   ├── character.go          # perso : race/classe, PV/mana, XP/niveaux, sorts, or
│   │   └── inventory_adapter.go  # adapte inventaire aux menus (shop/forge)
│   ├── combat/
│   │   └── combat.go             # regles combat : attaques, objets, fuite, effets, mana
│   ├── enemies/
│   │   └── enemies.go            # templates (rat, loup, sanglier, troll, gobelin...)
│   ├── guild/
│   │   └── quest_system.go       # rangs F->S, quetes (tuer X), promotion, recompenses
│   ├── spawner/
│   │   └── spawner.go            # peuplera la carte (30 mobs, respawn proche, gobelin fixe)
│   ├── inventory/
│   │   └── inventory.go          # sac : capacite, ajout, upgrades
│   ├── item/
│   │   ├── item.go               # base commune des objets
│   │   ├── consumable.go         # potions, livres de sort
│   │   ├── equipment.go          # equipements + bonus PV
│   │   ├── loot.go               # materiaux (fourrure, peau, cuir, plume)
│   │   ├── weapon.go             # armes (degats, mana)
│   │   └── item-effect.go        # effets generiques (poison, soin...)
│   ├── shop/shop.go              # marchand : catalogue, prix, achat/vente
│   ├── forge/forge.go            # forgeron : recettes, couts, fabrication
│   ├── spell/spell.go            # races, sous-classes, sorts (couts, degats, effets)
│   ├── world/
│   │   ├── logic.go              # tiles, chunks 16x16, zones, TileAt
│   │   ├── map.go                # parse map.txt, legende symboles -> tiles
│   │   ├── player.go             # joueur : position, direction, deplacement
│   │   ├── interior.go           # pieces generees (murs, sol, porte, PNJ)
│   │   └── maps/map.txt          # la grande carte (165x197)
│   └── cine/
│       ├── cine.go               # format .cine : Save + Load
│       └── play.go               # Play : joue un film (q pour passer)
├── docs/test.md                  # notes de test
├── docs/moteur.md                # doc moteur + gameplay (clavier, combat, quetes...)
├── go.mod                        # module runa, Go 1.27
└── README.md                     # ce fichier
gg
```
