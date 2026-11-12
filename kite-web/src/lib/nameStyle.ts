import type { CSSProperties } from "react";
import { colorIntToHex } from "@/tools/common/utils/color";

// The fonts and effects of the style of the bot's name in a server, the same
// as NameStyleFonts and NameStyleEffects in kite-service/pkg/provider. The
// labels are the names Discord shows.
export const nameStyleFonts = [
  { value: "default", label: "Default" },
  { value: "sakura", label: "Sakura" },
  { value: "jellybean", label: "Jellybean" },
  { value: "modern", label: "Modern" },
  { value: "medieval", label: "Medieval" },
  { value: "8bit", label: "8Bit" },
  { value: "vampyre", label: "Vampyre" },
  { value: "tempo", label: "Tempo" },
  { value: "monkey_bars", label: "Monkey Bars" },
  { value: "mainframe", label: "Mainframe" },
  { value: "headbang", label: "Headbang" },
  { value: "journal", label: "Journal" },
] as const;

// colors is how many colors Discord's own picker gives the effect. Discord
// takes up to maxNameStyleColors for every effect.
export const nameStyleEffects = [
  { value: "solid", label: "Solid", colors: 1 },
  { value: "gradient", label: "Gradient", colors: 2 },
  { value: "neon", label: "Neon", colors: 1 },
  { value: "toon", label: "Toon", colors: 1 },
  { value: "pop", label: "Pop", colors: 1 },
  { value: "prism", label: "Prism", colors: 5 },
  { value: "gummy", label: "Gummy", colors: 4 },
] as const;

export const maxNameStyleColors = 5;

export type NameStyleFont = (typeof nameStyleFonts)[number]["value"];

// Colors for new name styles and for colors added when the effect changes.
export const defaultNameStyleColors = [
  0x5865f2, 0xeb459e, 0xfee75c, 0x57f287, 0xed4245,
];

// Resizes colors to what the effect uses, keeping the ones set.
export function colorsForEffect<T>(
  effect: string,
  colors: T[],
  fill: (i: number) => T
) {
  return Array.from(
    { length: nameStyleEffects.find((e) => e.value === effect)?.colors ?? 1 },
    (_, i) => colors[i] ?? fill(i)
  );
}

// An approximation of how Discord draws the effect, for previews.
export function nameStyleEffectCss(
  effect: string,
  colors: number[]
): CSSProperties {
  const [first = 0xffffff, second] = colors;
  switch (effect) {
    case "gradient":
    case "prism":
    case "gummy":
      if (colors.length < 2) return { color: colorIntToHex(first) };
      return {
        backgroundImage: `linear-gradient(90deg, ${colors
          .map(colorIntToHex)
          .join(", ")})`,
        backgroundClip: "text",
        WebkitBackgroundClip: "text",
        color: "transparent",
      };
    case "neon":
      return {
        color: "#fff",
        textShadow: `0 0 4px ${colorIntToHex(first)}, 0 0 8px ${colorIntToHex(
          first
        )}`,
      };
    case "toon":
      return {
        color: colorIntToHex(first),
        WebkitTextStroke: "1px rgba(0, 0, 0, 0.6)",
      };
    case "pop":
      return {
        color: colorIntToHex(first),
        textShadow: `2px 2px 0 ${
          second !== undefined ? colorIntToHex(second) : "rgba(0, 0, 0, 0.6)"
        }`,
      };
    default:
      return { color: colorIntToHex(first) };
  }
}
