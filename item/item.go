package item

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type SocketGroup struct {
	R int
	G int
	B int
	W int
}

type ItemModifier struct {
	/*IncreasedArmourEvasion     float64
	PhysicalAttackLeech        float64
	AttackDodge                float64
	AddedLife                  float64
	LooseEnduranceChargesOnHit bool*/
}

type Item struct {
	Rarity       string
	Name         string
	Class        string
	BaseType     string
	IsIdentified bool
	IsCorrupted  bool
	//Requirements
	LvlReq float64
	StrReq float64
	IntReq float64
	DexReq float64
	//Properties
	Quality            float64
	MinPhysicalDamage  float64
	MaxPhysicalDamage  float64
	MinElementalDamage float64
	MaxElementalDamage float64
	CritChance         float64
	APS                float64
	Armour             float64
	Evasion            float64
	EnergyShield       float64
	BlockChance        float64
	//Sockets
	Sockets []SocketGroup
	//Item Level
	ItemLevel float64
	//Flavour Text
	FlavourText string
	//Modifiers
	ImplicitFloatMods             map[string]float64
	ImplicitBoolMods              map[string]bool
	ExplicitFloatMods             map[string]float64
	ExplicitBoolMods              map[string]bool
	MinPrefixes                   float64
	MaxPrefixes                   float64
	MinSuffixes                   float64
	MaxSuffixes                   float64
	HasMultipleAffixPossibilities bool
	PossibleAffixes               []AffixList
}

var baseTypes map[string]string

var classes []string //{"Ring", "Belt", "Amulet"}
var ItemFloatMods map[string]string
var ItemBoolMods map[string]string

//Properties
var PHYS_DAMAGE_REGEXP = regexp.MustCompile(`^Physical Damage: (\d+)-(\d+)$`)
var ELE_DAMAGE_REGEXP = regexp.MustCompile(`^Elemental Damage: ((\d+)-(\d+)(, \d+-\d+)*)$`)
var CRIT_CHANCE_REGEXP = regexp.MustCompile(`^Critical Strike Chance: (\d+\.?\d*)%$`)
var APS_REGEXP = regexp.MustCompile(`^Attacks per Second: (\d+\.?\d*)$`)
var QUALITY_REGEXP = regexp.MustCompile(`^Quality: \+(\d+)%$`)
var ARMOUR_REGEXP = regexp.MustCompile(`^Armour: (\d+)$`)
var EVASION_REGEXP = regexp.MustCompile(`^Evasion Rating: (\d+)$`)
var ENERGY_SHIELD_REGEXP = regexp.MustCompile(`^Energy Shield: (\d+)$`)
var BLOCK_CHANCE_REGEXP = regexp.MustCompile(`^Chance to Block: (\d+)%$`)

//Requirements
var LEVEL_REQUIREMENT_REGEXP = regexp.MustCompile(`^Level: (\d+)`) //May be followed by an " (unmet)"
var STR_REQUIREMENT_REGEXP = regexp.MustCompile(`^Str(ength)?: (\d+)`)
var INT_REQUIREMENT_REGEXP = regexp.MustCompile(`^Int(elligence)?: (\d+)`)
var DEX_REQUIREMENT_REGEXP = regexp.MustCompile(`^Dex(terity)?: (\d+)`)

var FLOAT_REGEXP = regexp.MustCompile(`\d+\.?\d*`)

func (it *Item) GetFloatModValue(key string) float64 {
	implicit, _ := it.ImplicitFloatMods[key]
	explicit, _ := it.ExplicitFloatMods[key]
	/*if strings.HasPrefix(key, "IMPLICIT") {
		return implicit
	} else if strings.HasPrefix(key, "EXPLICIT") {
		return explicit
	}*/
	return implicit + explicit
}

func (it *Item) GetBoolModValue(key string) bool {
	implicit, _ := it.ImplicitBoolMods[key]
	explicit, _ := it.ExplicitBoolMods[key]
	/*if strings.HasPrefix(key, "IMPLICIT") {
		return implicit
	} else if strings.HasPrefix(key, "EXPLICIT") {
		return explicit
	}*/
	return implicit || explicit
}

func (group SocketGroup) Size() float64 {
	return float64(group.R + group.B + group.G + group.W)
}

func (it *Item) SocketCount() float64 {
	result := 0.
	for i := 0; i < len(it.Sockets); i++ {
		result += it.Sockets[i].Size()
	}
	return result
}

