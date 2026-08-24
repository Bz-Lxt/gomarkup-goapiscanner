/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        ink: '#070B14',
        panel: '#101826',
        line: '#1C2A3F',
        phosphor: '#3EE0C5',
        crit: '#E84855',
        high: '#E88C30',
        mid: '#E6C443',
        low: '#7AA2C8',
        mute: '#8AA0B5',
      },
      fontFamily: {
        display: ['Oxanium', 'sans-serif'],
        sans: ['IBM Plex Sans', 'sans-serif'],
        mono: ['IBM Plex Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}
