package main

var AddLocationJSStr string = `const path = require("path");
module.exports = function (source) {
    if (source.indexOf("constructor(") >= 0) {
        const { resourcePath, rootContext } = this;

        /**add-
         * path.relative 根据当前src路径得到源码的相对路径
         */

        const rawShortFilePath = path
            .relative(rootContext || process.cwd(), resourcePath)
            .replace(/^(\.\.[\/\\])+/, "");
        // console.log("rawShortFilePath", rawShortFilePath);
        source = source.replace(
            "constructor",
            "sourcePath ='" +
                rawShortFilePath.replace(/\\/g, "/") +
                "';\n\nconstructor"
        );
    }

    return source;
};
`
var ExtraWebpackConfigJSStr string = `var openInEditor = require("ng-devtools-open-editor-middleware");
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
            app.use("/__open-in-editor", openInEditor("code"));
        },  
        // Webpack5 
        // setupMiddlewares: (middlewares, devServer) => {
        //     middlewares.unshift({
        //         name: "open-editor",
        //         path: "/__open-in-editor",
        //         middleware: openInEditor("code","name"),
        //     });
        //     return middlewares;
        // },       
    },
    module: {
        rules: [
            {
                test: /\.ts$/,
                use: "add-location",
                exclude: [/\.(spec|e2e|service|module)\.ts$/],
            },
        ],
    },
};
`
var AppComStr string = `
openSourceInEditor($event) {    
       $event.preventDefault();
       const path = $event.path;
       let sourcePath = '';
       for (let i = 0; i < path.length; i++) {
           const { localName } = path[i];
           if (localName && localName.indexOf('app-') >= 0) {
               if (path[i].__ngContext__?.component) {
                   sourcePath = path[i].__ngContext__.component.sourcePath;
               } else {
                   const temp = path[i].__ngContext__.find(
                       (e) =>
                           e &&
                           e.sourcePath &&
                           e.sourcePath.indexOf(
                               localName.substring(
                                   localName.indexOf('-') + 1,
                                   localName.length - 1
                               )
                           )
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

var AppComStr2 = ` const listener = new WeakMap();
   listener.set(document.body, this.openSourceInEditor);
   document.body.addEventListener(
       'contextmenu',
       listener.get(document.body),
       false
   );`
