import { cn } from "@/lib/utils";
import { useId } from "react";

type Surface = "background" | "muted";
type Variant = "arch" | "horizon" | "wave";
type Accent = "none" | "primary" | "discord";

const surfaceBg: Record<Surface, string> = {
  background: "bg-background",
  muted: "bg-muted/30",
};

const surfaceFill: Record<Surface, string> = {
  background: "fill-background",
  muted: "fill-muted/30",
};

// Every curve uses horizontal tangents at x=0 and x=1440 so transitions meet
// the viewport edges smoothly with no sharp corners or cut-off gradients.
const shapes: Record<
  Variant,
  { area: string; stroke: string; backdropArea?: string }
> = {
  arch: {
    area: "M0,84 C360,84 480,14 720,14 C960,14 1080,84 1440,84 L1440,96 L0,96 Z",
    stroke: "M0,84 C360,84 480,14 720,14 C960,14 1080,84 1440,84",
  },
  horizon: {
    backdropArea:
      "M0,74 C340,74 460,6 720,6 C980,6 1100,74 1440,74 L1440,86 C1100,86 980,22 720,22 C460,22 340,86 0,86 Z",
    area: "M0,86 C340,86 460,22 720,22 C980,22 1100,86 1440,86 L1440,96 L0,96 Z",
    stroke: "M0,86 C340,86 460,22 720,22 C980,22 1100,86 1440,86",
  },
  wave: {
    area: "M0,48 C240,48 360,84 600,84 C840,84 960,12 1200,12 C1320,12 1380,48 1440,48 L1440,96 L0,96 Z",
    stroke:
      "M0,48 C240,48 360,84 600,84 C840,84 960,12 1200,12 C1320,12 1380,48 1440,48",
  },
};

const accentColors: Record<Exclude<Accent, "none">, string> = {
  primary: "hsl(var(--primary))",
  discord: "#5865F2",
};

// Flowing organic transition between default background sections and muted bands.
export default function HomeSectionDivider({
  from = "background",
  to = "muted",
  variant = "arch",
  accent = "none",
  className,
}: {
  from?: Surface;
  to?: Surface;
  variant?: Variant;
  accent?: Accent;
  className?: string;
}) {
  const id = useId().replace(/:/g, "");
  const rimId = `divider-rim-${id}`;
  const bandId = `divider-band-${id}`;
  const shape = shapes[variant];

  return (
    <div
      aria-hidden="true"
      className={cn(
        "relative w-full overflow-hidden select-none pointer-events-none leading-none",
        surfaceBg[from],
        className
      )}
    >
      <svg
        viewBox="0 0 1440 96"
        preserveAspectRatio="none"
        className="relative block w-full h-10 sm:h-14 md:h-20"
      >
        {accent !== "none" && (
          <defs>
            <linearGradient id={rimId} x1="0%" y1="0%" x2="100%" y2="0%">
              <stop
                offset="0%"
                stopColor={accentColors[accent]}
                stopOpacity="0"
              />
              <stop
                offset="50%"
                stopColor={accentColors[accent]}
                stopOpacity="0.48"
              />
              <stop
                offset="100%"
                stopColor={accentColors[accent]}
                stopOpacity="0"
              />
            </linearGradient>
            <linearGradient id={bandId} x1="0%" y1="0%" x2="100%" y2="0%">
              <stop
                offset="0%"
                stopColor={accentColors[accent]}
                stopOpacity="0"
              />
              <stop
                offset="50%"
                stopColor={accentColors[accent]}
                stopOpacity="0.16"
              />
              <stop
                offset="100%"
                stopColor={accentColors[accent]}
                stopOpacity="0"
              />
            </linearGradient>
          </defs>
        )}

        {shape.backdropArea && (
          <path
            d={shape.backdropArea}
            fill={accent !== "none" ? `url(#${bandId})` : undefined}
            className={accent === "none" ? surfaceFill[to] : undefined}
            fillOpacity={accent === "none" ? "0.5" : undefined}
          />
        )}
        <path d={shape.area} className={surfaceFill[to]} />
        {accent !== "none" && (
          <path
            d={shape.stroke}
            fill="none"
            stroke={`url(#${rimId})`}
            strokeWidth="1.5"
            vectorEffect="non-scaling-stroke"
          />
        )}
      </svg>
    </div>
  );
}
