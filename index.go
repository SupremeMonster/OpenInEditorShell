package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

// Angular11以下采用名称匹配 Angular11及以上采用路径匹配
//TODO 输入project名 ✔
//TODO 检测angular大版本  ✔
//TODO 改angular.json   ✔
//TODO 改package.json   ✔
//TODO 添加add-location.js  ✔
//TODO 添加extra-webpack.config.js ✔
//TODO 修改app.component.ts ✔
//TODO 错误时回滚 ✔
//TODO 询问yarn/npm install ✔
//TODO 按q退出 ✔

type packageJSON struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Scripts         map[string]interface{} `json:"scripts"`
	Private         bool                   `json:"private"`
	Dependencies    map[string]interface{} `json:"dependencies"`
	DevDependencies map[string]interface{} `json:"devDependencies"`
}

var projectName string
var AngularVersion int
var packageJsonData packageJSON
var filePathMap = map[string]string{
	"package": "./package.json",
	"angular": "./angular.json",
	"app":     "./src/app/app.component.ts",
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

	// 执行步骤
	steps := []struct {
		name string
		fn   func() error
	}{
		{"get angular version", getAngularVersion},
		{"update package.json", editPackageJSON},
		{"update angular.json", editAngularJSON},
		{"update app.component.ts", editAppComTS},
		{"add location.js", addLocationJS},
		{"add extra-webpack.config.js", addExtraWebpackConfigJS},
		{"delete backup files", func() error { return deleteFiles(filePathMap, "backup") }},
		{"package install", runInstallTask},
	}

	for _, step := range steps {
		if err := step.fn(); err != nil {
			// 输出错误信息和步骤名称
			fmt.Printf("Error in %s: %s\n", step.name, err)
			rollbackFiles()
			return
		}
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
	fmt.Print("please input your project name，you can get it from 'defaultProject' property in angular.json.\n：")
	scanner.Scan()
	projectName = scanner.Text()
	fmt.Printf("Project Name: %v\n", projectName)
	return nil
}

func getAngularVersion() error {
	data, err := os.ReadFile(filePathMap["package"])
	if err != nil {
		return fmt.Errorf("failed to read package.json: %v", err)
	}
	if err = json.Unmarshal(data, &packageJsonData); err != nil {
		return fmt.Errorf("failed to parse package.json: %v", err)
	}
	coreVersion := packageJsonData.Dependencies["@angular/core"].(string)
	// 需要考虑存在 ~ ^情况
	coreVersion = strings.TrimLeft(strings.TrimLeft(coreVersion, "~"), "^")
	// 截取大版本号
	AngularVersion, _ = strconv.Atoi(strings.Split(coreVersion, ".")[0])
	fmt.Println("angular version：", AngularVersion)
	if AngularVersion < 6 {
		return fmt.Errorf("angular version is too old，please update to at least 6")
	}
	return nil
}

func editPackageJSON() error {
	// 暂时用大版本的包 观察有无问题 8版本之前custom-webpack没有dev-server
	if AngularVersion < 8 {
		packageJsonData.Dependencies["@angular-builders/dev-server"] = strconv.Itoa(AngularVersion) + ".0.0"
	}
	// 8.0.0版本的custom webpack有问题 dist目录都没有 怀疑作者忘记打包
	if AngularVersion == 8 {
		packageJsonData.Dependencies["@angular-builders/custom-webpack"] = strconv.Itoa(AngularVersion) + "8.1.0"
	} else {
		packageJsonData.Dependencies["@angular-builders/custom-webpack"] = strconv.Itoa(AngularVersion) + ".0.0"
	}
	packageJsonData.DevDependencies["ng-devtools-open-editor-middleware"] = "1.0.6"
	newData, _ := json.MarshalIndent(packageJsonData, "", "    ")
	if err := os.WriteFile(filePathMap["package"], newData, 0644); err != nil {
		return fmt.Errorf("failed to update package.json: %v", err)
	}
	fmt.Println("success to update package.json")

	return nil
}

func editAngularJSON() error {
	/**
	  替换@angular-devkit/build-angular为@angular-builders/custom-webpack
	*/
	data, err := os.ReadFile(filePathMap["angular"])
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
	options["customWebpackConfig"] = GetCustomWebpackConfig()

	newData, _ := json.MarshalIndent(angularjsonData, "", "    ")

	// browser|server|karma 替换  version>=8 才有dev-server version<8需要安装@angular-builders/dev-server
	newData = bytes.ReplaceAll(newData, []byte("@angular-devkit/build-angular:browser"), []byte("@angular-builders/custom-webpack:browser"))
	if AngularVersion < 8 {
		newData = bytes.ReplaceAll(newData, []byte("@angular-devkit/build-angular:dev-server"), []byte(`@angular-builders/dev-server:generic`))
	} else {
		newData = bytes.ReplaceAll(newData, []byte("@angular-devkit/build-angular:dev-server"), []byte(`@angular-builders/custom-webpack:dev-server`))
	}
	newData = bytes.ReplaceAll(newData, []byte("@angular-devkit/build-angular:karma"), []byte("@angular-builders/custom-webpack:karma"))
	if err = os.WriteFile(filePathMap["angular"], newData, 0644); err != nil {
		return fmt.Errorf("failed to update angular.json: %v", err)
	}
	fmt.Println("success to update angular.json")
	return nil
}

/*
根据ngOnInit替换 如果没有会报错
*/
func editAppComTS() error {
	data, err := os.ReadFile(filePathMap["app"])
	if err != nil {
		return fmt.Errorf("failed to read app.component.ts: %v", err)
	}
	// 检测如果已经替换过了 则跳过
	if strings.Contains(string(data), "openSourceInEditor") {
		fmt.Println("skip to update app.component.ts")
		return nil
	}
	// 在ngOnInit前添加一个方法
	newData := bytes.Replace(data, []byte("ngOnInit"), []byte(GetAppCompStr()+"\nngOnInit"), 1)
	// 在ngOnInit中添加代码 找到ngOnInit中后一个"{"
	ngOnInitIndex := strings.Index(string(newData), "ngOnInit")
	if ngOnInitIndex == -1 {
		return fmt.Errorf("failed to match ngOnInit hook function, please add an empty ngOnInit hook function manually")
	}
	leftBraceIndex := strings.Index(string(newData)[ngOnInitIndex:], "{")
	leftBraceIndex += ngOnInitIndex
	newContent := string(newData)[:leftBraceIndex+1] + "\n" + AppComStr2 + "\n" + string(newData)[leftBraceIndex+1:]
	if err = os.WriteFile(filePathMap["app"], []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update app.component.ts: %v", err)
	}
	return nil
}

func addLocationJS() error {
	if AngularVersion < 11 {
		return nil
	}
	return addFile(addFileMap["addLocation"], AddLocationJSStr)

}
func addExtraWebpackConfigJS() error {
	return addFile(addFileMap["webpack"], GetExtraWebpackConfigJSStr())
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
		fmt.Println("success to update", fileName)
	} else {
		fmt.Println("skip to create", fileName)
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

// 让客户选择是执行npm install 还是yarn install，上下选择，怎么写
func runInstallTask() error {
	// 选择npm还是yarn
	bower := ""
	prompt := promptui.Select{
		Label: "Select bower",
		Items: []string{"npm", "yarn"},
	}
	_, bower, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("failed to select bower: %v", err)
	}
	// 执行install
	if err := runPackageInstall(bower); err != nil {
		return err
	}

	return nil
}

// 执行 install
func runPackageInstall(bower string) error {
	cmd := exec.Command(bower, "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run %s install: %v", bower, err)
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
