/** @type {import('tailwindcss').Config} */
const defaultTheme = require('tailwindcss/defaultTheme');

module.exports = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx}", 
    "./components/**/*.{js,ts,jsx,tsx}",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#6969FF',
        secondary: '#2a282c',
      },
      fontFamily: {
        sans: ['Open Sans', ...defaultTheme.fontFamily.sans],
        serif: ['Merriweather', ...defaultTheme.fontFamily.serif]
      },
    },
  },
  safelist: [
    'bg-primary',
    'bg-secondary',
    'bg-slate-500',
    'bg-red-600',
    'bg-lime-700',
    'bg-neutral-100',
    'bg-neutral-900',
  ],
  plugins: [
    require('@tailwindcss/forms'),
  ],
}
