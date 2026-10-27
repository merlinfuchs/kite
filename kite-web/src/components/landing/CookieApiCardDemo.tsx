import { cn } from "@/lib/utils";
import { MouseEvent, useRef, useState } from "react";
import { Input } from "../ui/input";
import { WelcomeCard } from "./CookieApiVisuals";

const colors = ["#f97316", "#5865f2", "#10b981", "#ec4899"];

// A welcome card that fills in the name typed in, like Generate card fills in
// placeholders. It tilts towards the pointer.
export default function CookieApiCardDemo() {
  const [name, setName] = useState("Sam");
  const [color, setColor] = useState(colors[0]);
  const cardRef = useRef<HTMLDivElement>(null);

  function tilt(e: MouseEvent<HTMLDivElement>) {
    const el = cardRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width - 0.5;
    const y = (e.clientY - rect.top) / rect.height - 0.5;
    el.style.transform = `perspective(800px) rotateX(${-y * 12}deg) rotateY(${
      x * 16
    }deg) scale(1.02)`;
  }

  function untilt() {
    if (cardRef.current) cardRef.current.style.transform = "";
  }

  return (
    <div className="grid grid-cols-1 items-center gap-10 lg:grid-cols-2">
      <div>
        <h2 className="text-3xl md:text-4xl font-bold">
          Every member gets{" "}
          <span className="inline-block -rotate-2 rounded-lg bg-[#f9ad15] px-2 text-black">
            their own
          </span>{" "}
          card
        </h2>
        <p className="text-lg text-muted-foreground mt-4">
          Design the card in Cookie API&apos;s Card Builder and paste it into
          the Generate card block. Placeholders fill in each member&apos;s name
          and avatar. Try it:
        </p>
        <div className="mt-6 flex flex-col gap-4 sm:flex-row sm:items-center">
          <Input
            value={name}
            maxLength={20}
            onChange={(e) => setName(e.target.value)}
            placeholder="Your name"
            aria-label="Your name"
            className="sm:w-56"
          />
          <div className="flex gap-2">
            {colors.map((c) => (
              <button
                key={c}
                onClick={() => setColor(c)}
                aria-label={`Card color ${c}`}
                className={cn(
                  "h-8 w-8 rounded-full transition-transform hover:scale-110",
                  c === color &&
                    "ring-2 ring-foreground ring-offset-2 ring-offset-background"
                )}
                style={{ background: c }}
              />
            ))}
          </div>
        </div>
        <pre className="mt-6 overflow-x-auto rounded-lg border bg-muted/50 p-4 text-sm">
          <code>
            {'{ "type": "text", "text": "'}
            <span className="rounded bg-primary/20 px-1 text-primary">
              {"{{ user.display_name }}"}
            </span>
            {'" }'}
          </code>
        </pre>
      </div>

      <div onMouseMove={tilt} onMouseLeave={untilt} className="p-4">
        <div
          ref={cardRef}
          className="transition-transform duration-200 ease-out will-change-transform"
        >
          <WelcomeCard
            name={name}
            color={color}
            className="text-sm sm:text-lg shadow-2xl"
          />
        </div>
      </div>
    </div>
  );
}
