package main

var AddLocationJSStr string = `const path = require('path');
module.exports = function (source) {
    if (source.indexOf('constructor(') >= 0) {
        const { resourcePath, rootContext } = this;

        /**add-
         * path.relative 根据当前src路径得到源码的相对路径
         */

        const rawShortFilePath = path.relative(rootContext || process.cwd(), resourcePath).replace(/^(\.\.[\/\\])+/, '');
        const sourcePath = rawShortFilePath.replace(/\\/g, '/');
        if (source.indexOf('super()') >= 0) {
            source = source.replace(/super\(\);/, "super();\nthis.sourcePath='" + sourcePath + "';");
        } else {
            source = source.replace(/constructor\(([\s\S]*?)\)\s*{/gm, "constructor($1){\nthis.sourcePath='" + sourcePath + "';");
        }        
    }

    return source;
};`

var AppComStr2 = ` const listener = new WeakMap();
   listener.set(document.body, this.openSourceInEditor);
   document.body.addEventListener(
       'contextmenu',
       listener.get(document.body),
       false
   );`

func GetAppCompStr() string {
	if AngularVersion >= 11 {
		return `
  openSourceInEditor($event) {    
         $event.preventDefault();
         const path = $event.composedPath();
         let sourcePath = '';
         for (let i = 0; i < path.length; i++) {
             const { localName } = path[i];
             if (localName && localName.indexOf('app-') >= 0) {
                 if (path[i].__ngContext__&& path[i].__ngContext__.component) {
                     sourcePath = path[i].__ngContext__.component.sourcePath;
                 } else {
                     const temp = path[i].__ngContext__.find(
                         (e) =>
                             e &&
                             e.sourcePath &&
                             e.sourcePath.indexOf(localName.substring(localName.indexOf('-') + 1, localName.length)) >= 0
                     );
                     sourcePath = temp.sourcePath;
                 }
                 break;
             }
         }
         fetch('__open-in-editor?file='+sourcePath)
             .then((res) => {})
             .catch((err) => {});
     }`
	} else {
		return `openSourceInEditor($event) {            
            $event.preventDefault();
            const path = $event.composedPath();
            let componentName = '';
            for (let i = 0; i < path.length; i++) {
                const { localName } = path[i];
                if (localName && localName.indexOf('app-') >= 0) {
                    componentName = localName;
                    break;
                }
            }
            fetch('__open-in-editor?file='+sourcePath)
                .then((res) => {})
                .catch((err) => {});
        }`
	}
}

func GetExtraWebpackConfigJSStr() string {
	if AngularVersion >= 11 {
		return `var openInEditor = require("ng-devtools-open-editor-middleware");
        const path = require("path");        
        module.exports = {
            resolveLoader: {
                alias: {
                    "add-location": path.resolve("./add-location.js"),
                },
            },
            devServer: {
                // Webpack5以下
                before(app) {
                    app.use("/__open-in-editor", openInEditor("code","path"));
                },  
                // Webpack5 
                // setupMiddlewares: (middlewares, devServer) => {
                //     middlewares.unshift({
                //         name: "open-editor",
                //         path: "/__open-in-editor",
                //         middleware: openInEditor("code","path"),
                //     });
                //     return middlewares;
                // },       
            },
            module: {
                rules: [
                    {
                        test: /\.component\.ts$/,
                        use: "add-location",
                        exclude: [/\.(spec|e2e|service|module)\.ts$/],
                    },
                ],
            },
        };`
	} else {
		return `var openInEditor = require("ng-devtools-open-editor-middleware");
        const path = require("path");        
        module.exports = {           
            devServer: {                
                before(app) {
                    app.use("/__open-in-editor", openInEditor("code","name"));
                },                      
            },           
        };`
	}
}

func GetCustomWebpackConfig() map[string]interface{} {
	if AngularVersion >= 11 {
		return map[string]interface{}{
			"path": "./extra-webpack.config.js",
			"mergeRules": map[string]interface{}{
				"resolveLoader": "prepend",
				"devServer":     "prepend",
				"module": map[string]interface{}{
					"rules": "prepend",
				}}, "replaceDuplicatePlugins": false}
	} else {
		return map[string]interface{}{
			"path": "./extra-webpack.config.js",
			"mergeStrategies": map[string]interface{}{
				"resolveLoader": "prepend",
				"devServer":     "prepend",
				"module.rules":  "prepend",
			}, "replaceDuplicatePlugins": false}
	}
}
