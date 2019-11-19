Configurable eqdirectory and characters. Charaters is a comma separated list of colon separated character names and classes.

Install golang and exectute like this:

```
go run main.go -eqdirectory="C:\\Project1999" -characters "Researchchanter:Enchanter,Researchmage:Magician,Researchnecro:Necromancer,Researchwizard:Wizard"
```

Or build it and run it like this:

```
go build

./eq-inventoryfile-parser.exe -eqdirectory="C:\\Project1999"-characters "Researchchanter:Enchanter,Researchmage:Magician,Researchnecro:Necromancer,Researchwizard:Wizard"
```