# Lastest 
## 2023.4.4
- 测试angular8版本发现，custom-webpack8.0.0版本有问题，用8.1.0版本测试通过


## 2023.4.3
- 测试angular6版本后发现，custom-webpack在8版本之前都没有dev-server，因此angular6-8版本需要额外下载@angular-builders/dev-server，已兼容
- angular11版本前统一用名称匹配；angular11版本后统一用路径匹配；

## 2023.3.31

- 大量测试后做了兼容性配置处理，修改了add-location.js，兼容用正则表达式替换constructor字符串时各种场景

## 2023.3.29

- 更新Chrome新版本删除了鼠标右击的event.path属性，已更新app.component.ts代码，详情见：https://juejin.cn/post/7177645078146449466
 
# Feature
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
- 5、目前低于Angular11版本用名称匹配，存在查找到文件并打开时，触发编译，导致页面刷新，正在解决中...
- 6、需要对app.component.ts做一些处理，会替换里面的ngOnInit方法，如果没有，会提示错误，请手动添加一个ngOnInit空钩子函数。
