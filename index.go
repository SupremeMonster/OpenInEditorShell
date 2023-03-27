package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

/*
GO结构体 字段必须用大写
*/
type PackageJSON struct {
	Name            string
	Version         string
	Scripts         map[string]string
	Dependencies    map[string]string
	DevDependencies map[string]string
}

var angularVersion string
var editFileNames = [...]string{"./../package.json", "./../angular.json"}

func main() {
	// 备份文件 错误时回滚
	if backupFiles() != nil {
		fmt.Println("Failed to backup")
		return
	}
	//TODO 监测angular版本  ✔
	//TODO 改angular.json
	//TODO 改package.json   ✔
	//TODO 添加add-location.js  ✔
	//TODO 添加extra-webpack.config.js ✔
	getAngularVersion()
}

func getAngularVersion() {
	data, err := ioutil.ReadFile("./../package.json")
	if err != nil {
		fmt.Println("Failed to read package.json")
		return
	}
	var jsonData PackageJSON
	err = json.Unmarshal((data), &jsonData)
	if err != nil {
		return
	}
	// 截取大版本号
	angularVersion = strings.Split(strings.TrimLeft(jsonData.Dependencies["@angular/core"], "~"), ".")[0]
	fmt.Println("Angular Version：", angularVersion)

	if editPackageJSON(jsonData) != nil || editAngularJSON() != nil || addLocationJS() != nil || addExtraWebpackConfigJS() != nil {
		rollbackFiles()
	} else {
		deleteBackupFiles()
	}

}

func editPackageJSON(jsonData PackageJSON) error {
	dependencies, devDependencies := jsonData.Dependencies, jsonData.DevDependencies
	// 暂时用大版本的包 观察有无问题
	dependencies["@angular-builders/custom-webpack"] = angularVersion + ".0.0"
	devDependencies["ng-devtools-open-editor-middleware"] = "1.0.5"
	newData, _ := json.MarshalIndent(jsonData, "", "    ")
	err := ioutil.WriteFile("./../package.json", newData, 0644)
	if err != nil {
		fmt.Println("Failed to update package.json ", err)
	}
	fmt.Println("Success to update package.json")
	return nil
}

func editAngularJSON() error {
	/**
	TODO 替换@angular-devkit/build-angular为@angular-builders/custom-webpack
	*/
	data, err := ioutil.ReadFile("./../angular.json")
	if err != nil {
		fmt.Println("Failed to read angular.json")
		return err
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	// 可以用第三方库，懒
	options := result["projects"].(map[string]interface{})["clink2-agent"].(map[string]interface{})["architect"].(map[string]interface{})["build"].(map[string]interface{})["options"].(map[string]interface{})
	options["customWebpackConfig"] = map[string]interface{}{
		"path": "./extra-webpack.config.js",
		"mergeRules": map[string]interface{}{
			"resolveLoader": "prepend",
			"devServer":     "prepend",
			"module": map[string]interface{}{
				"rules": "prepend",
			}}, "replaceDuplicatePlugins": false}
	newData, _ := json.MarshalIndent(result, "", "    ")
	newData = []byte(strings.ReplaceAll(string(newData), "@angular-devkit/build-angular", "@angular-builders/custom-webpack"))
	err = ioutil.WriteFile("./../angular.json", newData, 0644)
	if err != nil {
		fmt.Println("Failed to update angular.json ", err)
		return err
	}
	fmt.Println("Success to update angular.json")
	return nil
}
func addLocationJS() error {
	return addFile("add-location.js", AddLocationJSStr)

}
func addExtraWebpackConfigJS() error {
	return addFile("extra-webpack.config.js", ExtraWebpackConfigJSStr)
}

func addFile(filename string, filecontent string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println("Failed to create ", filename, err)
			return err
		}
		defer file.Close()
		_, err = file.WriteString(filecontent)
		if err != nil {
			fmt.Println("Failed to update ", filename, err)
			return err
		}
		fmt.Println("Success to update ", filename)
	} else {
		fmt.Println("Skip to create ", filename)
	}
	return nil
}

// 备份文件
func backupFiles() error {
	for _, fileName := range editFileNames {
		src, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer src.Close()
		dst, err := os.Create(fileName + ".bak")
		if err != nil {
			return err
		}
		defer dst.Close()
		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}
	return nil
}

func rollbackFiles() {
	fmt.Println("Rollbacking...")

	for _, fileName := range editFileNames {
		_ = os.Remove(fileName)
		_ = os.Rename(fileName+".bak", fileName)
	}
	fmt.Println("Success to Rollback")
}

func deleteBackupFiles() {
	for _, fileName := range editFileNames {
		_ = os.Remove(fileName + ".bak")
	}
}
