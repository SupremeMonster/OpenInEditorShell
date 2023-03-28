package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//TODO 需要输入project名 ✔
//TODO 检测angular版本  ✔
//TODO 改angular.json   ✔
//TODO 改package.json   ✔
//TODO 添加add-location.js  ✔
//TODO 添加extra-webpack.config.js ✔
//TODO 修改app.component.ts ✔
//TODO npm install ✔
//TODO 询问退出 ✔
/*
GO结构体 字段必须用大写  json tag 指定反序列化时的字段名
*/

type packageJSON struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Scripts         map[string]interface{} `json:"scripts"`
	Private         bool                   `json:"private"`
	Dependencies    map[string]interface{} `json:"dependencies"`
	DevDependencies map[string]interface{} `json:"devDependencies"`
}

var projectName string
var angularVersion string
var packageJsonData packageJSON
var filePathMap = map[string]string{
	"package": "./package.json",
	"angular": "./angular.json",
	"app":     "./../src/app/app.component.ts",
}
var addFileMap = map[string]string{
	"addLocation": "add-location.js",
	"webpack":     "extra-webpack.config.js",
}

func main() {
	if err := getProjectName(); err != nil {
		fmt.Println(err.Error())
		return
	}
	// 备份文件 错误时回滚
	if err := backupFiles(); err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := getAngularVersion(); err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := editPackageJSON(); err != nil {
		fmt.Println(err.Error())
		rollbackFiles()
		return
	}

	if err := editAngularJSON(); err != nil {
		fmt.Println(err.Error())
		rollbackFiles()
		return
	}
	if err := editAppComTS(); err != nil {
		fmt.Println(err.Error())
		rollbackFiles()
		return
	}
	if err := addLocationJS(); err != nil {
		fmt.Println(err.Error())
		rollbackFiles()
		return
	}
	if err := addExtraWebpackConfigJS(); err != nil {
		fmt.Println(err.Error())
		rollbackFiles()
		return
	}
	if err := deleteFiles(filePathMap, "backup"); err != nil {
		fmt.Println(err.Error())
	}
	if err := runNpmInstallTask(); err != nil {
		fmt.Println(err.Error())
	}
	PressQToExit()

}

func getProjectName() error {
	cmd := exec.Command("cmd")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start cmd:%v", err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("please input your project name，you can get it from 'defaultProject' property in angular.json!\n：")
	scanner.Scan()
	projectName = scanner.Text()
	fmt.Printf("Project Name: %v\n", projectName)
	return nil
}

func getAngularVersion() error {
	data, err := ioutil.ReadFile(filePathMap["package"])
	if err != nil {
		return fmt.Errorf("failed to read package.json: %v", err)
	}
	if err = json.Unmarshal(data, &packageJsonData); err != nil {
		return fmt.Errorf("failed to parse package.json: %v", err)
	}
	coreVersion := packageJsonData.Dependencies["@angular/core"].(string)
	// 截取大版本号
	angularVersion = strings.Split(strings.TrimLeft(coreVersion, "~"), ".")[0]
	fmt.Println("angular Version：", angularVersion)
	if versionNum, _ := strconv.Atoi(angularVersion); versionNum < 6 {
		return fmt.Errorf("angular version is too old，please update to at least 6")
	}
	return nil
}

func editPackageJSON() error {
	// 暂时用大版本的包 观察有无问题
	packageJsonData.Dependencies["@angular-builders/custom-webpack"] = angularVersion + ".0.0"
	packageJsonData.DevDependencies["ng-devtools-open-editor-middleware"] = "1.0.5"
	newData, _ := json.MarshalIndent(packageJsonData, "", "    ")
	if err := ioutil.WriteFile(filePathMap["package"], newData, 0644); err != nil {
		return fmt.Errorf("failed to update package.json: %v", err)
	}
	fmt.Println("success to update package.json")
	return nil
}

func editAngularJSON() error {
	/**
	  替换@angular-devkit/build-angular为@angular-builders/custom-webpack
	*/
	data, err := ioutil.ReadFile(filePathMap["angular"])
	if err != nil {
		return fmt.Errorf("failed to read angular.json: %v", err)
	}
	var angularjsonData map[string]interface{}
	if err = json.Unmarshal(data, &angularjsonData); err != nil {
		return fmt.Errorf("failed to parse angular.json: %v", err)
	}
	// 可以用第三方库，懒
	project, ok := angularjsonData["projects"].(map[string]interface{})[projectName]
	if !ok {
		return fmt.Errorf("please check your input projectName")
	}
	options, ok := project.(map[string]interface{})["architect"].(map[string]interface{})["build"].(map[string]interface{})["options"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to get build options in angular.json: %v", err)
	}
	options["customWebpackConfig"] = map[string]interface{}{
		"path": "./extra-webpack.config.js",
		"mergeRules": map[string]interface{}{
			"resolveLoader": "prepend",
			"devServer":     "prepend",
			"module": map[string]interface{}{
				"rules": "prepend",
			}}, "replaceDuplicatePlugins": false}
	newData, _ := json.MarshalIndent(angularjsonData, "", "    ")
	newData = bytes.ReplaceAll(newData, []byte("@angular-devkit/build-angular"), []byte("@angular-builders/custom-webpack"))
	if err = ioutil.WriteFile(filePathMap["angular"], newData, 0644); err != nil {
		return fmt.Errorf("failed to update angular.json: %v", err)
	}
	fmt.Println("success to update angular.json")
	return nil
}