func (it *Item) RedSocketCount() float64 {
	result := 0.
	for i := 0; i < len(it.Sockets); i++ {
		result += float64(it.Sockets[i].R)
	}
	return result
}

func (it *Item) GreenSocketCount() float64 {
	result := 0.
	for i := 0; i < len(it.Sockets); i++ {
		result += float64(it.Sockets[i].G)
	}
	return result
}

func (it *Item) BlueSocketCount() float64 {
	result := 0.
	for i := 0; i < len(it.Sockets); i++ {
		result += float64(it.Sockets[i].B)
	}
	return result
}

func (it *Item) WhiteSocketCount() float64 {
	result := 0.
	for i := 0; i < len(it.Sockets); i++ {
		result += float64(it.Sockets[i].W)
	}
	return result
}

func (it *Item) MaxLinks() float64 {
	max := 0.
	for i := 0; i < len(it.Sockets); i++ {
		size := it.Sockets[i].Size()
		if size > max {
			max = size
		}
	}
	return max
}

func ParseRarity(lines []string, item *Item) (bool, error) {
	if len(lines) < 2 {
		return false, nil
	}
	if strings.HasPrefix(lines[len(lines)-1], "Superior ") {
		lines[1] = lines[len(lines)-1][9:]
	}
	switch lines[0] {
	case "Rarity: Quest":
		return true, errors.New("Quest item not yet implemented")
	case "Rarity: Normal":
		item.Rarity = "Normal"
		item.Name = lines[1]
		if AddBaseAndClassToItem(lines[1], item) {
			return true, nil
		} else {
			return false, errors.New("Unknown item base type '" + lines[1] + "'")
		}
	case "Rarity: Magic":
		item.Rarity = "Magic"
		item.Name = lines[1]
		parts := strings.Split(lines[1], " ")
		for cutLeft := 0; cutLeft < len(parts); cutLeft++ {
			for cutRight := len(parts); cutRight > 0; cutRight-- {
				name := ""
				for i := cutLeft; i < cutRight; i++ {
					if i != cutLeft {
						name += " "
					}
					name += parts[i]
				}
				if AddBaseAndClassToItem(name, item) {
					return true, nil
				}
			}
		}
		return false, errors.New("Unknown item base type " + lines[1])
	case "Rarity: Rare":
		item.Rarity = "Rare"
		if AddBaseAndClassToItem(lines[len(lines)-1], item) {
			if len(lines) == 3 {
				item.Name = lines[1]
			}
			return true, nil
		} else {
			return false, errors.New("Unknown item base type '" + lines[len(lines)-1] + "'")
		}
	case "Rarity: Unique":
		item.Rarity = "Unique"
		if len(lines) < 3 {
			return false, errors.New("Unique item is lacking either name or basetype.")
		}
		if AddBaseAndClassToItem(lines[2], item) {
			item.Name = lines[1]
			return true, nil
		} else {
			return false, errors.New("Unknown item base type '" + lines[2] + "'")
		}
	default:
		return false, nil
	}
	//return false
}

func ParseProperties(lines []string, item *Item) (bool, error) {
	//First line may be a class, further lines like physical damage, crit chance, etc.
	matchingLineFound := false
	start := 0
	for i := 0; i < len(classes); i++ {
		if classes[i] == lines[0] {
			start = 1
			matchingLineFound = true
			break
		}
	}
	invalids := make([]string, 0)
	for i := start; i < len(lines); i++ {
		line := strings.Replace(lines[i], " (augmented)", "", -1)
		var tmpFloat float64
		var err error
		if match := PHYS_DAMAGE_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.MinPhysicalDamage = tmpFloat
			if tmpFloat, err = strconv.ParseFloat(match[2], 64); err != nil {
				return true, err
			}
			item.MaxPhysicalDamage = tmpFloat
		} else if match := ELE_DAMAGE_REGEXP.FindStringSubmatch(line); match != nil {
			dmgs := strings.Split(match[1], ", ")
			for j := 0; j < len(dmgs); j++ {
				minmax := strings.Split(dmgs[j], "-")
				if len(minmax) != 2 {
					return true, errors.New("Invalid line :" + lines[i])
				}
				if tmpFloat, err = strconv.ParseFloat(minmax[0], 64); err != nil {
					return true, err
				}
				item.MinElementalDamage += tmpFloat
				if tmpFloat, err = strconv.ParseFloat(minmax[1], 64); err != nil {
					return true, err
				}
				item.MaxElementalDamage += tmpFloat
			}
		} else if match := CRIT_CHANCE_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.CritChance = tmpFloat
		} else if match := APS_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.APS = tmpFloat
		} else if match := QUALITY_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.Quality = tmpFloat
		} else if match := ARMOUR_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.Armour = tmpFloat
		} else if match := EVASION_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.Evasion = tmpFloat
		} else if match := BLOCK_CHANCE_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.BlockChance = tmpFloat
		} else if match := ENERGY_SHIELD_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.EnergyShield = tmpFloat
		} else {
			invalids = append(invalids, line)
			continue //don't go to the "found a line" at the end
		}
		matchingLineFound = true
	}
	if matchingLineFound {
		if len(invalids) > 0 {
			fmt.Println("Found Invalid Properties in Property block: ", invalids)
			log.Println("Found Invalid Properties in Property block: ", invalids)
		}
		return true, nil
	} else {
		return false, nil
	}
}

