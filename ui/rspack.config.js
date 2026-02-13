const path = require('path');

module.exports = {
    entry: './src/main.js',
    output: {
        path: path.resolve(__dirname, '../public/assets'),
        filename: 'main.js',
    },
    module: {
        rules: [
            {
                test: /\.css$/,
                type: 'css',
            },
        ],
    },
    experiments: {
        css: true
    }
};
