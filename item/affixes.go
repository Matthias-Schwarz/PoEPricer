package item

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Useability struct {
	values [19]string
}

type Mod struct {
	Text     string
	MinValue float64
	MaxValue float64
}

type Affix struct {
	Mods       ModList
	Name       string
	Group      int
	Level      int
	IsPrefix   bool
	Useability *Useability
}

type AffixList []*Affix
type ModList []Mod
type ModMap map[string]Mod

//var printMe = true

var descToText map[string]string
var affixList []*Affix

func (m ModMap) String() string {
	result := "["
	add := ""
	for _, v := range m {
		result += add + v.Text
		add = " "
	}
	return result + "]"
}

func (list AffixList) String() string {
	prefixes := make([]string, 0, 3)
	suffixes := make([]string, 0, 3)
	for i := 0; i < len(list); i++ {
		if list[i].IsPrefix {
			prefixes = append(prefixes, list[i].Name)
		} else {
			suffixes = append(suffixes, list[i].Name)
		}
	}
	sort.Strings(prefixes)
	sort.Strings(suffixes)
	result := ""
	for i := 0; i < len(prefixes); i++ {
		if i > 0 {
			result += ", "
		}
		result += prefixes[i]
	}
	if len(prefixes) > 0 && len(suffixes) > 0 {
		result += ", "
	}
	for i := 0; i < len(suffixes); i++ {
		if i > 0 {
			result += ", "
		}
		result += suffixes[i]
	}
	return result
}

func (list AffixList) magicName(baseName string) string {
	pre := ""
	numPres := 0
	suff := ""
	numSuffs := 0
	for i := 0; i < len(list); i++ {
		affix := list[i]
		if affix.IsPrefix {
			numPres++
			pre += affix.Name
		} else {
			numSuffs++
			suff += affix.Name
		}
	}
	if len(pre) > 0 {
		pre = pre + " "
	}
	if len(suff) > 0 {
		suff = " " + suff
	}
	if numPres > 1 || numSuffs > 1 {
		return "###@@@INVALID@@@###"
	}
	return pre + baseName + suff
}

func (mods ModMap) Fits(item *Item) bool {
	if len(item.ExplicitFloatMods) != len(mods) {
		return false
	}
	for _, mod := range mods {
		itVal, ok := item.ExplicitFloatMods[mod.Text]
		if !ok {
			fmt.Println("Bad: ", mod.Text)
			return false
		}
		if mod.MinValue > itVal || mod.MaxValue < itVal {
			return false
		}
	}
	return true
}

