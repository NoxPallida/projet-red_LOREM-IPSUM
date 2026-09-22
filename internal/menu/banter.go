package menu

import (
	"math/rand"

	"runa/internal/world"
)

// punchlines : une vanne au hasard par type de PNJ, affichee quand on
// lui parle (ESPACE a cote), juste avant d'ouvrir son menu.
// La cle = le genre de batiment, qui designe aussi son habitant
// (Shop -> Marchand, Forge -> Forgeron, Guild -> Maitre).
// Pas d'entree = pas de blague (ex : maison sans habitant).
var punchlines = map[world.ZoneKind][]string{
	world.ZoneShop: {
		"Tout est a vendre ici. Meme toi, pour le bon prix.",
		"Mes potions ? Fraiches du matin. Enfin... du mois dernier.",
		"Reviens quand tu seras riche. Ou mort, je prends les deux.",
	},
	world.ZoneForge: {
		"Mon enclume a vu passer plus de heros que toi.",
		"Tu veux du solide ? Arrete de pleurnicher et paie.",
		"Le fer ne ment jamais. Toi, un peu.",
	},
	world.ZoneGuild: {
		"La guilde observe tes exploits. Enfin... tes tentatives.",
		"Un rang se merite. Un cadavre, c'est plus rapide.",
		"Reviens quand tu auras tue quelque chose d'impressionnant.",
	},
}

// npcBanter tire une punchline du PNJ du batiment.
// false = aucun PNJ bavard ici (maison vide).
func npcBanter(kind world.ZoneKind) (string, bool) {
	lines := punchlines[kind]
	if len(lines) == 0 {
		return "", false
	}
	return lines[rand.Intn(len(lines))], true
}
