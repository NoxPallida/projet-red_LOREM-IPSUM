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
├── .github/
│   ├── pull_request_template.md      # Template de pull request
│   └── workflows/
│       └── ci.yml                    # CI : gofmt + vet + lint + tests
├── assets/
│   ├── ascii.go                      # Chargeur d'art ASCII (embarqué)
│   ├── embed.go                      # Embarque ost.mp3 dans le binaire
│   ├── audio/
│   │   └── ost.mp3                   # Musique (mono 44kHz 32k, ~8 Mo)
│   ├── cinematic/
│   │   ├── soul.cine                 # Cinématique d'intro
│   │   └── videoplayback.cine        # Autre cinématique
│   └── ascii/                        # Art ASCII ennemis / héros (15 fichiers)
│       ├── boar.txt
│       ├── goblin.txt
│       ├── harpy.txt
│       ├── hero.txt
│       ├── hobgoblin.txt
│       ├── kobold.txt
│       ├── minotaur.txt
│       ├── ogre.txt
│       ├── orc.txt
│       ├── rat.txt
│       ├── skeleton_warrior.txt
│       ├── slime.txt
│       ├── troll.txt
│       ├── wolf.txt
│       ├── wyrm.txt
│       └── wyvern.txt
├── src/
│   ├── cmd/
│   │   ├── runa/main.go              # Entrée du jeu : raw + musique + intro + menu + jeu
│   │   └── cine/main.go              # Convertit vidéo -> .cine, mode play pour tester
│   └── internal/                     # (ordre alphabétique)
│       ├── audio/
│       │   └── audio.go              # MP3 en mémoire + decode + lecture en boucle
│       ├── character/
│       │   ├── character.go          # Perso : race/classe, PV/mana, XP/niveaux, sorts, or
│       │   └── inventory_adapter.go  # Adapte inventaire aux menus (shop/forge)
│       ├── cine/
│       │   ├── cine.go               # Format .cine : Save + Load
│       │   ├── play.go               # Play : joue un film (q pour passer)
│       │   └── play_test.go          # Tests lecteur cinématique
│       ├── combat/
│       │   └── combat.go             # Règles combat : attaques, objets, fuite, effets, mana
│       ├── enemies/
│       │   └── enemies.go            # Templates (rat, loup, sanglier, troll, gobelin...)
│       ├── forge/
│       │   └── forge.go              # Forgeron : recettes, coûts, fabrication
│       ├── guild/
│       │   └── quest_system.go       # Rangs F->S, quêtes (tuer X), promotion, récompenses
│       ├── inventory/
│       │   └── inventory.go          # Sac : capacité, ajout, upgrades
│       ├── item/
│       │   ├── item.go               # Base commune des objets
│       │   ├── consumable.go         # Potions, livres de sort
│       │   ├── equipment.go          # Équipements + bonus PV
│       │   ├── loot.go               # Matériaux (fourrure, peau, cuir, plume)
│       │   ├── weapon.go             # Armes (dégâts, mana)
│       │   └── item-effect.go        # Effets génériques (poison, soin...)
│       ├── menu/                     # Écrans du jeu
│       │   ├── menu.go               # Setup (pseudo) + Run (perso -> StartGame)
│       │   ├── intro.go              # ShowIntro, histoire, choix race/classe, machine à sous
│       │   ├── game.go               # Boucle jeu : ZQSD, E inventaire, ESPACE interagir
│       │   ├── world.go              # Fond carte + ennemis + HUD + rencontre
│       │   ├── combat.go             # Écran combat (menus attaque/objet/fuite)
│       │   ├── death.go              # Écran DEAD + respawn 50 % PV
│       │   ├── interior.go           # Dedans/dehors bâtiments, PNJ, porte sortie
│       │   ├── shop.go               # Menu marchand (catalogue, acheter/vendre)
│       │   ├── forge.go              # Menu forgeron (recettes, fabriquer)
│       │   ├── guild.go              # Menu guilde (quêtes, rangs, récompenses)
│       │   ├── inventory.go          # Menu inventaire (voir, jeter, potions, livres)
│       │   ├── character.go          # Fiche perso (stats, sorts connus, arme)
│       │   ├── banter.go             # Banter PNJ / dialogues d'ambiance
│       │   └── chest.go              # Coffres / loot au sol
│       ├── shop/
│       │   └── shop.go               # Marchand : catalogue, prix, achat/vente
│       ├── spawner/
│       │   └── spawner.go            # Peuple la carte (30 mobs, respawn proche, gobelin fixe)
│       ├── spell/
│       │   └── spell.go              # Races, sous-classes, sorts (coûts, dégâts, effets)
│       ├── tui/                      # Moteur graphique (prêt)
│       │   ├── ansi.go               # Alt-screen, curseur, Flush / FlushStyled
│       │   ├── box.go                # Cadres, titres, remplissage, centrage
│       │   ├── canvas.go             # Mémoire vidéo (lettre + couleur), rendu brut / style
│       │   ├── dialogue.go           # Dialogue() : boîte en bas + typewriter + couleurs
│       │   ├── prompt.go             # PumpKeys, IsQuit/IsConfirm, TextBox, Ask/AskKeys
│       │   ├── color.go              # Palette 16 couleurs, dégradés, balises /RED/, boîte en dégradé
│       │   ├── errors.go             # GameError : fichier auto + opération + cause
│       │   ├── input.go              # ReadKey bloquant : flèches, ENTREE, ECHAP, UTF-8
│       │   ├── keyboard.go           # PollKey non-bloquant + protocole kitty (tenues)
│       │   ├── keyavail_unix.go      # Détecte touche dispo sans bloquer (Linux)
│       │   ├── keyavail_windows.go   # Détecte touche dispo sans bloquer (Windows)
│       │   ├── layer.go              # Calques superposés (le dernier gagne)
│       │   ├── loop.go               # Boucle affichage -> touche -> affichage
│       │   ├── terminal.go           # Taille écran dynamique, détection TTY, mode brut
│       │   └── typewriter.go         # Effet machine à écrire (fonctions, pas méthodes)
│       └── world/
│           ├── logic.go              # Tiles, chunks 16x16, zones, TileAt
│           ├── map.go                # Parse map.txt, légende symboles -> tiles
│           ├── player.go             # Joueur : position, direction, déplacement
│           ├── interior.go           # Pièces générées (murs, sol, porte, PNJ)
│           ├── interior_test.go      # Tests génération intérieure
│           └── maps/
│               └── map.txt           # La grande carte (165x197)
├── docs/
│   ├── test.md                       # Notes de test
│   └── moteur.md                     # Doc moteur + gameplay (clavier, combat, quêtes...)
├── go.mod                            # Module runa, Go 1.27
├── go.sum
├── .golangci.yml                     # Config golangci-lint
├── .gitignore
├── LICENSE
└── README.md                         # Ce fichier
```