func findAffixes(possibleOld AffixList, usedOld AffixList, item *Item) []AffixList {
	//fmt.Println("Already on: ", len(usedOld))
	result := make([]AffixList, 0)
	//Is the used list even valid?
	numPrefixes := 0
	numSuffixes := 0
	for i := 0; i < len(usedOld); i++ {
		if usedOld[i].IsPrefix {
			numPrefixes++
		} else {
			numSuffixes++
		}
	}
	usedMods := make(ModMap)
	for i := 0; i < len(usedOld); i++ {
		for j := 0; j < len(usedOld[i].Mods); j++ {
			mod := usedOld[i].Mods[j]
			oldVal, _ := usedMods[mod.Text]
			mod.MinValue += oldVal.MinValue
			mod.MaxValue += oldVal.MaxValue
			usedMods[mod.Text] = mod
		}
	}
	if usedMods.Fits(item) {
		used := make([]*Affix, len(usedOld))
		copy(used, usedOld)
		result = append(result, used)
	}
	//Find the affixes in possibleOld, that are still possible for the item, assuming usedOld Affixes are use
	usedGroups := make(map[int]bool)
	for i := 0; i < len(usedOld); i++ {
		//fmt.Println("Already in: ", usedOld[i].Name)
		usedGroups[usedOld[i].Group] = true
	}
	possible := make(AffixList, 0, len(possibleOld))
	for i := 0; i < len(possibleOld); i++ {
		/*isTestAffix := false
		if possibleOld[i].Name == "Burning" {
			isTestAffix = true
			fmt.Println("Working on testaffix")
		}*/
		_, groupAlreadyInUse := usedGroups[possibleOld[i].Group]
		if groupAlreadyInUse {
			continue
		}
		if (possibleOld[i].IsPrefix && numPrefixes == 3) || (!possibleOld[i].IsPrefix && numSuffixes == 3) {
			//can not add another
			continue
		}
		canAdd := true
		for j := 0; j < len(possibleOld[i].Mods); j++ {
			tested := possibleOld[i].Mods[j]
			curVal, _ := usedMods[tested.Text]
			itVal := item.ExplicitFloatMods[tested.Text]
			/*if isTestAffix {
				fmt.Println(itVal, curVal.MinValue, tested.MinValue)
				fmt.Println(itVal, curVal.MaxValue, tested.MaxValue)
			}*/
			if curVal.MinValue+tested.MinValue > itVal {
				canAdd = false
				break
			}
		}
		if canAdd {
			possible = append(possible, possibleOld[i])
		}
	}
	/*if len(possible) > 0 {
		fmt.Println("Has: ", usedOld, " Can: ", possible)
	}*/
	if len(possible) == 0 {
		return result //Contains itself, if set of affixes was fitting (determined at start)
	}
	//For all possible mods find the one with the least corresponding affixes
	corrAffixes := make(map[string]AffixList)
	for i := 0; i < len(possible); i++ {
		for j := 0; j < len(possible[i].Mods); j++ {
			mod := possible[i].Mods[j]
			curr, ok := corrAffixes[mod.Text]
			if !ok {
				curr = make(AffixList, 0)
			}
			curr = append(curr, possible[i])
			corrAffixes[mod.Text] = curr
		}
	}
	//For each ModName, corrAffixes now lists the corrseponding affixes
	//Find the most vulnerable mod, that is the one, that can be fullfilled by the least amount of affixes
	var minList AffixList
	var minModName string
	for name, list := range corrAffixes {
		if len(minList) == 0 || len(list) < len(minList) {
			minList = list
			minModName = name
		}
	}
	//minList now is our best shot. Try setting every affix as used and then see if the item can be completed recursively
	for i := 0; i < len(minList); i++ {
		//fmt.Println("Trying: ", minList[i].Name)
		used := make(AffixList, len(usedOld), len(usedOld)+1)
		copy(used, usedOld)
		used = append(used, minList[i])
		result = append(result, findAffixes(possible, used, item)...)
	}
	//What can happen too, is that the mod is fullfilled -> remove the affixes giving to the mod from the possible-list
	mod, ok := usedMods[minModName]
	if ok {
		if mod.MaxValue >= item.ExplicitFloatMods[minModName] {
			//We have such a case
			possibleReduced := make(AffixList, 0)
			for i := 0; i < len(possible); i++ {
				cut := false
				for j := 0; j < len(minList); j++ {
					if minList[j] == possible[i] {
						cut = true
						break
					}
				}
				if !cut {
					possibleReduced = append(possibleReduced, possible[i])
				}
			}
			result = append(result, findAffixes(possibleReduced, usedOld, item)...)
		}
	}
	return result
}

func removeDuplicates(in []AffixList) []AffixList {
	result := make([]AffixList, 0, len(in))
	known := make(map[string]bool)
	for i := 0; i < len(in); i++ {
		name := in[i].String()
		_, duplicate := known[name]
		if !duplicate {
			known[name] = true
			result = append(result, in[i])
		}
	}
	return result
}

func FindAffixes(item *Item) []AffixList {
	// First strip off the affixes that are not useable on this item
	possible := make([]*Affix, 0, len(affixList))
	for i := 0; i < len(affixList); i++ {
		affix := affixList[i]
		if item.ItemLevel >= float64(affix.Level) && affix.Useability.Fits(item) {
			possible = append(possible, affix)
		}
	}
	used := make([]*Affix, 0)
	//Now recursively try to add affixes
	result := removeDuplicates(findAffixes(possible, used, item))
	//If item is a magic one, we can further reduce the candidates, due to itemname being known
	if item.Rarity == "Magic" {
		temp := make([]AffixList, 0)
		for i := 0; i < len(result); i++ {
			if result[i].magicName(item.BaseType) == item.Name {
				temp = append(temp, result[i])
			}
		}
		result = temp
	}
	return result
}