func ParseRequirements(lines []string, item *Item) (bool, error) {
	if lines[0] != "Requirements:" {
		return false, nil
	}
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		var tmpFloat float64
		var err error
		if match := LEVEL_REQUIREMENT_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[1], 64); err != nil {
				return true, err
			}
			item.LvlReq = tmpFloat
		} else if match := STR_REQUIREMENT_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[2], 64); err != nil {
				return true, err
			}
			item.StrReq = tmpFloat
		} else if match := INT_REQUIREMENT_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[2], 64); err != nil {
				return true, err
			}
			item.IntReq = tmpFloat
		} else if match := DEX_REQUIREMENT_REGEXP.FindStringSubmatch(line); match != nil {
			if tmpFloat, err = strconv.ParseFloat(match[2], 64); err != nil {
				return true, err
			}
			item.DexReq = tmpFloat
		} else {
			fmt.Println("Warning, unknown item requirement: ", line)
			log.Println("Warning, unknown item requirement: ", line)
		}
	}
	return true, nil
}

func ParseSocketGroup(text string) (SocketGroup, error) {
	var result SocketGroup
	sockets := strings.Split(text, "-")
	for i := 0; i < len(sockets); i++ {
		switch sockets[i] {
		case "R":
			result.R++
		case "G":
			result.G++
		case "B":
			result.B++
		case "W":
			result.W++
		default:
			return result, errors.New("Unknown Socket colour: '" + sockets[i] + "'.")
		}
	}
	return result, nil
}

func ParseSockets(lines []string, item *Item) (bool, error) {
	line := lines[0]
	if len(line) < 9 {
		return false, nil
	}
	if line[:9] != "Sockets: " {
		return false, nil
	}
	sockets := line[9:]
	groups := strings.Split(sockets, " ")
	for i := 0; i < len(groups); i++ {
		if len(groups[i]) > 0 {
			group, err := ParseSocketGroup(groups[i])
			if err != nil {
				return true, err
			}
			item.Sockets = append(item.Sockets, group)
		}
	}
	return true, nil
}

func ParseItemlevel(lines []string, item *Item) (bool, error) {
	line := lines[0]
	if len(line) < 12 {
		return false, nil
	}
	if line[:12] != "Item Level: " {
		return false, nil
	}
	lvl := line[12:]
	tmp, err := strconv.ParseInt(lvl, 10, 64)
	if err != nil {
		return true, err
	}
	item.ItemLevel = float64(tmp)
	return true, nil
}

func AddBaseAndClassToItem(base string, item *Item) bool {
	class, ok := baseTypes[base]
	if !ok {
		return false
	} else {
		item.BaseType = base
		item.Class = class
		return true
	}
}

func ParseFlavour(lines []string, item *Item) (bool, error) {
	if len(lines) == 1 && lines[0] == "Corrupted" {
		return false, nil
	}
	for i := 0; i < len(lines); i++ {
		if i != 0 {
			item.FlavourText += "\n"
		}
		item.FlavourText += lines[i]
	}
	return true, nil
}

// Tries to set a modifier by a line like "+# to maximum Life", returns whether such a line actually existed
func setFloatModByLine(mods map[string]float64, line string, value string) bool {
	name, ok := ItemFloatMods[line] //Name like AddedColdDamage
	if !ok {
		return false
	}
	var tmpFloat float64
	var err error
	if tmpFloat, err = strconv.ParseFloat(value, 64); err != nil {
		fmt.Println("setModByLine: ", err) //Should not be possible, as the value just passed our float regexp
		log.Println("setModByLine: ", err)
	}
	oldVal, _ := mods[name] //Defaults to 0 if non-existant
	mods[name] = oldVal + tmpFloat
	return true
}

