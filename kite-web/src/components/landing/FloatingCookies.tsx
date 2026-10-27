import { cn } from "@/lib/utils";
import { useState } from "react";
import { CookieShape } from "./CookieApiVisuals";

const cookies = [
  { left: "6%", top: "18%", size: 64, duration: 7, delay: 0 },
  { left: "12%", top: "40%", size: 40, duration: 6, delay: 1.5 },
  { left: "84%", top: "12%", size: 52, duration: 8, delay: 0.8 },
  { left: "90%", top: "56%", size: 72, duration: 6.5, delay: 2.2 },
  { left: "4%", top: "72%", size: 34, duration: 5.5, delay: 0.4 },
];

// Cookies floating behind the hero. Clicking one takes a bite, the third
// eats it. Phones have no room for them next to the text.
export default function FloatingCookies() {
  const [bites, setBites] = useState(cookies.map(() => 0));

  return (
    <div className="pointer-events-none absolute inset-0 overflow-hidden">
      {cookies.map((cookie, i) => (
        <button
          key={i}
          tabIndex={-1}
          aria-hidden
          onClick={() =>
            setBites((prev) => prev.map((b, j) => (j === i ? b + 1 : b)))
          }
          className="pointer-events-auto absolute hidden cursor-pointer md:block"
          style={{ left: cookie.left, top: cookie.top }}
        >
          <div
            className="motion-safe:animate-float"
            style={{
              animationDuration: `${cookie.duration}s`,
              animationDelay: `${cookie.delay}s`,
            }}
          >
            <CookieShape
              id={String(i)}
              bites={bites[i]}
              className={cn(
                "drop-shadow-lg transition duration-300 hover:scale-110 hover:rotate-12 active:scale-90",
                bites[i] >= 3 && "scale-0 opacity-0"
              )}
              style={{ width: cookie.size, height: cookie.size }}
            />
          </div>
        </button>
      ))}
    </div>
  );
}