func (use *Useability) Fits(item *Item) bool {
	var txt string
	switch item.Class {
	case "Amulet":
		txt = use.values[1]
	case "Belt":
		txt = use.values[2]
	case "Body Armour":
		txt = use.values[6]
	case "Boots":
		txt = use.values[5]
	case "Bow":
		txt = use.values[14]
	case "Claw":
		txt = use.values[11]
	case "Dagger":
		txt = use.values[10]
	case "Gloves":
		txt = use.values[4]
	case "Helmet":
		txt = use.values[3]
	case "Jewel": //Todo
		return false
	case "One Handed Axe", "One Handed Sword", "Thrusting One Handed Sword":
		txt = use.values[15]
	case "One Handed Mace":
		txt = use.values[17]
	case "Quiver":
		txt = use.values[8]
	case "Ring":
		txt = use.values[0]
	case "Sceptre":
		txt = use.values[12]
	case "Shield":
		txt = use.values[7]
	case "Staff":
		txt = use.values[13]
	case "Two Handed Axe", "Two Handed Sword":
		txt = use.values[16]
	case "Two Handed Mace":
		txt = use.values[18]
	case "Wand":
		txt = use.values[9]
	default:
		return false
	}
	switch txt {
	case "Yes":
		return true
	case "No":
		return false
	case "Yes (str)":
		return item.Armour > 0
	case "Yes (int)":
		return item.EnergyShield > 0
	case "Yes (dex)":
		return item.Evasion > 0
	case "Yes (str-only)":
		return item.Armour > 0 && item.Evasion == 0 && item.EnergyShield == 0
	case "Yes (int-only)":
		return item.Armour == 0 && item.Evasion == 0 && item.EnergyShield > 0
	case "Yes (dex-only)":
		return item.Armour == 0 && item.Evasion > 0 && item.EnergyShield == 0
	case "Yes (int-str)":
		return item.Armour > 0 && item.Evasion == 0 && item.EnergyShield > 0
	case "Yes (int-dex)":
		return item.Armour == 0 && item.Evasion > 0 && item.EnergyShield > 0
	case "Yes (dex-str)":
		return item.Armour > 0 && item.Evasion > 0 && item.EnergyShield == 0
	case "Yes (not Int)":
		return item.EnergyShield == 0
	case "Triple-Only":
		return item.Armour > 0 && item.Evasion > 0 && item.EnergyShield > 0
	default:
		fmt.Println("Invalid usability: ", txt)
		log.Println("Invalid usability: ", txt)
		return false
	}
}

func initError(a ...interface{}) {
	fmt.Println(a)
	log.Println(a)
	os.Exit(1)
}

func newUsability(in []string) *Useability {
	if len(in) != 19 {
		initError("Error reading affixes.csv: Usability params count invalid.")
	}
	result := new(Useability)
	for i := 0; i < len(in); i++ {
		result.values[i] = in[i]
		if in[i] != "Yes" && in[i] != "No" && in[i] != "Yes (str)" && in[i] != "Yes (int)" && in[i] != "Yes (dex)" &&
			in[i] != "Yes (str-only)" && in[i] != "Yes (int-only)" && in[i] != "Yes (dex-only)" &&
			in[i] != "Yes (int-str)" && in[i] != "Yes (int-dex)" && in[i] != "Yes (dex-str)" &&
			in[i] != "Yes (not Int)" && in[i] != "Triple-Only" {
			initError("Error reading affixes.csv: Usability Yes/No: Invalid ", in[i], ".")
		}
	}
	return result
}

