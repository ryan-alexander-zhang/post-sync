import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "#04111f",
        foreground: "#eff7ff",
        card: "rgba(9,22,38,0.78)",
        border: "rgba(128,166,198,0.18)",
        primary: "#39d98a",
        accent: "#7dd3fc",
        muted: "rgba(93,122,149,0.18)",
      },
      fontFamily: {
        sans: ["var(--font-sans)", "sans-serif"],
        mono: ["var(--font-mono)", "monospace"],
      },
    },
  },
  plugins: [],
};

export default config;
