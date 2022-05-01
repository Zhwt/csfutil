# csfutil

A tool for easy accessing Red Alert 2 and Yuri's revenge CSF (string table) files.

## Usage

### Command-line tool:

Grab pre-compiled binary from [release page](https://github.com/Zhwt/csfutil/releases) or use `go install github.com/Zhwt/csfutil/cmd`.

Provides basic operations such as:
* Export to spreadsheet.
* Import from spreadsheet.
* Merge two files into one.

Run 'csfutil help' for usage.

### As Go library:

Install with `go get`:

```shell
go get -u -v github.com/Zhwt/csfutil
```

Usage:

```go
package main

import (
	"fmt"
	"github.com/Zhwt/csfutil"
)

func main() {
	// Open file
	u := csfutil.MustOpen("ra2md.csf")
    
	// Print file header information
	fmt.Println(u.Version)
	fmt.Println(u.NumLabels)
	fmt.Println(u.NumStrings)
	fmt.Println(u.Unused)
	fmt.Println(u.LanguageName())
    
	// Print single item
	// Prints "VOX:ceva001 -> Warning: Nuclear Silo detected. , ceva001c"
	fmt.Println(u.Values["VOX:CEVA001"])
	
	// Print Label and Value part
	// Prints "VOX:ceva001"
	fmt.Println(u.Values["VOX:CEVA001"].Label.ValueString()) 
	// Prints "Warning: Nuclear Silo detected."
	fmt.Println(u.Values["VOX:CEVA001"].Value.ValueString())
	// Prints "ceva001c"
	fmt.Println(u.Values["VOX:CEVA001"].Value.ExtraValueString())
    
	// Item order and categories are stored in Order and Categories field
	// respectively.
	for _, name := range u.Order {
		fmt.Println(name)
	}
	for name, strings := range u.Categories {
		fmt.Println(name, strings)
	}
	
	// Write or overwrite single item, false is to preserve original label when overwrite
	u.WriteLabelValue(csfutil.NewLabelValue("TXT:Greeting", "Hello!"), false)
	
	// Craft LabelValue programmatically
	lbl := csfutil.Label{}
	lbl.Write("TXT:Hello")

	// Do not write to Value.Value field directly, due to it needs to perform
	// extra transform to store string value.
	val := csfutil.Value{}
	val.Write("Hi!")
	val.WriteExtra("Woohoo!")

	lv := csfutil.LabelValue{
		Label: lbl,
		Value: val,
	}

	u.WriteLabelValue(lv, false)
	
	// Sync changes to local file.
	err := u.Save()
	if err != nil {
		panic(err)
    }
}
```

## Motivation

Last year I found a Yuri's Revenge mod called [Rise of the East](https://riseoftheeastmod.com/) which is awesome. And at that point it had no Chinese support, but it said it will add Chinese translation in the upcoming releases. But a few days ago I found out the translation work has been suspended due to the previous translator is missing. So I started my own version of translation. But the CSF file this mod uses was using lower case label name, which makes merging existing translated CSF file into this one very hard, so I made this tool.

## License

MIT License.