func setBoolModByLine(mods map[string]bool, line string) bool {
	name, ok := ItemBoolMods[line] //Name like AddedColdDamage
	if !ok {
		return false
	}
	mods[name] = true
	return true
}

func ParseMods(lines []string, item *Item) (bool, error) {
	floatMods := make(map[string]float64)
	boolMods := make(map[string]bool)
	foundMod := false
	invalids := make([]string, 0)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			continue
		}
		sanitized := FLOAT_REGEXP.ReplaceAllString(line, "#")
		matches := FLOAT_REGEXP.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			if setBoolModByLine(boolMods, sanitized) {
				foundMod = true
			} else {
				invalids = append(invalids, line)
			}
		} else if len(matches) == 1 {
			if setFloatModByLine(floatMods, sanitized, matches[0][0]) {
				foundMod = true
			} else {
				invalids = append(invalids, line)
			}
		} else if len(matches) == 2 {
			san1 := strings.Replace(strings.Replace(sanitized, "#", "?", -1), "?", "#", 1) //Adds #-? to Cold Damage
			san2 := strings.Replace(sanitized, "#", "?", 1)                                //Adds ?-# to Cold Damage
			if setFloatModByLine(floatMods, san1, matches[0][0]) {
				foundMod = true
			} else {
				invalids = append(invalids, line)
			}
			if setFloatModByLine(floatMods, san2, matches[1][0]) {
				foundMod = true
			} else {
				invalids = append(invalids, line)
			}
		}
	}
	if foundMod {
		if item.ExplicitFloatMods == nil {
			item.ExplicitFloatMods = floatMods
			item.ExplicitBoolMods = boolMods
		} else if item.Rarity == "Normal" {
			item.ImplicitFloatMods = floatMods
			item.ImplicitBoolMods = boolMods
		} else {
			item.ImplicitFloatMods = item.ExplicitFloatMods
			item.ImplicitBoolMods = item.ExplicitBoolMods
			item.ExplicitFloatMods = floatMods
			item.ExplicitBoolMods = boolMods
		}
		if len(invalids) > 0 && item.Rarity != "Unique" {
			fmt.Println("Found ", len(invalids), " Invalid Mods: ", invalids)
			log.Println("Found ", len(invalids), " Invalid Mods: ", invalids)
		}
		return true, nil
	} else {
		return false, nil
	}

}

func ParseUnidentified(lines []string, item *Item) (bool, error) {
	if len(lines) < 1 {
		item.IsIdentified = true
		return false, nil
	}
	if lines[0] != "Unidentified" {
		item.IsIdentified = true
		return false, nil
	} else {
		item.ImplicitBoolMods = item.ExplicitBoolMods
		item.ImplicitFloatMods = item.ExplicitFloatMods
		item.ExplicitBoolMods = make(map[string]bool)
		item.ExplicitFloatMods = make(map[string]float64)
		item.IsIdentified = false
		return true, nil
	}
}

func ParseCorrupted(lines []string, item *Item) (bool, error) {
	if len(lines) < 1 {
		item.IsCorrupted = false
		return false, nil
	}
	if lines[0] != "Corrupted" {
		item.IsCorrupted = false
		return false, nil
	} else {
		item.IsCorrupted = true
		return true, nil
	}
}

func createItemMapsIfNil(it *Item) {
	if it.ExplicitFloatMods == nil {
		it.ExplicitFloatMods = make(map[string]float64)
	}
	if it.ExplicitBoolMods == nil {
		it.ExplicitBoolMods = make(map[string]bool)
	}
	if it.ImplicitFloatMods == nil {
		it.ImplicitFloatMods = make(map[string]float64)
	}
	if it.ImplicitBoolMods == nil {
		it.ImplicitBoolMods = make(map[string]bool)
	}
}

func (item *Item) SetPrefixSuffixCount() {
	item.HasMultipleAffixPossibilities = (len(item.PossibleAffixes) > 1)
	item.MinSuffixes = 99999
	item.MinPrefixes = 99999
	for i := 0; i < len(item.PossibleAffixes); i++ {
		//fmt.Println(i, ": ", result.PossibleAffixes[i])
		preCount := float64(0)
		suffCount := float64(0)
		for j := 0; j < len(item.PossibleAffixes[i]); j++ {
			if item.PossibleAffixes[i][j].IsPrefix {
				preCount++
			} else {
				suffCount++
			}
		}
		if suffCount > item.MaxSuffixes {
			item.MaxSuffixes = suffCount
		}
		if suffCount < item.MinSuffixes {
			item.MinSuffixes = suffCount
		}
		if preCount > item.MaxPrefixes {
			item.MaxPrefixes = preCount
		}
		if preCount < item.MinPrefixes {
			item.MinPrefixes = preCount
		}
	}
}

