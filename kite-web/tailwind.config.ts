import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: "class",
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      backgroundImage: {
        "gradient-radial": "radial-gradient(var(--tw-gradient-stops))",
        "gradient-conic":
          "conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))",
      },
      colors: {
        primary: "#F97316",
        "primary-dark": "",
        dark: {
          base: "#1c1b26",
          "0": "#060505",
          "1": "#0c0b0a",
          "2": "#12110f",
          "3": "#181614",
          "4": "#1e1c19",
          "5": "#24211e",
          "6": "#2a2723",
          "7": "#302c28",
          "8": "#36322d",
          "9": "#3c3732",
          "10": "#504b47",
        },
      },
      scale: {
        101: "1.01",
      },
      saturate: {
        80: "0.8",
      },
    },
  },
  plugins: [],
};
export default config;
