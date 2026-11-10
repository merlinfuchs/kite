import type { CSSProperties } from "react";

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
export type NameStyleEffect = (typeof nameStyleEffects)[number]["value"];

// Colors for new name styles and for colors added when the effect changes.
export const defaultNameStyleColors = [
  0x5865f2, 0xeb459e, 0xfee75c, 0x57f287, 0xed4245,
];

export function nameStyleEffectColors(effect: string) {
  return nameStyleEffects.find((e) => e.value === effect)?.colors ?? 1;
}

// Resizes colors to what the effect uses, keeping the ones set.
export function colorsForEffect<T>(
  effect: string,
  colors: T[],
  fill: (i: number) => T
) {
  return Array.from(
    { length: nameStyleEffectColors(effect) },
    (_, i) => colors[i] ?? fill(i)
  );
}

function hex(color: number) {
  return "#" + color.toString(16).padStart(6, "0");
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
      if (colors.length < 2) return { color: hex(first) };
      return {
        backgroundImage: `linear-gradient(90deg, ${colors
          .map(hex)
          .join(", ")})`,
        backgroundClip: "text",
        WebkitBackgroundClip: "text",
        color: "transparent",
      };
    case "neon":
      return {
        color: "#fff",
        textShadow: `0 0 4px ${hex(first)}, 0 0 8px ${hex(first)}`,
      };
    case "toon":
      return {
        color: hex(first),
        WebkitTextStroke: "1px rgba(0, 0, 0, 0.6)",
      };
    case "pop":
      return {
        color: hex(first),
        textShadow: `2px 2px 0 ${
          second !== undefined ? hex(second) : "rgba(0, 0, 0, 0.6)"
        }`,
      };
    default:
      return { color: hex(first) };
  }
}
