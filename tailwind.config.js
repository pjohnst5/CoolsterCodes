const colors = require('tailwindcss/colors')
const defaultTheme = require('tailwindcss/defaultTheme')

/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./content/articles/*.md",
        "./content/markdown/**/*.md",
        "./web/html/layouts/**/*.{html,js}",
        "./web/html/pages/**/*.{html,js}",
        "./web/html/_*.ace",
        "./web/html/**/*.{html,js}"
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
                        // Removes backtick before and after inline code
                        'code::before': {
                            content: '""',
                        },
                        'code::after': {
                            content: '""',
                        },
                        '--tw-prose-body': '#fff',       // prose body white
                        '--tw-prose-code': '#c8d1d9',    // inline code
                        '--tw-prose-links': '#5da7d8',   // This was so annoying to figure out
                        '--tw-prose-counters': '#fff',   // Make numbers next to TOC white
                        '--tw-prose-captions': '#adb1ba' // Figure captions
                    },
                },
            },
        }
    },
    plugins: [
        require( '@tailwindcss/typography' )
    ],
}
