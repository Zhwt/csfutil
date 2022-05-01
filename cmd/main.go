package main

import (
	"fmt"
	"github.com/Zhwt/csfutil"
	"github.com/Zhwt/csfutil/utils"
	"os"
	"strconv"
)

var usageMessages = map[string]string{
	"export": `Usage: csfutil export <filename.csf> [output.xlsx]

Exports all CSF LabelValue items inside the given CSF file into a spreadsheet. The exported file will have the following structure:

	<CSF Label Name>	<CSF Value Value>	<CSF Value ExtraValue>

This file can be imported into a CSF file later.`,
	"import": `Usage: csfutil import <input.xlsx> <filename.csf>

Import all CSF LabelValue items inside the spread sheet into the CSF file. Existing items will be overwritten and missing items will be created. The file must have the following structure:

	<CSF Label Name>	<CSF Value Value>	<CSF Value ExtraValue>

Otherwise the`,
	"merge": `Usage: csfutil merge <source.csf> <destination.csf>

Merges all CSF LabelValue items inside the source file into the destination file. Only existing items will be overwritten and missing items won't' be created.`,
	"new": `Usage: csfutil new <filename.csf> [language code]

Create a empty Version 3 csf file. Valid language code can be 0~9, otherwise it will be recognized as "Unknown".`,
	"help": `csfutil is a tool for manipulating CSF files.

Usage:

	csfutil <command> [arguments]

The commands are:

	export  convert a CSF file to a spreadsheet
	import  merge items from a spreadsheet into a CSF file
	merge   merge one CSF file into another
	new     create empty Version 3 CSF file

Use "csfutil help <command>" for more information about a command.`,
}

func help(cmd string) {
	if !utils.Contains(utils.StringMapKeys(usageMessages), cmd) {
		fmt.Println("Unknown command:", cmd)
		fmt.Println("Run 'csfutil help' for usage.")
	} else {
		fmt.Println(usageMessages[cmd])
	}
}

func printError(err error) {
	fmt.Println("Error:", err)
}

func merge(src, dst string) error {
	srcFile, err := csfutil.Open(src)
	if err != nil {
		return fmt.Errorf("%s: %w", src, err)
	}
	dstFile, err := csfutil.Open(dst)
	if err != nil {
		return fmt.Errorf("%s: %w", dst, err)
	}

	for _, s := range dstFile.Order {
		if v, ok := srcFile.Values[s]; ok {
			dstFile.WriteLabelValue(v, false)
		}
	}

	err = dstFile.Save()
	return err
}

func main() {
	argCount := len(os.Args)
	if argCount < 3 {
		if argCount == 1 {
			fmt.Println("Run 'csfutil help' for usage.")
		} else {
			help(os.Args[1])
		}
		return
	} else {
		switch os.Args[1] {
		case "export":
			if argCount == 3 || argCount == 4 {
				input := os.Args[2]
				output := "output.xlsx"
				if argCount == 4 {
					output = os.Args[3]
				}

				csf, err := csfutil.Open(input)
				if err != nil {
					printError(err)
					return
				}

				err = csfutil.ExportExcel(csf, output)
				if err != nil {
					printError(err)
					return
				}
			} else {
				fmt.Println("Incorrect argument count, want 3 or 4, got", argCount)
				help(os.Args[1])
			}
		case "import":
			if argCount == 4 {
				input := os.Args[2]
				output := os.Args[3]
				err := csfutil.ImportExcel(input, output)
				if err != nil {
					printError(err)
					return
				}
			} else {
				fmt.Println("Incorrect argument count, want 4, got", argCount)
				help(os.Args[1])
			}
		case "merge":
			if argCount == 4 {
				err := merge(os.Args[2], os.Args[3])
				if err != nil {
					printError(err)
					return
				}
			} else {
				fmt.Println("Incorrect argument count, want 4, got", argCount)
				help(os.Args[1])
			}
		case "new":
			if argCount == 3 || argCount == 4 {
				output := os.Args[2]
				language := 0
				var err error
				if argCount == 4 {
					language, err = strconv.Atoi(os.Args[3])
					if err != nil {
						printError(err)
						return
					}
				}

				csf := csfutil.New(output, 3, 0, uint(language))
				if err != nil {
					printError(err)
					return
				}

				if err = csf.Save(); err != nil {
					printError(err)
					return
				}
			} else {
				fmt.Println("Incorrect argument count, want 3 or 4, got", argCount)
				help(os.Args[1])
			}
		case "help":
			if argCount > 2 {
				help(os.Args[2])
			}
		default:
			help(os.Args[1])
		}
	}
}
