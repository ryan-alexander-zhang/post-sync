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
        background: "#f6f1e8",
        foreground: "#1d1b19",
        card: "#fffaf2",
        border: "#d4c7b5",
        primary: "#125b50",
        accent: "#f4b860",
        muted: "#ece2d3",
      },
    },
  },
  plugins: [],
};

export default config;
