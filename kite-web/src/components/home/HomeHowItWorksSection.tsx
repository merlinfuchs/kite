import ShowcaseVisual from "@/components/landing/ShowcaseVisual";
import { banFlow } from "@/components/landing/flows";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { ArrowRightIcon, CheckIcon } from "lucide-react";
import Link from "next/link";
import { memo, useState } from "react";
import { StatusMockup } from "./ProfileMockups";
import { usePinnedScrollProgress } from "./usePinnedScrollProgress";

const banShowcase = {
  flow: banFlow,
  command: "ban",
  reply: "Banned the user.",
};

const StepThreeShowcase = memo(function StepThreeShowcase() {
  return <ShowcaseVisual showcase={banShowcase} />;
});

const intents = [
  { label: "Presence Intent", at: 0.14 },
  { label: "Server Members Intent", at: 0.38 },
  { label: "Message Content Intent", at: 0.6 },
];

const steps = [
  {
    title: "Create a Discord app",
    description:
      "Create an application in the Discord Developer Portal and copy its bot token. It takes two minutes.",
  },
  {
    title: "Add it to Kite",
    description:
      "Sign in with Discord, paste the token, and invite the bot to your server. It's online right away.",
  },
  {
    title: "Build your commands",
    description:
      "Connect blocks in the editor, add a template from the Marketplace, or describe a command to Flow AI. Save it and it shows up in Discord.",
  },
];

function clamp01(n: number) {
  return Math.min(1, Math.max(0, n));
}

