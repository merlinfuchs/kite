import {
  Cherry_Bomb_One,
  Chicle,
  Kalam,
  MuseoModerno,
  New_Rocker,
  Orbitron,
  Pirata_One,
  Pixelify_Sans,
  Playpen_Sans,
  Uncial_Antiqua,
  Zilla_Slab,
} from "next/font/google";
import type { CSSProperties } from "react";
import { nameStyleEffectCss, NameStyleFont } from "@/lib/nameStyle";
import { NameStyle } from "@/lib/types/wire.gen";

// The fonts of name styles for previews. They're only downloaded once a
// preview uses them, in the weight of the name. Medieval and Vampyre use Néo-Castel and Sinistre in
// Discord, which aren't on Google Fonts, so similar ones stand in. Next needs
// the options of each written out.

const sakura = Cherry_Bomb_One({
  subsets: ["latin"],
  weight: "400",
  preload: false,
});
const jellybean = Chicle({ subsets: ["latin"], weight: "400", preload: false });
const modern = MuseoModerno({
  subsets: ["latin"],
  weight: "600",
  preload: false,
});
const medieval = Uncial_Antiqua({
  subsets: ["latin"],
  weight: "400",
  preload: false,
});
const pixel = Pixelify_Sans({
  subsets: ["latin"],
  weight: "600",
  preload: false,
});
const vampyre = Pirata_One({
  subsets: ["latin"],
  weight: "400",
  preload: false,
});
const tempo = Zilla_Slab({ subsets: ["latin"], weight: "600", preload: false });
const monkeyBars = Playpen_Sans({
  subsets: ["latin"],
  weight: "600",
  preload: false,
});
const mainframe = Orbitron({
  subsets: ["latin"],
  weight: "600",
  preload: false,
});
const headbang = New_Rocker({
  subsets: ["latin"],
  weight: "400",
  preload: false,
});
const journal = Kalam({ subsets: ["latin"], weight: "400", preload: false });

const fontFamilies: Record<NameStyleFont, string | undefined> = {
  default: undefined,
  sakura: sakura.style.fontFamily,
  jellybean: jellybean.style.fontFamily,
  modern: modern.style.fontFamily,
  medieval: medieval.style.fontFamily,
  "8bit": pixel.style.fontFamily,
  vampyre: vampyre.style.fontFamily,
  tempo: tempo.style.fontFamily,
  monkey_bars: monkeyBars.style.fontFamily,
  mainframe: mainframe.style.fontFamily,
  headbang: headbang.style.fontFamily,
  journal: journal.style.fontFamily,
};

export function nameStyleFontFamily(font: NameStyleFont) {
  return fontFamilies[font];
}

// How the name looks with the style in previews. Only pages with a name
// style import this, as it adds the CSS of every font.
export function nameStyleCss(style: NameStyle): CSSProperties {
  return {
    fontFamily: nameStyleFontFamily(style.font as NameStyleFont),
    ...nameStyleEffectCss(style.effect, style.colors),
  };
}
