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

var (
	everquestDirectory = flag.String("eqdirectory", "C:\\Users\\thegr\\Desktop\\Project1999", "Everquest Directory")
	characterNames     = flag.String("characters", "Researchchanter:Enchanter,Researchmage:Magician,Researchnecro:Necromancer,Researchwizard:Wizard", "Character inventory files to parse")

	researchPageDB     = []ResearchItem{}
	researchPageDBMap  = map[int]ResearchItem{}
	blacklistItemDB    = []BlacklistItem{}
	blacklistItemDBMap = map[int]BlacklistItem{}
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	flag.Parse()

	log.Println("Inventory Parser")
	log.Printf("Everquest Directory: %s\n", *everquestDirectory)
	log.Printf("Characters: %s\n", *characterNames)

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

	// Read character inventory files
	characterNamesSplit := strings.Split(*characterNames, ",")

	for _, characterNameClass := range characterNamesSplit {

		characterNameClassSplit := strings.Split(characterNameClass, ":")
		if len(characterNameClassSplit) != 2 {
			log.Printf("WARNING: Invalid character name and class tuple: %s\n", characterNameClass)
			continue
		}

		characterPages, err := parseFile(characterNameClassSplit[0], Class(characterNameClassSplit[1]))
		check(err)
		researchPages = append(researchPages, characterPages...)
	}

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

	if len(classPages[Enchanter]) > 0 {
		fmt.Printf("\n==== Enchanter Pages ====\n\n")
		for _, researchPage := range classPages[Enchanter] {
			fmt.Printf("%dx\t%s\n", researchPage.Qty, researchPage.Name)
		}
	}

	if len(classPages[Magician]) > 0 {
		fmt.Printf("\n==== Magician Pages ====\n\n")
		for _, researchPage := range classPages[Magician] {
			fmt.Printf("%dx\t%s\n", researchPage.Qty, researchPage.Name)
		}
	}

	if len(classPages[Necromancer]) > 0 {
		fmt.Printf("\n==== Necromancer Pages ====\n\n")
		for _, researchPage := range classPages[Necromancer] {
			fmt.Printf("%dx\t%s\n", researchPage.Qty, researchPage.Name)
		}
	}

	if len(classPages[Wizard]) > 0 {
		fmt.Printf("\n==== Wizard Pages ====\n\n")
		for _, researchPage := range classPages[Wizard] {
			fmt.Printf("%dx\t%s\n", researchPage.Qty, researchPage.Name)
		}
	}
}

func parseFile(characterName string, class Class) (respitems []ResearchItem, err error) {

	researchItems := map[int]*ResearchItem{}

	readBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s-Inventory.txt", *everquestDirectory, characterName))
	if err != nil {
		return
	}

	lines := strings.Split(string(readBytes), "\n")

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
