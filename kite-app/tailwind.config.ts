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
        primary: "#6375FF",
        dark: {
          base: "#1c1b26",
          "0": "#060609",
          "1": "#13131a",
          "2": "#201f2c",
          "3": "#2d2c3d",
          "4": "#3a384f",
          "5": "#474560",
          "6": "#545172",
          "7": "#615d84",
          "8": "#6e6a95",
          "9": "#7f7ba2",
          "10": "#908dae",
        },
      },
    },
  },
  plugins: [],
};
export default config;
