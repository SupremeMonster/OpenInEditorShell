# OpenInEditorShell
angular右击组件在vscode中直接打开源码，工程所需调整的脚本，用go实现；



功能链接地址 ：[https://github.com/SupremeMonster/ng-devtools-openInEditor]

## 使用方法：
- 1、已经编译好的.exe放到angular工程中，和package.json同级；
- 2、运行openInEditor.exe，自动执行文件处理，任一错误都会回滚，不会影响源项目；
- 3、成功后重启项目，右键即可使用

## 注意事项:
- 1、目前只支持Windows
- 2、文件处理后，会自动执行npm install，依实际情况可自行安装
- 3、针对理想化的angular尚未开启自定义webpack配置做的文件处理，如果项目已经配置了webpack，需要自行检查处理结果。
- 4、生产环境需要关闭
