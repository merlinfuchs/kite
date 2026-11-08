import { lucideName, templateIcons } from "@/lib/flow/templates";
import { cn } from "@/lib/utils";
import { LucideIcon, PackageIcon } from "lucide-react";
import { createContext, useContext, useEffect, useState } from "react";

// TemplateIconSvgs are the inner SVG markup of the icons of the templates a
// page rendered on the server shows, by name. See templateIconSvgs.
export type TemplateIconSvgs = Record<string, string>;

export const TemplateIconSvgsContext = createContext<TemplateIconSvgs>({});

let lucide: Promise<Record<string, LucideIcon>> | undefined;
let loaded: Record<string, LucideIcon> | undefined;

// loadLucide loads all Lucide icons, once. It's a lot, so only pages in the
// app and the icon picker do.
export function loadLucide() {
  lucide ??= import("lucide-react").then((m) => (loaded = m.icons));
  return lucide;
}

// TemplateIcon shows the icon of a template by its name. The icons of Kite's
// templates are bundled, and pages rendered on the server have the SVGs of
// the others. Elsewhere the first other icon loads all of Lucide. Unknown
// icons are a package.
export default function TemplateIcon({
  name,
  className,
}: {
  name: string;
  className?: string;
}) {
  const svgs = useContext(TemplateIconSvgsContext);
  const svg = svgs[name];
  const Bundled = templateIcons[name];
  const [icons, setIcons] = useState(loaded);
  const needsLucide = !Bundled && svg === undefined && !icons;

  useEffect(() => {
    if (!needsLucide) return;
    let live = true;
    loadLucide().then((icons) => live && setIcons(icons));
    return () => {
      live = false;
    };
  }, [needsLucide]);

  if (Bundled) return <Bundled className={className} />;
  if (svg !== undefined) {
    return (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        className={cn("lucide", className)}
        dangerouslySetInnerHTML={{ __html: svg }}
      />
    );
  }
  if (!icons) return <span className={cn("inline-block", className)} />;
  const Icon = icons[lucideName(name)] ?? PackageIcon;
  return <Icon className={className} />;
}
