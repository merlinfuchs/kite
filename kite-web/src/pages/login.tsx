import BaseLayout from "@/components/common/BaseLayout";
import ThemeSwitch from "@/components/common/ThemeSwitch";
import DiscordIcon from "@/components/icons/DiscordIcon";
import { Button } from "@/components/ui/button";
import logo from "@/assets/logo/orange@1024.png";
import { loginUrl, loginUrlTo } from "@/lib/api/client";
import {
  ArrowLeftIcon,
  BlocksIcon,
  CloudIcon,
  Loader2Icon,
  SparklesIcon,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useMemo, useState } from "react";

const highlights = [
  { icon: BlocksIcon, label: "Visual flow editor, no coding required" },
  { icon: CloudIcon, label: "Your bot stays online 24/7" },
  { icon: SparklesIcon, label: "Free to get started" },
];

// Only paths on this site, so /login?redirect= can't send anyone elsewhere.
// The API checks the redirect again after the login.
function safeRedirect(value: unknown) {
  if (typeof value !== "string") return null;
  if (!value.startsWith("/") || value.startsWith("//")) return null;
  if (value.includes("\\")) return null;
  return value;
}

export default function LoginPage() {
  const router = useRouter();
  const [redirecting, setRedirecting] = useState(false);

  const href = useMemo(() => {
    const redirect = safeRedirect(router.query.redirect);
    return redirect ? loginUrlTo(redirect) : loginUrl;
  }, [router.query.redirect]);

  useEffect(() => {
    // Coming back from Discord with the back button restores this page from
    // the cache, where it shouldn't look like it's still redirecting.
    function onPageShow(e: PageTransitionEvent) {
      if (e.persisted) setRedirecting(false);
    }

    window.addEventListener("pageshow", onPageShow);
    return () => window.removeEventListener("pageshow", onPageShow);
  }, []);

  return (
    <BaseLayout title="Log in" noindex>
      <div className="relative flex min-h-[100dvh] w-full flex-col overflow-hidden bg-background">
        {/* Backdrop */}
        <div
          aria-hidden
          className="pointer-events-none absolute left-1/2 top-1/2 h-[520px] w-[520px] -translate-x-1/2 -translate-y-[60%] rounded-full bg-primary/20 blur-[120px]"
        />

        <header className="relative z-10 flex h-14 items-center justify-between px-4 sm:px-8">
          <Link
            href="/"
            className="flex items-center gap-2 rounded-md font-semibold outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <img src={logo.src} className="h-6 w-6" alt="" />
            Kite
          </Link>
          <ThemeSwitch className="m-0 h-9 w-9 text-muted-foreground hover:text-foreground [&_svg]:h-[18px] [&_svg]:w-[18px]" />
        </header>

        <main className="relative z-10 flex flex-1 items-center justify-center px-5 pb-16 pt-6">
          <div className="w-full max-w-sm">
            <div className="rounded-2xl border bg-card/80 p-7 shadow-xl shadow-black/5 backdrop-blur-sm dark:shadow-black/40 sm:p-8">
              <div className="mb-6 flex flex-col items-center text-center">
                <div className="mb-5 flex h-14 w-14 items-center justify-center rounded-2xl border bg-background shadow-sm">
                  <img src={logo.src} className="h-8 w-8" alt="Kite" />
                </div>
                <h1 className="text-2xl font-semibold tracking-tight">
                  Welcome to Kite
                </h1>
                <p className="mt-1.5 text-sm text-muted-foreground">
                  Log in with your Discord account to build and manage your
                  bots.
                </p>
              </div>

              <Button
                asChild
                size="lg"
                className="h-11 w-full gap-2.5 bg-[#5865F2] text-white hover:bg-[#4752C4] focus-visible:ring-[#5865F2]"
              >
                <a href={href} onClick={() => setRedirecting(true)}>
                  {redirecting ? (
                    <Loader2Icon className="h-[18px] w-[18px] animate-spin" />
                  ) : (
                    <DiscordIcon className="h-[18px] w-[18px]" />
                  )}
                  {redirecting
                    ? "Redirecting to Discord…"
                    : "Continue with Discord"}
                </a>
              </Button>

              <div className="my-6 h-px bg-border" />

              <ul className="space-y-2.5">
                {highlights.map(({ icon: Icon, label }) => (
                  <li
                    key={label}
                    className="flex items-center gap-3 text-sm text-muted-foreground"
                  >
                    <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
                      <Icon className="h-4 w-4" />
                    </span>
                    {label}
                  </li>
                ))}
              </ul>
            </div>

            <p className="mt-6 text-center text-xs leading-relaxed text-muted-foreground">
              By continuing, you agree to Kite&apos;s{" "}
              <Link
                href="/terms"
                className="underline underline-offset-2 hover:text-foreground"
              >
                Terms of Service
              </Link>{" "}
              and{" "}
              <Link
                href="/privacy"
                className="underline underline-offset-2 hover:text-foreground"
              >
                Privacy Policy
              </Link>
              .
            </p>

            <div className="mt-4 flex justify-center">
              <Button
                variant="ghost"
                size="sm"
                className="text-muted-foreground"
                asChild
              >
                <Link href="/">
                  <ArrowLeftIcon className="mr-1.5 h-4 w-4" />
                  Back to home
                </Link>
              </Button>
            </div>
          </div>
        </main>
      </div>
    </BaseLayout>
  );
}
