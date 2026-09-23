package forge

import "runa/src/internal/item"

// Crafter est l'interface minimale dont la forge a besoin, côté joueur.
// Volontairement distincte de shop.Customer : la forge a besoin de compter
// des quantités de matériaux, pas juste vérifier la présence d'un item.
type Crafter interface {
	Money() uint16
	SpendMoney(amount uint16) bool

	// CountItem renvoie combien d'exemplaires d'un item le joueur possède,
	// nécessaire pour vérifier des recettes multi-matériaux (ex: 2 fourrures).
	CountItem(i item.Item) uint8
	// RemoveItemQty retire une quantité précise d'un item ; renvoie false
	// si le joueur n'en a pas assez (ne doit jamais être appelé après un
	// CountItem qui a déjà validé la quantité, mais reste défensif).
	RemoveItemQty(i item.Item, qty uint8) bool

	CanFitItem(i item.Item) bool
	AddItem(i item.Item)
}

// Ingredient décrit un matériau requis et sa quantité pour une recette.
type Ingredient struct {
	Item     item.Item
	Quantity uint8
}

// Recipe décrit ce que le forgeron peut fabriquer.
type Recipe struct {
	Result      item.Item
	Ingredients []Ingredient
	Price       uint16
}

type CraftResult int

const (
	Success CraftResult = iota
	ErrRecipeNotFound
	ErrMissingMaterials
	ErrInsufficientMoney
	ErrInventoryFull
)

// Craft exécute la fabrication, dans l'ordre imposé par le README (§18) :
// vérifier matériaux ET argent avant de retirer quoi que ce soit.
func Craft(c Crafter, result item.Item) (CraftResult, string) {
	recipe, ok := findRecipe(result)
	if !ok {
		return ErrRecipeNotFound, ""
	}

	// 1. vérifier les matériaux (sans rien retirer encore)
	for _, ing := range recipe.Ingredients {
		if c.CountItem(ing.Item) < ing.Quantity {
			return ErrMissingMaterials, ""
		}
	}

	// 2. vérifier l'argent
	if c.Money() < recipe.Price {
		return ErrInsufficientMoney, ""
	}

	// 3. vérifier la capacité de l'inventaire pour le résultat
	// (même raisonnement que shop.Buy : jamais retirer avant de savoir
	// qu'on peut réellement recevoir l'objet)
	if !c.CanFitItem(recipe.Result) {
		return ErrInventoryFull, ""
	}

	// 4. retirer matériaux + argent, puis ajouter le résultat
	for _, ing := range recipe.Ingredients {
		c.RemoveItemQty(ing.Item, ing.Quantity) // déjà validé à l'étape 1
	}
	c.SpendMoney(recipe.Price) // déjà validé à l'étape 2

	c.AddItem(recipe.Result)

	return Success, recipe.Result.Name()
}

func findRecipe(result item.Item) (Recipe, bool) {
	for _, r := range Recipes {
		if r.Result.Name() == result.Name() {
			return r, true
		}
	}
	return Recipe{}, false
}

