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
// the viewport edges seamlessly with no sharp corners or clipped gradients.
const shapes: Record<
  Variant,
  {
    area: string;
    stroke: string;
    backdropArea?: string;
    backdropStroke?: string;
  }
> = {
  arch: {
    backdropArea:
      "M0,66 C340,66 480,10 720,10 C960,10 1100,66 1440,66 L1440,70 C1100,70 960,20 720,20 C480,20 340,70 0,70 Z",
    backdropStroke: "M0,66 C340,66 480,10 720,10 C960,10 1100,66 1440,66",
    area: "M0,70 C340,70 480,20 720,20 C960,20 1100,70 1440,70 L1440,80 L0,80 Z",
    stroke: "M0,70 C340,70 480,20 720,20 C960,20 1100,70 1440,70",
  },
  horizon: {
    backdropArea:
      "M0,66 C360,66 480,8 720,8 C960,8 1080,66 1440,66 L1440,72 C1080,72 960,22 720,22 C480,22 360,72 0,72 Z",
    backdropStroke: "M0,66 C360,66 480,8 720,8 C960,8 1080,66 1440,66",
    area: "M0,72 C360,72 480,22 720,22 C960,22 1080,72 1440,72 L1440,80 L0,80 Z",
    stroke: "M0,72 C360,72 480,22 720,22 C960,22 1080,72 1440,72",
  },
  wave: {
    area: "M0,44 C180,44 300,64 480,64 C600,64 840,24 960,24 C1140,24 1260,44 1440,44 L1440,80 L0,80 Z",
    stroke:
      "M0,44 C180,44 300,64 480,64 C600,64 840,24 960,24 C1140,24 1260,44 1440,44",
  },
};

const accentColors: Record<Exclude<Accent, "none">, string> = {
  primary: "hsl(var(--primary))",
  discord: "#5865F2",
};

// Layered architectural divider between background and muted homepage bands.
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
  const hairlineId = `divider-hairline-${id}`;
  const rimId = `divider-rim-${id}`;
  const haloId = `divider-halo-${id}`;
  const bandId = `divider-band-${id}`;
  const shape = shapes[variant];

  return (
    <div
      aria-hidden="true"
      className={cn(
        "relative z-10 w-full overflow-hidden select-none pointer-events-none leading-none",
        from === "muted" ? surfaceBg[from] : "bg-transparent",
        className
      )}
    >
      <svg
        viewBox="0 0 1440 80"
        preserveAspectRatio="none"
        className="relative block w-full h-10 sm:h-12 md:h-16"
      >
        <defs>
          <linearGradient id={hairlineId} x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="hsl(var(--border))" stopOpacity="0" />
            <stop
              offset="20%"
              stopColor="hsl(var(--border))"
              stopOpacity="0.45"
            />
            <stop
              offset="50%"
              stopColor="hsl(var(--border))"
              stopOpacity="0.85"
            />
            <stop
              offset="80%"
              stopColor="hsl(var(--border))"
              stopOpacity="0.45"
            />
            <stop
              offset="100%"
              stopColor="hsl(var(--border))"
              stopOpacity="0"
            />
          </linearGradient>

          {accent !== "none" && (
            <>
              <linearGradient id={rimId} x1="0%" y1="0%" x2="100%" y2="0%">
                <stop
                  offset="0%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0"
                />
                <stop
                  offset="18%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.14"
                />
                <stop
                  offset="50%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.72"
                />
                <stop
                  offset="82%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.14"
                />
                <stop
                  offset="100%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0"
                />
              </linearGradient>

              <linearGradient id={haloId} x1="0%" y1="0%" x2="100%" y2="0%">
                <stop
                  offset="0%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0"
                />
                <stop
                  offset="25%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.04"
                />
                <stop
                  offset="50%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.18"
                />
                <stop
                  offset="75%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.04"
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
                  offset="22%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.04"
                />
                <stop
                  offset="50%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.14"
                />
                <stop
                  offset="78%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0.04"
                />
                <stop
                  offset="100%"
                  stopColor={accentColors[accent]}
                  stopOpacity="0"
                />
              </linearGradient>
            </>
          )}
        </defs>

        {shape.backdropArea && (
          <path
            d={shape.backdropArea}
            fill={accent !== "none" ? `url(#${bandId})` : undefined}
            className={accent === "none" ? surfaceFill[to] : undefined}
            fillOpacity={accent === "none" ? "0.45" : undefined}
          />
        )}
        {shape.backdropStroke && accent !== "none" && (
          <path
            d={shape.backdropStroke}
            fill="none"
            stroke={`url(#${rimId})`}
            strokeOpacity="0.28"
            strokeWidth="1"
            vectorEffect="non-scaling-stroke"
          />
        )}

        <path d={shape.area} className="fill-background" />
        {to !== "background" && (
          <path d={shape.area} className={surfaceFill[to]} />
        )}

        <path
          d={shape.stroke}
          fill="none"
          stroke={`url(#${hairlineId})`}
          strokeWidth="1"
          vectorEffect="non-scaling-stroke"
        />

        {accent !== "none" && (
          <>
            <path
              d={shape.stroke}
              fill="none"
              stroke={`url(#${haloId})`}
              strokeWidth="6"
              vectorEffect="non-scaling-stroke"
            />
            <path
              d={shape.stroke}
              fill="none"
              stroke={`url(#${rimId})`}
              strokeWidth="1.5"
              vectorEffect="non-scaling-stroke"
            />
          </>
        )}
      </svg>
    </div>
  );
}
