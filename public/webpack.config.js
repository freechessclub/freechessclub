var webpack = require('webpack');

module.exports = {
    entry: "./src/index.tsx",
    output: {
        path: __dirname + "/assets/js/",
        filename: "bundle.js"
    },
    resolve: {
        // Add '.ts' and '.tsx' as a resolvable extension.
        extensions: [".webpack.js", ".web.js", ".ts", ".tsx", ".js"]
    },
    module: {
        loaders: [
            // all files with a '.ts' or '.tsx' extension will be handled by 'ts-loader'
            { test: /\.tsx?$/, loader: "ts-loader" },
            { test: /\.css$/, loader: "style!css" }
        ]   
    },
    plugins: [
        new webpack.ProvidePlugin({
         $: "jquery",
         jQuery: "jquery",
         "window.jQuery": "jquery"
        })
    ]
}