var Recipes = []Recipe{
	// Axes
	{
		Result: item.GoblinCleaver,
		Ingredients: []Ingredient{
			{Item: item.GoblinSkin, Quantity: 4},
			{Item: item.RatHide, Quantity: 2},
		},
		Price: 4,
	},
	{
		Result: item.BoarSplitter,
		Ingredients: []Ingredient{
			{Item: item.BoarLeather, Quantity: 4},
			{Item: item.OrcHide, Quantity: 2},
		},
		Price: 8,
	},
	{
		Result: item.WarAxe,
		Ingredients: []Ingredient{
			{Item: item.BoarTusk, Quantity: 2},
			{Item: item.OrcTusk, Quantity: 2},
		},
		Price: 15,
	},
	{
		Result: item.OrcGreataxe,
		Ingredients: []Ingredient{
			{Item: item.OrcTusk, Quantity: 4},
			{Item: item.TrollHide, Quantity: 2},
		},
		Price: 25,
	},
	{
		Result: item.MinotaurExecutioner,
		Ingredients: []Ingredient{
			{Item: item.MinotaurHorn, Quantity: 2},
			{Item: item.OgreClub, Quantity: 1},
		},
		Price: 60,
	},
	{
		Result: item.WyvernDecapitator,
		Ingredients: []Ingredient{
			{Item: item.WyvernFang, Quantity: 2},
			{Item: item.WyvernScale, Quantity: 3},
		},
		Price: 200,
	},
	// Staffs
	{
		Result: item.SlimeScepter,
		Ingredients: []Ingredient{
			{Item: item.SlimeMucus, Quantity: 5},
			{Item: item.GoblinSkin, Quantity: 2},
		},
		Price: 4,
	},
	{
		Result: item.ApprenticeStaff,
		Ingredients: []Ingredient{
			{Item: item.KoboldFang, Quantity: 4},
			{Item: item.SlimeMucus, Quantity: 3},
		},
		Price: 12,
	},
	{
		Result: item.HarpyWand,
		Ingredients: []Ingredient{
			{Item: item.HarpyFeather, Quantity: 3},
			{Item: item.WolfClaw, Quantity: 2},
		},
		Price: 20,
	},
	{
		Result: item.BoneStaff,
		Ingredients: []Ingredient{
			{Item: item.BoneShard, Quantity: 5},
			{Item: item.TrollHide, Quantity: 2},
		},
		Price: 35,
	},
	{
		Result: item.WyrmRod,
		Ingredients: []Ingredient{
			{Item: item.WyrmScale, Quantity: 3},
			{Item: item.OgreClub, Quantity: 1},
		},
		Price: 90,
	},
	{
		Result: item.ArchmageWyvernStaff,
		Ingredients: []Ingredient{
			{Item: item.WyvernScale, Quantity: 4},
			{Item: item.WyvernFang, Quantity: 1},
		},
		Price: 250,
	},
	// Swords
	{
		Result: item.KoboldDagger,
		Ingredients: []Ingredient{
			{Item: item.KoboldFang, Quantity: 3},
			{Item: item.RatHide, Quantity: 2},
		},
		Price: 3,
	},
	{
		Result: item.WolfToothBlade,
		Ingredients: []Ingredient{
			{Item: item.WolfClaw, Quantity: 4},
			{Item: item.WolfFur, Quantity: 3},
		},
		Price: 10,
	},
	{
		Result: item.Spellblade,
		Ingredients: []Ingredient{
			{Item: item.HarpyFeather, Quantity: 2},
			{Item: item.BoneShard, Quantity: 3},
		},
		Price: 40,
	},
	{
		Result: item.TrollSlayer,
		Ingredients: []Ingredient{
			{Item: item.TrollHide, Quantity: 3},
			{Item: item.BoarTusk, Quantity: 2},
		},
		Price: 50,
	},
	{
		Result: item.WyrmSaber,
		Ingredients: []Ingredient{
			{Item: item.WyrmScale, Quantity: 3},
			{Item: item.MinotaurHorn, Quantity: 1},
		},
		Price: 100,
	},
	{
		Result: item.WyvernFangBlade,
		Ingredients: []Ingredient{
			{Item: item.WyvernFang, Quantity: 3},
			{Item: item.WyvernScale, Quantity: 2},
		},
		Price: 220,
	},
	// Bows
	{
		Result: item.RatRunnerBow,
		Ingredients: []Ingredient{
			{Item: item.RatHide, Quantity: 4},
			{Item: item.WolfFur, Quantity: 2},
		},
		Price: 3,
	},
	{
		Result: item.BoarRecurveBow,
		Ingredients: []Ingredient{
			{Item: item.BoarLeather, Quantity: 4},
			{Item: item.WolfClaw, Quantity: 2},
		},
		Price: 9,
	},
	{
		Result: item.HarpyFeatherBow,
		Ingredients: []Ingredient{
			{Item: item.HarpyFeather, Quantity: 4},
			{Item: item.OrcHide, Quantity: 2},
		},
		Price: 20,
	},
	{
		Result: item.BoneLongbow,
		Ingredients: []Ingredient{
			{Item: item.BoneShard, Quantity: 4},
			{Item: item.BoarTusk, Quantity: 2},
		},
		Price: 45,
	},
	{
		Result: item.MinotaurGreatbow,
		Ingredients: []Ingredient{
			{Item: item.MinotaurHorn, Quantity: 2},
			{Item: item.OgreClub, Quantity: 1},
		},
		Price: 110,
	},
	{
		Result: item.WyvernStalkerBow,
		Ingredients: []Ingredient{
			{Item: item.WyvernScale, Quantity: 3},
			{Item: item.WyvernFang, Quantity: 1},
		},
		Price: 230,
	},

	// Armures
	{
		Result: item.LeatherCap,
		Ingredients: []Ingredient{
			{Item: item.WolfFur, Quantity: 3},
			{Item: item.RatHide, Quantity: 2},
		},
		Price: 3,
	},
	{
		Result: item.LeatherTunic,
		Ingredients: []Ingredient{
			{Item: item.BoarLeather, Quantity: 4},
			{Item: item.WolfFur, Quantity: 3},
		},
		Price: 8,
	},
	{
		Result: item.LeatherPants,
		Ingredients: []Ingredient{
			{Item: item.BoarLeather, Quantity: 3},
			{Item: item.WolfFur, Quantity: 2},
		},
		Price: 6,
	},
	{
		Result: item.LeatherBoots,
		Ingredients: []Ingredient{
			{Item: item.WolfFur, Quantity: 2},
			{Item: item.WolfClaw, Quantity: 2},
		},
		Price: 4,
	},

	{
		Result: item.OrcHelmet,
		Ingredients: []Ingredient{
			{Item: item.OrcHide, Quantity: 3},
			{Item: item.BoarTusk, Quantity: 1},
		},
		Price: 10,
	},
	{
		Result: item.OrcChestplate,
		Ingredients: []Ingredient{
			{Item: item.OrcHide, Quantity: 5},
			{Item: item.TrollHide, Quantity: 2},
		},
		Price: 20,
	},
	{
		Result: item.OrcGreaves,
		Ingredients: []Ingredient{
			{Item: item.OrcHide, Quantity: 4},
			{Item: item.TrollHide, Quantity: 1},
		},
		Price: 15,
	},
	{
		Result: item.OrcBoots,
		Ingredients: []Ingredient{
			{Item: item.OrcHide, Quantity: 2},
			{Item: item.OrcTusk, Quantity: 1},
		},
		Price: 10,
	},

	{
		Result: item.HarpyCowl,
		Ingredients: []Ingredient{
			{Item: item.HarpyFeather, Quantity: 3},
			{Item: item.WolfFur, Quantity: 2},
		},
		Price: 12,
	},
	{
		Result: item.HarpyRobe,
		Ingredients: []Ingredient{
			{Item: item.HarpyFeather, Quantity: 5},
			{Item: item.BoneShard, Quantity: 2},
		},
		Price: 25,
	},
	{
		Result: item.HarpyBreeches,
		Ingredients: []Ingredient{
			{Item: item.HarpyFeather, Quantity: 4},
			{Item: item.BoarLeather, Quantity: 2},
		},
		Price: 18,
	},
	{
		Result: item.HarpySandals,
		Ingredients: []Ingredient{
			{Item: item.HarpyFeather, Quantity: 2},
			{Item: item.SlimeMucus, Quantity: 3},
		},
		Price: 12,
	},

	{
		Result: item.MinotaurHelm,
		Ingredients: []Ingredient{
			{Item: item.MinotaurHorn, Quantity: 1},
			{Item: item.TrollHide, Quantity: 3},
		},
		Price: 30,
	},
	{
		Result: item.MinotaurCuirass,
		Ingredients: []Ingredient{
			{Item: item.MinotaurHorn, Quantity: 2},
			{Item: item.OgreClub, Quantity: 1},
		},
		Price: 60,
	},
	{
		Result: item.MinotaurLeggings,
		Ingredients: []Ingredient{
			{Item: item.MinotaurHorn, Quantity: 1},
			{Item: item.OrcHide, Quantity: 4},
		},
		Price: 45,
	},
	{
		Result: item.MinotaurStompers,
		Ingredients: []Ingredient{
			{Item: item.MinotaurHorn, Quantity: 1},
			{Item: item.BoarTusk, Quantity: 2},
		},
		Price: 30,
	},

	{
		Result: item.WyrmScaleCrown,
		Ingredients: []Ingredient{
			{Item: item.WyrmScale, Quantity: 2},
			{Item: item.HarpyFeather, Quantity: 3},
		},
		Price: 35,
	},
	{
		Result: item.WyrmScaleHauberk,
		Ingredients: []Ingredient{
			{Item: item.WyrmScale, Quantity: 4},
			{Item: item.BoneShard, Quantity: 4},
		},
		Price: 70,
	},
	{
		Result: item.WyrmScaleGreaves,
		Ingredients: []Ingredient{
			{Item: item.WyrmScale, Quantity: 3},
			{Item: item.TrollHide, Quantity: 2},
		},
		Price: 50,
	},
	{
		Result: item.WyrmScaleBoots,
		Ingredients: []Ingredient{
			{Item: item.WyrmScale, Quantity: 2},
			{Item: item.MinotaurHorn, Quantity: 1},
		},
		Price: 35,
	},

	{
		Result: item.WyvernHelm,
		Ingredients: []Ingredient{
			{Item: item.WyvernScale, Quantity: 2},
			{Item: item.WyvernFang, Quantity: 1},
		},
		Price: 100,
	},
	{
		Result: item.WyvernBreastplate,
		Ingredients: []Ingredient{
			{Item: item.WyvernScale, Quantity: 4},
			{Item: item.WyvernFang, Quantity: 2},
		},
		Price: 200,
	},
	{
		Result: item.WyvernGreaves,
		Ingredients: []Ingredient{
			{Item: item.WyvernScale, Quantity: 3},
			{Item: item.WyvernFang, Quantity: 1},
		},
		Price: 150,
	},
	{
		Result: item.WyvernSabatons,
		Ingredients: []Ingredient{
			{Item: item.WyvernScale, Quantity: 2},
			{Item: item.WyrmScale, Quantity: 2},
		},
		Price: 100,
	},
}