func newMod(txt, values string) Mod {
	var result Mod
	vs := strings.Split(values, " to ")
	min, err := strconv.ParseFloat(vs[0], 64)
	if err != nil {
		initError("Error reading affixes.csv: Parsing mod ", txt, ": Could not read values: ", err)
	}
	result.MinValue = min
	if len(vs) > 1 {
		max, err := strconv.ParseFloat(vs[1], 64)
		if err != nil {
			initError("Error reading affixes.csv: Parsing mod ", txt, ": Could not read values: ", err)
		}
		result.MaxValue = max
	} else {
		result.MaxValue = result.MinValue
	}
	result.Text = txt
	if result.MinValue < 0 {
		result.MinValue = -result.MinValue
	}
	if result.MaxValue < 0 {
		result.MaxValue = -result.MaxValue
	}
	return result
}

func init() {
	descToText = make(map[string]string)
	descToText["Base Min Added Cold Dmg / Base Max Added Cold Dmg"] = "AddedMinColdAttackDamage / AddedMaxColdAttackDamage"
	descToText["Base Min Added Fire Dmg / Base Max Added Fire Dmg"] = "AddedMinFireAttackDamage / AddedMaxFireAttackDamage"
	descToText["Base Min Added Lightning Dmg / Base Max Added Lightning Dmg"] = "AddedMinLightningAttackDamage / AddedMaxLightningAttackDamage"
	descToText["Base Min Added Physical Dmg / Base Max Added Physical Dmg"] = "AddedMinPhysicalAttackDamage / AddedMaxPhysicalAttackDamage"
	descToText["Local Min Added Chaos Dmg / Local Max Added Chaos Dmg"] = "AddedMinChaosDamage / AddedMaChaosDamage"
	descToText["Local Min Added Physical Dmg / Local Max Added Physical Dmg"] = "AddedMinPhysicalDamage / AddedMaxPhysicalDamage"
	descToText["Local Min Added Cold Dmg / Local Max Added Cold Dmg"] = "AddedMinColdDamage / AddedMaxColdDamage"
	descToText["Local Min Added Lightning Dmg / Local Max Added Lightning Dmg"] = "AddedMinLightningDamage / AddedMaxLightningDamage"
	descToText["Local Min Added Fire Dmg / Local Max Added Fire Dmg"] = "AddedMinFireDamage / AddedMaxFireDamage"
	descToText["Physical Dmg To Return To Melee Attacker"] = "PhysicalMeleeReflect"
	descToText["Local Physical Dmg +%"] = "IncreasedPhysicalDamage"
	descToText["Local Physical Dmg +% / Local Accuracy Rating"] = "IncreasedPhysicalDamage / AddedAccuracy"
	descToText["Armor Rating"] = "AddedArmour"
	descToText["Local Armor Rating"] = "AddedArmour"
	descToText["Local Armor +%"] = "IncreasedArmour"
	descToText["Armor Rating +%"] = "IncreasedArmour"
	descToText["Local Armor +% / Base Stun Recovery +%"] = "IncreasedArmour / IncreasedStunRecovery"
	descToText["Base Max Energy Shield"] = "AddedEnergyShield"
	descToText["Local Energy Shield"] = "AddedEnergyShield"
	descToText["Local Energy Shield +%"] = "IncreasedEnergyShield"
	descToText["Max Energy Shield +%"] = "IncreasedEnergyShield"
	descToText["Local Energy Shield +% / Base Stun Recovery +%"] = "IncreasedEnergyShield / IncreasedStunRecovery"
	descToText["Base Evasion Rating"] = "AddedEvasion"
	descToText["Local Evasion Rating"] = "AddedEvasion"
	descToText["Evasion Rating +%"] = "IncreasedEvasion"
	descToText["Local Evasion Rating +%"] = "IncreasedEvasion"
	descToText["Local Evasion Rating +% / Base Stun Recovery +%"] = "IncreasedEvasion / IncreasedStunRecovery"
	descToText["Flask Life Recovery Rate +%"] = "IncreasedFlaskLifeRecoveryRate"
	descToText["Flask Mana Recovery Rate +%"] = "IncreasedFlaskManaRecoveryRate"
	descToText["Local Socketed Bow Gem Level +"] = "AddedBowGemLevel"
	descToText["Local Socketed Cold Gem Level +"] = "AddedColdGemLevel"
	descToText["Local Socketed Fire Gem Level +"] = "AddedFireGemLevel"
	descToText["Local Socketed Lightning Gem Level +"] = "AddedLightningGemLevel"
	descToText["Local Socketed Melee Gem Level +"] = "AddedMeleeGemLevel"
	descToText["Local Socketed Minion Gem Level +"] = "AddedMinionGemLevel"
	descToText["Local Socketed Gem Level +"] = "AddedGemLevel"
	descToText["Local Armour And Energy Shield +%"] = "IncreasedArmourEnergyShield"
	descToText["Local Armour And Evasion +%"] = "IncreasedArmourEvasion"
	descToText["Local Evasion And Energy Shield +%"] = "IncreasedEvasionEnergyShield"
	descToText["Local Armour And Evasion And Energy Shield +%"] = "IncreasedArmourEvasionEnergyShield"
	descToText["Local Armour And Energy Shield +% / Base Stun Recovery +%"] = "IncreasedArmourEnergyShield / IncreasedStunRecovery"
	descToText["Local Armour And Evasion +% / Base Stun Recovery +%"] = "IncreasedArmourEvasion / IncreasedStunRecovery"
	descToText["Local Evasion And Energy Shield +% / Base Stun Recovery +%"] = "IncreasedEvasionEnergyShield / IncreasedStunRecovery"
	descToText["Local Armour And Evasion And Energy Shield +% / Base Stun Recovery +%"] = "IncreasedArmourEvasionEnergyShield / IncreasedStunRecovery"
	descToText["Base Item Found Rarity +%"] = "IncreasedItemRarity"
	descToText["Base Max Life"] = "AddedLife"
	descToText["Life Leech From Physical Dmg %"] = "PhysicalAttackLifeLeech"
	descToText["Base Max Mana"] = "AddedMana"
	descToText["Mana Leech From Physical Dmg %"] = "PhysicalAttackManaLeech"
	descToText["Base Movement Velocity +%"] = "IncreasedMovementSpeed"
	descToText["Spell Dmg +%"] = "IncreasedSpellDamage"
	descToText["Weapon Elemental Dmg +%"] = "IncreasedWeaponElementalDamage"
	descToText["Spell Dmg +% / Base Max Mana"] = "IncreasedSpellDamage / AddedMana"
	descToText["Projectile speed +%"] = "IncreasedProjectileSpeed"
	descToText["Accuracy Rating"] = "AddedAccuracy"
	descToText["Local Accuracy Rating"] = "AddedAccuracy"
	descToText["Light Radius / +Accuracy Rating"] = "IncreasedLightRadius / AddedAccuracy"
	descToText["Light Radius / Accuracy Rating %"] = "IncreasedLightRadius / IncreasedAccuracy"
	descToText["Attack Speed +%"] = "IncreasedAttackSpeed"
	descToText["Local Attack Speed +%"] = "IncreasedAttackSpeed"
	descToText["Additional All Attributes"] = "AddedAllAttributes"
	descToText["Additional Dexterity"] = "AddedDexterity"
	descToText["Additional Intelligence"] = "AddedIntelligence"
	descToText["Additional Strength"] = "AddedStrength"
	descToText["Base Cast Speed +%"] = "IncreasedCastSpeed"
	descToText["Base Critical Strike Multiplier +%"] = "IncreasedGlobalCritMultiplier"
	descToText["Weapon-only Critical Strike Multiplier +%"] = "IncreasedGlobalCritMultiplier"
	descToText["Critical Strike Chance +%"] = "IncreasedGlobalCritChance"
	descToText["Local Critical Strike Chance +%"] = "IncreasedLocalCritChance"
	descToText["Spell Critical Strike Chance +%"] = "IncreasedSpellCritChance"
	descToText["Cold Dmg +%"] = "IncreasedColdDamage"
	descToText["Fire Dmg +%"] = "IncreasedFireDamage"
	descToText["Lightning Dmg +%"] = "IncreasedLightningDamage"
	descToText["Charges Gained +%"] = "IncreasedFlaskChargesGained"
	descToText["Flask Charges Used +%"] = "ReducedFlaskChargesUsed"
	descToText["Flask Duration +%"] = "IncreasedFlaskDuration"
	descToText["Life Gain Per Target"] = "LifePerAttackHit"
	descToText["Life Gained On Enemy Death"] = "LifePerKill"
	descToText["Base Life Regeneration Rate Per Second"] = "LifePerSecond"
	descToText["Mana Gained On Enemy Death"] = "ManaPerKill"
	descToText["Mana Regeneration Rate +%"] = "IncreasedManaRegeneration"
	descToText["Base Cold Dmg Resistance %"] = "ColdResistance"
	descToText["Base Fire Dmg Resistance %"] = "FireResistance"
	descToText["Base Lightning Dmg Resistance %"] = "LightningResistance"
	descToText["Base Chaos Dmg Resistance %"] = "ChaosResistance"
	descToText["Base Resist All Elements %"] = "ElementalResistances"
	descToText["Local Attribute Requirements -%"] = "ReducedAttributeRequirements"
	descToText["Local Additional Block Chance %"] = "AddedBlockChance"
	descToText["Base Stun Duration +%"] = "IncreasedEnemyStunDuration"
	descToText["Base Stun Threshold Reduction +%"] = "ReducedEnemyStunThreshold"
	descToText["Base Stun Recovery +%"] = "IncreasedStunRecovery"
	descToText["Spell Minimum Added Cold Damage / Spell Maximum Added Cold Damage"] = "AddedMinColdSpellDamage / AddedMaxColdSpellDamage"
	descToText["Spell Minimum Added Fire Damage / Spell Maximum Added Fire Damage"] = "AddedMinFireSpellDamage / AddedMaxFireSpellDamage"
	descToText["Spell Minimum Added Lightning Damage / Spell Maximum Added Lightning Damage"] = "AddedMinLightningSpellDamage / AddedMaxLightningSpellDamage"
}

