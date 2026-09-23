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
		"Everything is for sale here. Even you, for the right price.",
		"My potions? Fresh this morning. Well... last month.",
		"Come back when you're rich. Or dead, I take both.",
	},
	world.ZoneForge: {
		"My anvil has seen more heroes than you.",
		"Want something sturdy? Stop whining and pay.",
		"Iron never lies. You, a little.",
	},
	world.ZoneGuild: {
		"The guild watches your exploits. Well... your attempts.",
		"A rank must be earned. A corpse is faster.",
		"Come back when you've killed something impressive.",
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