function DevPortalStage({ progress }: { progress: number }) {
  const tokenReady = progress >= 0.78;

  return (
    <div className="rounded-xl border bg-[#313338] text-white p-6 space-y-6 shadow-sm">
      <div>
        <div className="text-xs font-bold uppercase text-gray-400 mb-2">
          Token
        </div>
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex-auto rounded bg-[#1e1f22] px-3 py-2 text-sm truncate transition-colors duration-300 border",
              tokenReady
                ? "border-[#23a55a]/60 text-gray-200"
                : "border-transparent text-gray-500"
            )}
          >
            {tokenReady
              ? "MTI4NzQ5 ••••••••••••••••••••••••"
              : "Click Reset Token to generate"}
          </div>
          <div
            className={cn(
              "flex-none whitespace-nowrap rounded px-4 py-2 text-sm font-medium transition-colors duration-300 flex items-center gap-1.5",
              tokenReady ? "bg-[#23a55a] text-white" : "bg-[#5865F2] text-white"
            )}
          >
            {tokenReady ? (
              <>
                <CheckIcon className="h-4 w-4" />
                Copied
              </>
            ) : (
              "Reset Token"
            )}
          </div>
        </div>
      </div>

      <div>
        <div className="text-xs font-bold uppercase text-gray-400 mb-3">
          Privileged Gateway Intents
        </div>
        <div className="space-y-3">
          {intents.map((intent) => {
            const on = progress >= intent.at;
            return (
              <div
                key={intent.label}
                className="flex items-center justify-between"
              >
                <div className="text-sm">{intent.label}</div>
                <div
                  className={cn(
                    "w-10 h-6 rounded-full flex items-center px-1 transition-colors duration-300",
                    on ? "bg-[#23a55a]" : "bg-[#4e5058]"
                  )}
                >
                  <div
                    className={cn(
                      "w-4 h-4 rounded-full bg-white transition-transform duration-300",
                      on ? "translate-x-4" : "translate-x-0"
                    )}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

function AddAppStage({ progress }: { progress: number }) {
  const tokenPasted = progress >= 0.25;
  const created = progress >= 0.65;

  return (
    <div className="space-y-4">
      <div className="rounded-xl border bg-card p-6 space-y-4 shadow-sm">
        <div>
          <div className="text-lg font-semibold">Create app</div>
          <div className="text-sm text-muted-foreground">
            Kite runs your bot for you. Paste your Discord bot token below to
            bring it online.
          </div>
        </div>
        <div className="space-y-2">
          <Label>Bot token</Label>
          <Input
            type="password"
            value={tokenPasted ? "placeholder-token-value" : ""}
            placeholder="Paste your Discord bot token"
            readOnly
            className="transition-colors duration-300"
          />
        </div>
        <div className="flex justify-end">
          <Button
            tabIndex={-1}
            className={cn(
              "transition-all duration-300",
              created && "bg-emerald-600 hover:bg-emerald-600 text-white"
            )}
          >
            {created ? (
              <>
                <CheckIcon className="mr-1.5 h-4 w-4" />
                Bot online
              </>
            ) : (
              "Create app"
            )}
          </Button>
        </div>
      </div>

      <div
        className={cn(
          "transition-all duration-300",
          created
            ? "opacity-100 translate-y-0"
            : "opacity-40 translate-y-1 pointer-events-none"
        )}
      >
        <StatusMockup
          statuses={[
            created ? "🎮 Playing with 1,204 members" : "Connecting...",
          ]}
        />
      </div>
    </div>
  );
}

export default function HomeHowItWorksSection() {
  const {
    sectionRef,
    stageRef,
    progress: scrollProgress,
    pinState,
    isDesktop,
    reducedMotion,
    scrollToProgress,
  } = usePinnedScrollProgress();
  const [mobileStep, setMobileStep] = useState(0);

  const stepProgresses = [
    clamp01(scrollProgress / 0.34),
    clamp01((scrollProgress - 0.34) / 0.34),
    clamp01((scrollProgress - 0.68) / 0.32),
  ];
  const scrollActiveStep =
    scrollProgress < 0.34 ? 0 : scrollProgress < 0.68 ? 1 : 2;
  const activeStep = isDesktop ? scrollActiveStep : mobileStep;

  function jumpToStep(index: number) {
    setMobileStep(index);
    scrollToProgress(index === 0 ? 0.2 : index === 1 ? 0.56 : 0.92);
  }

  return (
    <section
      id="how-it-works"
      ref={sectionRef}
      className={cn(
        "relative bg-muted/30 scroll-mt-16",
        !reducedMotion && "lg:h-[280vh]"
      )}
    >
      <div
        ref={stageRef}
        className={cn(
          "w-full py-16 sm:py-24 lg:py-10 lg:min-h-[calc(100vh-3.5rem)] lg:flex lg:items-center",
          !reducedMotion &&
            pinState === "before" &&
            "lg:absolute lg:top-0 lg:inset-x-0",
          !reducedMotion &&
            pinState === "pinned" &&
            "lg:fixed lg:top-14 lg:inset-x-0 lg:z-20",
          !reducedMotion &&
            pinState === "after" &&
            "lg:absolute lg:bottom-0 lg:inset-x-0"
        )}
      >
        <div className="container">
          <div className="max-w-5xl mx-auto">
            <h2 className="text-3xl md:text-4xl font-bold text-center">
              Online in three steps
            </h2>

            {!reducedMotion && (
              <div className="hidden lg:flex items-center gap-4 max-w-md mx-auto mt-4">
                <span className="text-xs font-medium text-muted-foreground tabular-nums w-20">
                  Step {scrollActiveStep + 1} of 3
                </span>
                <div className="h-1.5 flex-1 rounded-full bg-border overflow-hidden">
                  <div
                    className="h-full w-full rounded-full bg-primary origin-left transition-transform duration-150 ease-out"
                    style={{ transform: `scaleX(${scrollProgress})` }}
                  />
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-center mt-8">
              {/* Step list with scroll-driven per-step progress tracks */}
              <ol className="lg:col-span-5 space-y-5">
                {steps.map((step, i) => {
                  const isActive = activeStep === i;
                  const fill = reducedMotion ? 1 : stepProgresses[i];

                  return (
                    <li key={step.title}>
                      <button
                        type="button"
                        onClick={() => jumpToStep(i)}
                        className={cn(
                          "w-full text-left relative pl-6 py-2 transition-opacity duration-300 cursor-pointer",
                          isActive
                            ? "opacity-100"
                            : "opacity-45 hover:opacity-75"
                        )}
                      >
                        <span
                          aria-hidden="true"
                          className="absolute left-0 top-2 bottom-2 w-0.5 rounded-full bg-border"
                        >
                          <span
                            className="hidden lg:block h-full w-full origin-top rounded-full bg-primary transition-transform duration-150 ease-out"
                            style={{ transform: `scaleY(${fill})` }}
                          />
                          <span
                            className={cn(
                              "block lg:hidden h-full w-full rounded-full bg-primary transition-opacity",
                              isActive ? "opacity-100" : "opacity-0"
                            )}
                          />
                        </span>

                        <div className="flex items-center gap-3">
                          <span
                            className={cn(
                              "flex h-7 w-7 flex-none items-center justify-center rounded-full text-sm font-bold transition-colors",
                              isActive
                                ? "bg-primary text-primary-foreground"
                                : "bg-muted text-muted-foreground"
                            )}
                          >
                            {i + 1}
                          </span>
                          <h3 className="text-xl font-semibold">
                            {step.title}
                          </h3>
                        </div>
                        <p className="text-muted-foreground mt-2 pl-10">
                          {step.description}
                        </p>
                      </button>
                    </li>
                  );
                })}
              </ol>

              <div className="lg:col-span-7">
                <div className="relative h-[440px]">
                  {[0, 1, 2].map((stageIndex) => {
                    const isStageActive = activeStep === stageIndex;
                    const stageFill = !isDesktop
                      ? isStageActive
                        ? 1
                        : 0
                      : reducedMotion
                      ? 1
                      : stepProgresses[stageIndex];

                    return (
                      <div
                        key={stageIndex}
                        className={cn(
                          "absolute inset-0 flex items-center justify-center transition-opacity duration-300 ease-out",
                          isStageActive
                            ? "opacity-100 z-10 pointer-events-auto"
                            : "opacity-0 z-0 pointer-events-none"
                        )}
                      >
                        <div className="w-full">
                          {stageIndex === 0 && (
                            <DevPortalStage progress={stageFill} />
                          )}
                          {stageIndex === 1 && (
                            <AddAppStage progress={stageFill} />
                          )}
                          {stageIndex === 2 && (
                            <StepThreeShowcase
                              key={
                                isDesktop
                                  ? "desktop"
                                  : isStageActive
                                  ? "mobile-active"
                                  : "mobile-idle"
                              }
                            />
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>

            <div className="text-center mt-8">
              <Link
                href="/guides/how-to-make-a-discord-bot"
                className="inline-flex items-center gap-2 text-primary font-medium hover:underline"
              >
                Read the step-by-step guide
                <ArrowRightIcon className="w-4 h-4" />
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
