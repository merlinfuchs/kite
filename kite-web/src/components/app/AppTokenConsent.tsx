import Link from "next/link";

// Shown next to every button that submits a bot token, so users agree to Kite
// acting as their service provider (section 12(a) of the Discord Developer
// Terms) before Kite uses the token.
export default function AppTokenConsent({ action }: { action: string }) {
  return (
    <p className="text-xs text-muted-foreground">
      By {action}, you agree to the{" "}
      <Link href="/terms" target="_blank" className="underline">
        Terms of Service
      </Link>
      , including that Kite runs your bot for you{" "}
      <Link
        href="/terms#service-provider"
        target="_blank"
        className="underline"
      >
        as your service provider
      </Link>{" "}
      under Discord&apos;s Developer Terms.
    </p>
  );
}