func init() {
	data, err := ioutil.ReadFile("data/affixes.csv")
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		os.Exit(1)
	}
	datastr := strings.Replace(string(data), "\r", "", -1)
	lines := strings.Split(datastr, "\n")
	groupCounter := 0
	knownGroups := make(map[string]int)
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) == 0 {
			continue
		}
		affix := new(Affix)
		columns := strings.Split(lines[i], ";")
		if len(columns) != 24 {
			initError("Error reading affixes.csv: Invalid line ", i+1, ": Does not have the required column count.")
		}
		if columns[4] == "Prefix" {
			affix.IsPrefix = true
		} else if columns[4] == "Suffix" {
			affix.IsPrefix = false
		} else {
			initError("Error reading affixes.csv: Invalid line ", i+1, ": Could not determine whether Prefix or Suffix.")
		}
		txt, ok := descToText[columns[0]]
		if !ok {
			initError("Error reading affixes.csv: Invalid line ", i+1, ": Unknown affix description \""+columns[0]+"\"")
		}
		groupId, ok := knownGroups[txt+" "+columns[4]]
		if !ok {
			affix.Group = groupCounter
			knownGroups[txt+" "+columns[4]] = groupCounter
			groupCounter++
		} else {
			affix.Group = groupId
		}
		mods := strings.Split(txt, " / ")
		values := strings.Split(columns[1], " / ")
		if len(mods) != len(values) {
			initError("Error reading affixes.csv: Invalid line ", i+1, ": Amount of modifiers and values mismatch.")
		}
		affix.Mods = make([]Mod, len(mods))
		for j := 0; j < len(mods); j++ {
			affix.Mods[j] = newMod(mods[j], values[j])
		}
		affix.Name = columns[2]
		lvl, err := strconv.ParseInt(columns[3], 10, 64)
		if err != nil {
			initError("Error reading affixes.csv: Invalid line ", i+1, ": AffixLevel must be a valid integer.")
		}
		affix.Level = int(lvl)
		affix.Useability = newUsability(columns[5:])
		affixList = append(affixList, affix)
	}
}