func ParseItem(text string) (*Item, error) {
	var result Item
	result.Sockets = make([]SocketGroup, 0)
	result.IsIdentified = true //Might not reach a "Unidentified" block
	text = strings.Replace(text, "\r", "", -1)
	blocks := strings.Split(text, "\n--------\n")
	lines := make([][]string, len(blocks))
	for i := 0; i < len(blocks); i++ {
		lines[i] = strings.Split(blocks[i], "\n")
		if len(lines[i]) < 1 {
			return nil, errors.New("Empty property block for item. (Two '--------' followed one another")
		}
	}
	currIndex := 0
	funcs := []func([]string, *Item) (bool, error){ParseRarity, ParseProperties, ParseRequirements, ParseSockets, ParseItemlevel, ParseMods, ParseMods, ParseUnidentified, ParseFlavour, ParseCorrupted}
	for i := 0; i < len(funcs); i++ {
		if currIndex >= len(blocks) {
			break
		}
		matched, err := funcs[i](lines[currIndex], &result)
		if err != nil {
			return nil, err
		}
		if matched {
			currIndex++
		} else {
			//fmt.Println("Function ", i, " didn't match", lines[currIndex])
		}
	}
	if currIndex != len(blocks) {
		//-1, because last block (flavour text) always matches
		fmt.Println("Item contains unparsed block: ", blocks[currIndex-1])
		log.Println("Item contains unparsed blocks.", blocks[currIndex-1])
		return nil, errors.New("Item contains uparsed blocks.")
	}
	createItemMapsIfNil(&result)
	//fmt.Println("FloatMods: ", result.ExplicitFloatMods)
	if (result.Rarity == "Magic" || result.Rarity == "Rare") && (result.Class != "Jewel") {
		result.PossibleAffixes = FindAffixes(&result)
		if len(result.PossibleAffixes) == 0 { //Code failed, as it SHOULD be able to detect all possibilities
			fmt.Println("No possible affix-combination found for item with the following explicits: ", result.ExplicitFloatMods)
			log.Println("No possible affix-combination found for item with the following explicits: ", result.ExplicitFloatMods)
			return nil, errors.New("Could not determine affixes.")
		}
		result.SetPrefixSuffixCount()
	}
	return &result, nil
}

func AddBaseTypes(data []byte, base string) {
	refined := strings.Replace(string(data), "\r", "", -1)
	lines := strings.Split(refined, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if len(line) > 0 {
			if line[0] != '#' {
				baseTypes[line] = base
			}
		}
	}
}

func init() {
	baseTypes = make(map[string]string)
	classes = make([]string, 0)
	files, err := ioutil.ReadDir("data/ItemClasses")
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		os.Exit(1)
	}
	for i := 0; i < len(files); i++ {
		name := files[i].Name()
		name = strings.Replace(name, ".txt", "", -1)
		classes = append(classes, name)
		data, err := ioutil.ReadFile("data/ItemClasses/" + classes[i] + ".txt")
		if err != nil {
			fmt.Println(err)
			log.Println(err)
			os.Exit(1)
		}
		AddBaseTypes(data, classes[i])
	}
}

func init() {
	ItemFloatMods = make(map[string]string)
	ItemBoolMods = make(map[string]string)
	data, err := ioutil.ReadFile("data/mods.txt")
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		os.Exit(1)
	}
	dataString := strings.Replace(string(data), "\r", "", -1)
	lines := strings.Split(dataString, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "//") || len(line) == 0 {
			continue //Skip the line
		}
		content := strings.Split(line, "\t\t")
		if len(content) != 2 {
			fmt.Println("Error in data/mods.txt, line ", i+1, ": ", line)
			log.Println("Error in data/mods.txt, line ", i+1, ": ", line)
			os.Exit(1)
		}
		if strings.Contains(content[1], "#") {
			ItemFloatMods[content[1]] = content[0]
		} else {
			ItemBoolMods[content[1]] = content[0]
		}
		//fmt.Println("<tr><td>" + content[0] + "</td><td>" + content[1] + "</td></tr>")
	}
}
