import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useUser } from "@/lib/hooks/api";
import { formatLegalDate, legalUpdatedAt } from "@/lib/legal";
import { useLegalUpdateStore } from "@/lib/state/legalUpdate";
import Link from "next/link";
import { useEffect, useState } from "react";

// How long after the update users who haven't seen it yet are told about it.
const showForMs = 60 * 24 * 60 * 60 * 1000;

// Tells users who signed up before the last change to the terms or privacy
// policy about it, once per browser.
export default function LegalUpdateDialog() {
  const user = useUser();
  const dismissedAt = useLegalUpdateStore((s) => s.dismissedAt);
  const dismiss = useLegalUpdateStore((s) => s.dismiss);

  // The store is read from localStorage, which the prerendered page can't
  // know about, so wait for mount to avoid a hydration mismatch.
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);

  const updatedAt = legalUpdatedAt.getTime();
  const open =
    mounted &&
    !!user &&
    new Date(user.created_at).getTime() < updatedAt &&
    dismissedAt < updatedAt &&
    Date.now() < updatedAt + showForMs;

  return (
    <Dialog open={open} onOpenChange={(open) => !open && dismiss(updatedAt)}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>We updated our terms</DialogTitle>
          <DialogDescription>
            Our{" "}
            <Link href="/terms" target="_blank" className="underline">
              Terms of Service
            </Link>{" "}
            and{" "}
            <Link href="/privacy" target="_blank" className="underline">
              Privacy Policy
            </Link>{" "}
            were rewritten on {formatLegalDate(legalUpdatedAt)} and apply from
            then on.
          </DialogDescription>
        </DialogHeader>
        <ul className="list-disc space-y-1.5 pl-5 text-sm">
          <li>
            Kite runs your bot as your service provider under Discord&apos;s
            Developer Terms, and only for you.
          </li>
          <li>
            They include a data processing agreement for the data your bot
            handles.
          </li>
          <li>
            They set the rules for templates you share and list in the
            marketplace.
          </li>
          <li>
            The privacy policy names every service we share data with, like
            OpenAI for the AI features.
          </li>
        </ul>
        <DialogFooter>
          <Button onClick={() => dismiss(updatedAt)}>Got it</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
