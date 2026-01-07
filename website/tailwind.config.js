/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        cream: {
          50: '#FDFCFA',
          100: '#F9F6F1',
          200: '#F5F0E8',
          300: '#EDE5D8',
          400: '#E0D4C3',
        },
        coral: {
          400: '#F09B7A',
          500: '#E87A4F',
          600: '#D66A3F',
          700: '#C45A30',
        },
        dark: {
          100: '#4A4A4A',
          200: '#3A3A3A',
          300: '#2A2A2A',
          400: '#1A1A1A',
        }
      },
      fontFamily: {
        display: ['"DM Serif Display"', 'Georgia', 'serif'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      borderRadius: {
        'xl': '1rem',
        '2xl': '1.5rem',
        '3xl': '2rem',
      },
    },
  },
  plugins: [],
}
