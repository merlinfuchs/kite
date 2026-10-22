import { CheckIcon, SparklesIcon } from "lucide-react";

const points = [
  "Click on Ask AI in the editor and describe what the command should do.",
  "Kite's AI adds and connects the blocks. You check the changes and save.",
  "Your bot can use AI too: answer questions and search the web, without an API key.",
];

export default function HomeAISection() {
  return (
    <section id="ai" className="scroll-mt-16 container py-16 sm:py-24">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
        <div>
          <h2 className="text-3xl md:text-4xl font-bold">
            Describe It,{" "}
            <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
              Kite Builds It
            </span>
          </h2>
          <p className="text-xl text-muted-foreground mt-4">
            Not sure which blocks you need? Tell Kite&apos;s AI what you want in
            your own words.
          </p>
          <ul className="mt-6 space-y-3">
            {points.map((point) => (
              <li key={point} className="flex gap-3">
                <CheckIcon className="flex-none w-5 h-5 mt-0.5 text-primary" />
                <span>{point}</span>
              </li>
            ))}
          </ul>
        </div>

        {/* A simplified version of the Ask AI panel in the editor. */}
        <div className="rounded-xl border bg-card shadow-sm p-5 space-y-4">
          <div className="flex items-center gap-2 font-semibold">
            <SparklesIcon className="w-5 h-5 text-primary" />
            Ask AI
          </div>
          <div className="flex justify-end">
            <div className="rounded-lg bg-primary text-primary-foreground px-4 py-2 max-w-[85%]">
              Make a /purge command that deletes the last messages and says how
              many were deleted.
            </div>
          </div>
          <div className="rounded-lg bg-muted px-4 py-2 max-w-[85%]">
            I added a number argument for how many messages to delete, a Bulk
            delete messages block and a reply with the count. Check the
            highlighted blocks and save when you&apos;re happy.
          </div>
          <div className="flex gap-2">
            <div className="rounded-md bg-primary text-primary-foreground text-sm px-3 py-1.5">
              Save Changes
            </div>
            <div className="rounded-md border text-sm px-3 py-1.5">Undo</div>
          </div>
        </div>
      </div>
    </section>
  );
}