/*
根据ngOnInit替换 如果没有会报错
*/
func editAppComTS() error {
	data, err := ioutil.ReadFile(filePathMap["app"])
	if err != nil {
		return fmt.Errorf("failed to read app.component.ts: %v", err)
	}
	// 在ngOnInit前添加一个方法
	newData := bytes.Replace(data, []byte("ngOnInit"), []byte(AppComStr+"\nngOnInit"), 1)
	// 在ngOnInit中添加代码 找到ngOnInit中后一个"{"
	ngOnInitIndex := strings.Index(string(newData), "ngOnInit")
	if ngOnInitIndex == -1 {
		return fmt.Errorf("no match found for ngOnInit")
	}
	leftBraceIndex := strings.Index(string(newData)[ngOnInitIndex:], "{")
	leftBraceIndex += ngOnInitIndex
	newContent := string(newData)[:leftBraceIndex+1] + "\n" + AppComStr2 + "\n" + string(newData)[leftBraceIndex+1:]
	if err = ioutil.WriteFile(filePathMap["app"], []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update app.component.ts: %v", err)
	}
	return nil
}

func addLocationJS() error {
	return addFile(addFileMap["addLocation"], AddLocationJSStr)

}
func addExtraWebpackConfigJS() error {
	return addFile(addFileMap["webpack"], ExtraWebpackConfigJSStr)
}

func addFile(fileName string, fileContent string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create file  %s: %v", fileName, err)
		}
		defer file.Close()
		_, err = file.WriteString(fileContent)
		if err != nil {
			return fmt.Errorf("failed to update file  %s: %v", fileName, err)
		}
		fmt.Println("success to update ", fileName)
	} else {
		fmt.Println("skip to create ", fileName)
	}
	return nil
}

// 备份文件
func backupFiles() error {
	for _, filePath := range filePathMap {
		src, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", filePath, err)
		}
		defer src.Close()
		dst, err := os.Create(filePath + ".bak")
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", filePath+".bak", err)
		}
		defer dst.Close()
		_, err = io.Copy(dst, src)
		if err != nil {
			return fmt.Errorf("failed to copy file %s: %v", filePath+".bak", err)
		}
	}
	return nil
}

func rollbackFiles() error {
	// 回滚源文件
	for _, filePath := range filePathMap {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to delete file %s: %v", filePath, err)
		}
		if err := os.Rename(filePath+".bak", filePath); err != nil {
			return fmt.Errorf("failed to rename file %s: %v", filePath+".bak", err)
		}
	}

	// 删除生成文件和备份脚本
	if err := deleteFiles(addFileMap, "backup"); err != nil {
		return err
	}
	if err := deleteFiles(addFileMap, "js"); err == nil {
		return err
	}
	fmt.Println("success to Rollback")
	return nil
}

// 删除文件  备份文件/添加的js脚本
func deleteFiles(fileMap map[string]string, deleteType string) error {
	fileName := ""
	for _, filePath := range fileMap {
		if deleteType == "backup" {
			fileName = filePath + ".bak"
		} else {
			fileName = "./" + fileName
		}
		_, err := os.Stat(fileName)
		if !os.IsNotExist(err) {
			if err = os.Remove(fileName); err != nil {
				return fmt.Errorf("failed to delete file %s: %v", fileName, err)
			}
		}
	}
	return nil
}

// 执行npm安装
func runNpmInstallTask() error {
	fmt.Println("installing packages，waiting ......")
	cmd := exec.Command("cmd", "/c", "yarn install")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	fmt.Println(out.String())
	if err != nil {
		return fmt.Errorf("failed to install packages，please try it yourself again：%v", err)
	}
	return nil
}
func PressQToExit() {
	fmt.Print("press q to exit :")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ticker.C:
			fmt.Println("press q to exit :")
		default:
			if scanner.Scan() {
				text := scanner.Text()
				if text == "q" {
					return
				}
			}
		}
	}

}
