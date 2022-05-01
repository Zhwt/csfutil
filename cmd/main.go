package main

import (
	"fmt"
	"github.com/Zhwt/csfutil"
	"github.com/Zhwt/csfutil/utils"
	"os"
)

var usageMessages = map[string]string{
	"export": `Usage: csfutil export <filename.csf> [output.xlsx]

Exports all CSF LabelValue items inside the given CSF file into a spreadsheet. The exported file will have the following structure:

	<CSF Label Name>	<CSF Value Value>	<CSF Value ExtraValue>

This file can be imported into a CSF file later.`,
	"import": `Usage: csfutil import <filename.csf> <input.xlsx>

Import all CSF LabelValue items inside the spread sheet into the CSF file. Existing items will be overwritten and missing items will be created.`,
	"merge": `Usage: csfutil merge <source.csf> <destination.csf>

Merges all CSF LabelValue items inside the source file into the destination file. Only existing items will be overwritten and missing items won't' be created.`,
	"new": `Usage: csfutil new <filename.csf>

Create a empty csf file.`,
	"help": `csfutil is a tool for manipulating CSF files.

Usage:

	csfutil <command> [arguments]

The commands are:

	export  convert a CSF file to a spreadsheet
	import  merge items from a spreadsheet into a CSF file
	merge   merge one CSF file into another
	new     create empty CSF file

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

func merge(src, dst string) error {
	srcFile, err := csfutil.Open(src)
	if err != nil {
		return fmt.Errorf("%s: %w", src, err)
	}
	dstFile, err := csfutil.Open(dst)
	if err != nil {
		return fmt.Errorf("%s: %w", src, err)
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
			fmt.Println("not implemented")
		case "import":
			fmt.Println("not implemented")
		case "merge":
			if argCount == 4 {
				err := merge(os.Args[2], os.Args[3])
				if err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println("Incorrect argument count, want 4, got", argCount)
				help("merge")
			}
		case "new":
			fmt.Println("not implemented")
		case "help":
			if argCount > 2 {
				help(os.Args[2])
			}
		default:
			help(os.Args[1])
		}
	}
}
