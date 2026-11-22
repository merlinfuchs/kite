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
const showForMs = 30 * 24 * 60 * 60 * 1000;

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
            were updated on {formatLegalDate(legalUpdatedAt)} and apply from
            then on.
          </DialogDescription>
        </DialogHeader>
        <ul className="list-disc space-y-1.5 pl-5 text-sm">
          <li>
            They cover Quick setup, where apps run on the Kite bot instead of a
            bot of your own.
          </li>
          <li>
            The privacy policy says what the Kite bot does with the data of
            servers and their members.
          </li>
          <li>
            The privacy policy also covers badges, API tokens and AI apps you
            connect to Kite.
          </li>
        </ul>
        <DialogFooter>
          <Button onClick={() => dismiss(updatedAt)}>Got it</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
