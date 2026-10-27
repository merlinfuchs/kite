import { cn } from "@/lib/utils";
import { CSSProperties } from "react";

// Mockups of what Cookie API's blocks return, for the Cookie API page.

export function WelcomeCard({
  name,
  color,
  className,
}: {
  name: string;
  color: string;
  // Sizes are in em, so the text-* size of className scales the card.
  className?: string;
}) {
  return (
    <div
      className={cn(
        "relative aspect-[2/1] w-full overflow-hidden rounded-xl text-white select-none",
        className
      )}
      style={{
        background: `linear-gradient(135deg, ${color}, #1e1f22 85%)`,
      }}
    >
      <div
        className="absolute inset-0 opacity-20"
        style={{
          backgroundImage: "radial-gradient(white 1px, transparent 1px)",
          backgroundSize: "14px 14px",
        }}
      />
      <div className="relative h-full flex items-center gap-[6%] px-[7%]">
        <div
          className="flex-none w-[26%] aspect-square rounded-full ring-[0.25em] ring-white/80 flex items-center justify-center font-bold text-[2.4em] shadow-lg"
          style={{ background: color }}
        >
          {(name.trim()[0] ?? "?").toUpperCase()}
        </div>
        <div className="min-w-0">
          <div className="uppercase tracking-[0.3em] text-[0.7em] opacity-80">
            Welcome
          </div>
          <div className="font-bold text-[1.9em] leading-tight truncate">
            {name.trim() || "Member"}
          </div>
          <div className="text-[0.75em] opacity-70 mt-[0.3em]">
            Member #1,204
          </div>
        </div>
      </div>
    </div>
  );
}

// Steve's face, 8 by 8 pixels.
const steve = [
  "hhhhhhhh",
  "hhhhhhhh",
  "hssssssh",
  "ssssssss",
  "swbssbws",
  "sssnnsss",
  "ssmmmmss",
  "ssmsssms",
];
const steveColors: Record<string, string> = {
  h: "#3b2614",
  s: "#c69c6d",
  w: "#ffffff",
  b: "#4a3fb5",
  n: "#8b5a3c",
  m: "#6b3d2a",
};

export function PixelHead({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 8 8"
      className={cn("rounded-sm", className)}
      shapeRendering="crispEdges"
      aria-hidden
    >
      {steve.flatMap((row, y) =>
        [...row].map((c, x) => (
          <rect
            key={`${x}-${y}`}
            x={x}
            y={y}
            width={1}
            height={1}
            fill={steveColors[c]}
          />
        ))
      )}
    </svg>
  );
}

// A made-up QR code with dots, like Cookie API's dots style. It doesn't scan.
const qrSize = 21;
const qrCells = (() => {
  let seed = 42;
  const random = () => {
    seed = (seed * 1103515245 + 12345) % 2147483648;
    return seed / 2147483648;
  };
  const cells: { x: number; y: number; finder: boolean }[] = [];
  const corners = [
    [0, 0],
    [qrSize - 7, 0],
    [0, qrSize - 7],
  ];
  for (let y = 0; y < qrSize; y++) {
    for (let x = 0; x < qrSize; x++) {
      const corner = corners.find(
        ([cx, cy]) => x >= cx - 1 && x <= cx + 7 && y >= cy - 1 && y <= cy + 7
      );
      if (corner) {
        const dx = x - corner[0];
        const dy = y - corner[1];
        const inside = dx >= 0 && dx <= 6 && dy >= 0 && dy <= 6;
        const ring = dx === 0 || dx === 6 || dy === 0 || dy === 6;
        const center = dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4;
        if (inside && (ring || center)) cells.push({ x, y, finder: true });
      } else if (random() < 0.5) {
        cells.push({ x, y, finder: false });
      }
    }
  }
  return cells;
})();

export function QrCode({ className }: { className?: string }) {
  return (
    <svg
      viewBox={`-2 -2 ${qrSize + 4} ${qrSize + 4}`}
      className={cn("bg-white rounded-lg", className)}
      aria-hidden
    >
      {qrCells.map(({ x, y, finder }) =>
        finder ? (
          <rect
            key={`${x}-${y}`}
            x={x}
            y={y}
            width={1.02}
            height={1.02}
            fill="#1e1f22"
          />
        ) : (
          <circle
            key={`${x}-${y}`}
            cx={x + 0.5}
            cy={y + 0.5}
            r={0.42}
            fill="#1e1f22"
          />
        )
      )}
      <rect x={8} y={8} width={5} height={5} rx={1} fill="white" />
      <circle cx={10.5} cy={10.5} r={1.8} fill="#f9ad15" />
    </svg>
  );
}

// A cookie with chocolate chips. Each bite takes a piece off its edge.
const bitePositions = [
  [92, 30],
  [12, 78],
  [80, 94],
];

export function CookieShape({
  bites = 0,
  id,
  className,
  style,
}: {
  bites?: number;
  // Unique on the page, for the mask of the bites.
  id: string;
  className?: string;
  style?: CSSProperties;
}) {
  return (
    <svg viewBox="0 0 100 100" className={className} style={style} aria-hidden>
      <defs>
        <mask id={`bites-${id}`}>
          <rect width="100" height="100" fill="white" />
          {bitePositions.slice(0, bites).map(([x, y]) => (
            <circle key={`${x}-${y}`} cx={x} cy={y} r={17} fill="black" />
          ))}
        </mask>
      </defs>
      <g mask={`url(#bites-${id})`}>
        <circle cx="50" cy="50" r="46" fill="#b9773a" />
        <circle cx="50" cy="50" r="40" fill="#d99a55" />
        <circle cx="34" cy="34" r="6" fill="#4a2c1a" />
        <circle cx="62" cy="28" r="5" fill="#4a2c1a" />
        <circle cx="68" cy="56" r="7" fill="#4a2c1a" />
        <circle cx="40" cy="64" r="5" fill="#4a2c1a" />
        <circle cx="52" cy="46" r="3" fill="#4a2c1a" />
        <circle cx="26" cy="52" r="3.5" fill="#4a2c1a" />
      </g>
    </svg>
  );
}
