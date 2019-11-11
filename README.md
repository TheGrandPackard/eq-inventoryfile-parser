NOTE: Currently the character names and classes are hardcoded, so this will only work out of the box with the 4 inventory files present in your Everquest directory.

Install golang and exectute like this:

```
go run main.go -eqdirectory="C:\\Project1999"
```

Or build it and run it like this:

```
go build

./eq-inventoryfile-parser -eqdirectory="C:\\Project1999"
```