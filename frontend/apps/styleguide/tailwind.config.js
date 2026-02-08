const config = require('ui/tailwind.config.js')
config.content.push('./node_modules/ui/components/**/*.{js,ts,jsx,tsx}')
config.content.push('./examples/**/*.{js,ts,jsx,tsx}')

module.exports = config
