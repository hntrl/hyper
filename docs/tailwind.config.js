const { fontFamily } = require('tailwindcss/defaultTheme')

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.html"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter", ...fontFamily.sans]
      }
    },
  },
  variants: {
    typography: ["dark"]
  },
  plugins: [
    require("@tailwindcss/typography")
  ],
}
