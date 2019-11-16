package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)

var (
	everquestDirectory = flag.String("eqdirectory", "C:\\Users\\thegr\\Desktop\\Project1999", "Everquest Directory")

	researchEnchanterName   = "Researchchanter"
	researchMagicianName    = "Researchmage"
	researchNecromancerName = "Researchnecro"
	researchWizardName      = "Researchwizard"

	researchPageDB     = []ResearchItem{}
	researchPageDBMap  = map[int]ResearchItem{}
	blacklistItemDB    = []BlacklistItem{}
	blacklistItemDBMap = map[int]BlacklistItem{}
)

type Class string

var Enchanter Class = "Enchanter"
var Magician Class = "Magician"
var Necromancer Class = "Necromancer"
var Wizard Class = "Wizard"

type ResearchItem struct {
	ID    int    `csv:"id"`
	Name  string `csv:"name"`
	Class Class  `csv:"class"`
	Qty   int
}

type BlacklistItem struct {
	ID   int    `csv:"id"`
	Name string `csv:"name"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	flag.Parse()

	log.Println("Inventory Parser")
	log.Printf("Everquest Directory: %s\n", *everquestDirectory)

	// Load research pages database file
	researchpageDBData, err := ioutil.ReadFile("researchpagedb.txt")
	check(err)
	err = gocsv.UnmarshalBytes(researchpageDBData, &researchPageDB)
	check(err)

	// Build map for faster lookup
	for _, researchPage := range researchPageDB {
		researchPageDBMap[researchPage.ID] = researchPage
	}

	log.Printf("Loaded %d Items from Database\n", len(researchPageDB))

	// Load blacklist database file
	blacklistDBData, err := ioutil.ReadFile("blacklist.txt")
	check(err)
	err = gocsv.UnmarshalBytes(blacklistDBData, &blacklistItemDB)
	check(err)

	// Build map for faster lookup
	for _, blacklistItem := range blacklistItemDB {
		blacklistItemDBMap[blacklistItem.ID] = blacklistItem
	}

	log.Printf("Loaded %d Blacklist Items from Database\n", len(blacklistItemDB))

	// Read in inventory files
	researchPages := []ResearchItem{}

	// Read Enchanter character
	enchanterPages, err := parseFile(researchEnchanterName, Enchanter)
	check(err)
	researchPages = append(researchPages, enchanterPages...)

	// // Read Magician character
	magicianPages, err := parseFile(researchMagicianName, Magician)
	check(err)
	researchPages = append(researchPages, magicianPages...)

	// // Read Necromancer character
	necromancerPages, err := parseFile(researchNecromancerName, Necromancer)
	check(err)
	researchPages = append(researchPages, necromancerPages...)

	// Read Wizard character
	wizardPages, err := parseFile(researchWizardName, Wizard)
	check(err)
	researchPages = append(researchPages, wizardPages...)

	log.Printf("Found %d Research Items\n", len(researchPages))

	// Sort into buckets by class then dump report by class
	sort.Slice(researchPages, func(i, j int) bool {

		if researchPages[i].Class < researchPages[j].Class {
			return true
		}
		if researchPages[i].Class > researchPages[j].Class {
			return false
		}

		return researchPages[i].Name < researchPages[j].Name
	})

	classPages := map[Class][]ResearchItem{}

	for _, researchPage := range researchPages {
		classPages[researchPage.Class] = append(classPages[researchPage.Class], researchPage)
	}

	fmt.Printf("\n==== Enchanter Pages ====\n")
	for _, researchPage := range classPages[Enchanter] {
		if strings.HasPrefix(researchPage.Name, "Part of ") {
			// TODO: Feature: Fix the left/right logic because it doesn't work for even/odd
			if researchPage.ID%2 == 0 {
				fmt.Printf("%dx\t%s (Left)\n", researchPage.Qty, researchPage.Name)
			} else {
				fmt.Printf("%dx\t%s (Right)\n", researchPage.Qty, researchPage.Name)
			}
		} else if strings.Contains(researchPage.Name, "Faded") {
			// TODO: Feature: Translate the ID to whichever page it corresponds to
			fmt.Printf("%dx\t%s (%d)\n", researchPage.Qty, researchPage.Name, researchPage.ID)
		} else {
			fmt.Printf("%dx\t%s\n", researchPage.Qty, researchPage.Name)
		}
	}

	fmt.Printf("\n==== Magician Pages ====\n\n")
	for _, researchPage := range classPages[Magician] {
		fmt.Printf("%dx\t%s\n", researchPage.Qty, researchPage.Name)
	}

	fmt.Printf("\n==== Necromancer Pages ====\n\n")
	for _, researchPage := range classPages[Necromancer] {
		fmt.Printf("%dx\t%s\n", researchPage.Qty, researchPage.Name)
	}

	fmt.Printf("\n==== Wizard Pages ====\n\n")
	for _, researchPage := range classPages[Wizard] {
		fmt.Printf("%dx\t%s\n", researchPage.Qty, researchPage.Name)
	}
}

func parseFile(characterName string, class Class) (respitems []ResearchItem, err error) {

	researchItems := map[int]*ResearchItem{}

	readBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s-Inventory.txt", *everquestDirectory, characterName))
	if err != nil {
		return
	}

	lines := strings.Split(string(readBytes), "\n")
	// fmt.Printf("%s: %d lines\n", characterName, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		lineSplit := strings.Split(line, "\t")
		if len(lineSplit) != 5 {
			continue
		}

		// ex: General1-Slot1	Part of Tasarin's Grimoire Pg. 312	16076	6	5
		slot := lineSplit[0]
		name := lineSplit[1]
		id, _ := strconv.Atoi(lineSplit[2])
		qty, _ := strconv.Atoi(lineSplit[3])

		if !strings.HasPrefix(slot, "General") && !strings.HasPrefix(slot, "Bank") {
			// log.Printf("INFO: Skipping non-inventory slot: %s\n", slot)
			continue
		} else if name == "Empty" {
			// log.Printf("INFO: Skipping empty slot: %s\n", slot)
			continue
		} else if _, ok := blacklistItemDBMap[id]; ok {
			// log.Printf("INFO: Skipping blacklisted item: %s\n", blacklistItem.Name)
			continue
		}

		// Get item from database
		if researchItem, ok := researchPageDBMap[id]; ok {

			if researchItem.Class != class {
				log.Printf("WARNING: %s has wrong class research item: %s [%s]\n", characterName, name, researchItem.Class)
			}

			// Get item from local map if exists
			if item, ok := researchItems[id]; ok {
				item.Qty += qty
				if !strings.HasPrefix(name, "Spell: ") {
					log.Printf("WARNING: Duplicate stack of item: %s on character: %s\n", name, characterName)
				}
			} else {
				researchItem.Qty += qty
				researchItems[id] = &researchItem
			}

		} else {
			log.Printf("ERROR: No item in database for ID: %d\n", id)
		}
	}

	for _, researchItem := range researchItems {
		respitems = append(respitems, *researchItem)
	}

	return
}
