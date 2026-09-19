package forge

import "runa/internal/item"

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

var Recipes = []Recipe{
	{
		Result: item.AdventurerHat,
		Ingredients: []Ingredient{
			{item.RavenFeather, 1},
			{item.BoarLeather, 1},
		},
		Price: 5,
	},
	{
		Result: item.AdventurerTunic,
		Ingredients: []Ingredient{
			{item.WolfFur, 2},
			{item.TrollHide, 1},
		},
		Price: 5,
	},
	{
		Result: item.AdventurerBoots,
		Ingredients: []Ingredient{
			{item.WolfFur, 1},
			{item.BoarLeather, 1},
		},
		Price: 5,
	},
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
