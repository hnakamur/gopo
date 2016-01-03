package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robfig/gettext-go/gettext/po"
)

const usage = `gopo is a tool for managing gettext *.po files.

Usage:
    gopo <command> <srcDir> <destDir>

Commands:
    cp          copy msgstrs
    orphans     show orphans msgstrs
`

func walkPoFiles(root string, walkFn filepath.WalkFunc) error {
	myWalkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && filepath.Ext(info.Name()) == ".po" {
			return walkFn(path, info, err)
		}
		return nil
	}
	return filepath.Walk(root, myWalkFn)
}

func buildMsgIdToFilePathMaps(root string) (map[string]string, error) {
	maps := make(map[string]string)
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		poFile, err := po.Load(path)
		if err != nil {
			return err
		}

		for _, msg := range poFile.Messages {
			maps[msg.MsgId] = path
		}

		return nil
	}
	err := walkPoFiles(root, walkFn)
	if err != nil {
		return nil, err
	}
	return maps, nil
}

func updatePoFileWithMessage(path string, msg po.Message) error {
	poFile, err := po.Load(path)
	if err != nil {
		return err
	}

	for i, m := range poFile.Messages {
		if m.MsgId == msg.MsgId {
			poFile.Messages[i].MsgStr = msg.MsgStr
			err := poFile.Save(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func cpCommand(args []string) error {
	srcDir := args[0]
	destDir := args[1]
	fmt.Printf("command:gopo cp\tsrcDir:%s\tdestDir:%s\n", srcDir, destDir)
	msgIdToFilePathMaps, err := buildMsgIdToFilePathMaps(destDir)
	if err != nil {
		return err
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		poFile, err := po.Load(path)
		if err != nil {
			return err
		}

		for _, msg := range poFile.Messages {
			if msg.MsgStr != "" {
				destPath, ok := msgIdToFilePathMaps[msg.MsgId]
				if ok {
					fmt.Printf("srcPath:%s\tdestPath:%s\tmsgId:%s\tmsgStr:%s\n", path, destPath, msg.MsgId, msg.MsgStr)
					err := updatePoFileWithMessage(destPath, msg)
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	}
	return walkPoFiles(srcDir, walkFn)
}

func orphansCommand(args []string) error {
	srcDir := args[0]
	destDir := args[1]
	fmt.Printf("command:gopo orphans\tsrcDir:%s\tdestDir:%s\n", srcDir, destDir)
	msgIdToFilePathMaps, err := buildMsgIdToFilePathMaps(destDir)
	if err != nil {
		return err
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		poFile, err := po.Load(path)
		if err != nil {
			return err
		}

		for _, msg := range poFile.Messages {
			if msg.MsgStr != "" {
				_, ok := msgIdToFilePathMaps[msg.MsgId]
				if !ok {
					fmt.Printf("srcPath:%s\tmsgId:%s\tmsgStr:%s\n", path, msg.MsgId, msg.MsgStr)
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	}
	return walkPoFiles(srcDir, walkFn)
}

func main() {
	flag.Usage = func() {
		fmt.Print(usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) != 3 {
		flag.Usage()
		os.Exit(1)
	}
	switch args[0] {
	case "cp":
		err := cpCommand(args[1:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "orphans":
		err := orphansCommand(args[1:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
}
