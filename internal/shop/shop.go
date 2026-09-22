package shop

import "runa/internal/item"

// Customer est l'interface minimale dont le shop a besoin, côté joueur.
// character.Character l'implémentera — on évite d'importer character ici
// (character → shop, jamais l'inverse, cf. architecture du README).
type Customer interface {
	Money() uint16
	SpendMoney(amount uint16) bool
	EarnMoney(amount uint16)

	// CanFitItem vérifie la capacité SANS ajouter l'item (étape 3 du README,
	// distincte de l'ajout réel à l'étape 6 — sinon on risque de retirer
	// l'argent avant de savoir si l'inventaire a de la place).
	CanFitItem(i item.Item) bool
	AddItem(i item.Item)
	RemoveItem(i item.Item) bool

	HasClaimedFreePotion() bool
	ClaimFreePotion()

	CanUpgradeInventory() bool
	UpgradeInventory()
}

// CatalogEntry décrit un item vendu par le marchand, avec son propre prix
// d'achat. Volontairement indépendant de item.PriceBuy() : certains items
// (les matériaux/loot) n'implémentent pas PriceBuy puisqu'ils ne sont pas
// achetables "par eux-mêmes" (ex: butin de forgeron), mais le SHOP, lui,
// peut très bien décider de les vendre à un prix qu'il fixe.
type CatalogEntry struct {
	Item  item.Item
	Price uint16
}

var Catalog = []CatalogEntry{
	{item.SmallHealPotion, 2},
	{item.HealPotion, 10},
	{item.LargeHealPotion, 20},
	{item.TitanicHealPotion, 30},
	{item.PoisonPotion, 15},
}

// InventoryUpgradePrice : prix de l'augmentation d'inventaire.
// Le README le liste explicitement comme "à définir avec le groupe" —
// ne pas l'inventer silencieusement, cette valeur est un placeholder.
const InventoryUpgradePrice uint16 = 0

type PurchaseResult int

const (
	Success PurchaseResult = iota
	ErrItemNotSold
	ErrInsufficientMoney
	ErrInventoryFull
	ErrUpgradeMaxed
)

// Buy exécute l'achat d'un item du catalogue, dans l'ordre imposé
// par le cahier des charges (voir README §16).
func Buy(c Customer, target item.Item) (PurchaseResult, string) {
	entry, ok := findCatalogEntry(target)
	if !ok {
		return ErrItemNotSold, ""
	}

	// 1. déterminer le prix (avec la règle "première potion gratuite")
	price := entry.Price
	isHealPotion := entry.Item.Name() == item.HealPotion.Name()
	if isHealPotion && !c.HasClaimedFreePotion() {
		price = 0
	}

	// 2. vérifier l'argent
	if c.Money() < price {
		return ErrInsufficientMoney, "Too poor :("
	}

	// 3. vérifier la capacité de l'inventaire (sans ajouter encore)
	if !c.CanFitItem(entry.Item) {
		return ErrInventoryFull, "Not enough space"
	}

	// 4-6. effectuer l'achat : retirer l'argent, puis ajouter l'item
	if price > 0 {
		c.SpendMoney(price) // déjà validé à l'étape 2, ne peut pas échouer ici
	}
	c.AddItem(entry.Item)

	if isHealPotion {
		c.ClaimFreePotion()
	}

	// 7. afficher le nom (fait par l'appelant/menu, on le renvoie juste)
	return Success, entry.Item.Name()
}

// BuyInventoryUpgrade suit la même logique de vérification, mais ne
// manipule pas item.Item : c'est une capacité, pas un objet.
func BuyInventoryUpgrade(c Customer) PurchaseResult {
	if !c.CanUpgradeInventory() {
		return ErrUpgradeMaxed
	}
	if c.Money() < InventoryUpgradePrice {
		return ErrInsufficientMoney
	}
	c.SpendMoney(InventoryUpgradePrice)
	c.UpgradeInventory()
	return Success
}

// Sell permet au joueur de revendre un item qu'il possède, au prix
// que l'item lui-même déclare (PriceSell, fourni par item.Base,
// donc disponible sur absolument tous les items).
func Sell(c Customer, target item.Item) (PurchaseResult, string) {
	if !c.RemoveItem(target) {
		return ErrItemNotSold, ""
	}
	c.EarnMoney(uint16(target.PriceSell()))
	return Success, target.Name()
}

func findCatalogEntry(target item.Item) (CatalogEntry, bool) {
	for _, e := range Catalog {
		if e.Item.Name() == target.Name() {
			return e, true
		}
	}
	return CatalogEntry{}, false
}
