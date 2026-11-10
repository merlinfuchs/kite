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
import { NameStyleFont } from "@/lib/nameStyle";

// The fonts of name styles for previews. They're only downloaded once a
// preview uses them. Medieval and Vampyre use Néo-Castel and Sinistre in
// Discord, which aren't on Google Fonts, so similar ones stand in. Next needs
// the options of each written out.

const sakura = Cherry_Bomb_One({
  subsets: ["latin"],
  weight: "400",
  preload: false,
});
const jellybean = Chicle({ subsets: ["latin"], weight: "400", preload: false });
const modern = MuseoModerno({ subsets: ["latin"], preload: false });
const medieval = Uncial_Antiqua({
  subsets: ["latin"],
  weight: "400",
  preload: false,
});
const pixel = Pixelify_Sans({ subsets: ["latin"], preload: false });
const vampyre = Pirata_One({
  subsets: ["latin"],
  weight: "400",
  preload: false,
});
const tempo = Zilla_Slab({ subsets: ["latin"], weight: "600", preload: false });
const monkeyBars = Playpen_Sans({ subsets: ["latin"], preload: false });
const mainframe = Orbitron({ subsets: ["latin"], preload: false });
const headbang = New_Rocker({
  subsets: ["latin"],
  weight: "400",
  preload: false,
});
const journal = Kalam({ subsets: ["latin"], weight: "400", preload: false });

export const nameStyleFontFamilies: Record<NameStyleFont, string | undefined> =
  {
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
