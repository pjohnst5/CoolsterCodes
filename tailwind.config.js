const colors = require('tailwindcss/colors')
const defaultTheme = require('tailwindcss/defaultTheme')

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./content/articles/*.md",
        "./content/markdown/**/*.md",
        "./html/layouts/**/*.{html,js}",
        "./html/pages/**/*.{html,js}",
        "./html/_*.ace",
        "./html/**/*.{html,js}"
    ],
    darkMode: 'selector',
    theme: {
        extend: {
            colors: {
                myblue: '#5da7d8',
            },
            fontFamily: {
            },
            typography: {
                DEFAULT: {
                    css: {
                        blockquote: {
                            // Disables the quotes around blockquotes that
                            // Tailwind includes by default. They look decent,
                            // but turn into a real mess if you do things like
                            // cite a source (tick appears after the source's
                            // name) or include a list.
                            quotes: "none",
                        },
                        '--tw-prose-links': '#5da7d8', // This was so annoying to figure out
                    },
                },
            },
        }
    },
    plugins: [
        require( '@tailwindcss/typography' )
    ],
}
