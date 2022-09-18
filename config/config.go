package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/Unknwon/goconfig"
)

var (
	File *goconfig.ConfigFile
	ROOT string
)

const mainIniPath = "/data/conf/env.ini"

func init() {

	curDir, _ := os.Getwd()
	ROOT = curDir

	configPath := ROOT + mainIniPath

	file, err := goconfig.LoadConfigFile(configPath)
	File = file

	if err != nil {
		fmt.Println("load config file error:", err)
		File, _ = goconfig.LoadFromData([]byte(""))
	}

	if err = loadIncludeFiles(); err != nil {
		panic("load include files error:" + err.Error())
	}

	go signalReload()
}

func ReloadConfigFile() {
	var err error
	configPath := ROOT + mainIniPath
	File, err = goconfig.LoadConfigFile(configPath)
	if err != nil {
		fmt.Println("reload config file, error:", err)
		return
	}

	if err = loadIncludeFiles(); err != nil {
		fmt.Println("reload files include files error:", err)
		return
	}
	fmt.Println("reload config file successfully！")
}

func SaveConfigFile() error {
	err := goconfig.SaveConfigFile(File, ROOT+mainIniPath)
	if err != nil {
		fmt.Println("save config file error:", err)
		return err
	}

	fmt.Println("save config file successfully!")
	return nil
}

func loadIncludeFiles() error {
	includeFile := File.MustValue("include_files", "path", "")
	if includeFile != "" {
		includeFiles := strings.Split(includeFile, ",")
		return File.AppendFiles(includeFiles...)
	}

	return nil
}

// fileExist 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
