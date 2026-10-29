import type { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { icons } from "lucide-react";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { lucideName, templateIcons } from "./templates";

// templateIconSvgs are the SVGs of the icons that aren't bundled, for pages
// rendered on the server. It loads all Lucide icons, so only use it in
// getStaticProps, where Next leaves it out of the page.
export function templateIconSvgs(
  names: (string | undefined)[]
): TemplateIconSvgs {
  const svgs: TemplateIconSvgs = {};
  for (const name of names) {
    if (!name || name in templateIcons || name in svgs) continue;
    const Icon = icons[lucideName(name) as keyof typeof icons];
    if (!Icon) continue;
    const svg = renderToStaticMarkup(createElement(Icon));
    svgs[name] = svg.slice(svg.indexOf(">") + 1, svg.lastIndexOf("</svg>"));
  }
  return svgs;
}
