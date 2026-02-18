const config = require('ui/tailwind.config.js')
config.content.push('./node_modules/ui/components/**/*.{js,ts,jsx,tsx}')

const activityColors = require('./app/common/variables').activityColors
activityColors.forEach(color => {
  config.safelist.push('text-' + color + '-900')
  config.safelist.push('bg-' + color + '-200')
  config.safelist.push('bg-' + color + '-300')
})

module.exports = config
