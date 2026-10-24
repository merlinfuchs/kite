import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import env from "@/lib/env/client";
import { cn } from "@/lib/utils";
import {
  BookOpenIcon,
  ExternalLinkIcon,
  HelpCircleIcon,
  XIcon,
} from "lucide-react";
import { useEffect, useState } from "react";

const PERMANENT_STORAGE_KEY = "kite_flow_help_dismissed";
const SESSION_STORAGE_KEY = "kite_flow_help_session_dismissed";

interface Props {
  className?: string;
}

export default function FlowHelpPopup({ className }: Props) {
  const [open, setOpen] = useState(false);
  const [doNotShowAgain, setDoNotShowAgain] = useState(false);

  useEffect(() => {
    try {
      const permanent =
        localStorage.getItem(PERMANENT_STORAGE_KEY) === "true" ||
        document.cookie
          .split("; ")
          .some((c) => c.startsWith(`${PERMANENT_STORAGE_KEY}=true`));
      if (permanent) return;

      const session = sessionStorage.getItem(SESSION_STORAGE_KEY) === "true";
      if (session) return;

      setOpen(true);
    } catch {
      setOpen(true);
    }
  }, []);

  const handleDismissSession = () => {
    try {
      sessionStorage.setItem(SESSION_STORAGE_KEY, "true");
    } catch {}
    setOpen(false);
  };

  const handleDismissPermanent = () => {
    try {
      localStorage.setItem(PERMANENT_STORAGE_KEY, "true");
      document.cookie = `${PERMANENT_STORAGE_KEY}=true; path=/; max-age=31536000; SameSite=Lax`;
    } catch {}
    setOpen(false);
  };

  const handleOkClick = () => {
    if (doNotShowAgain) {
      handleDismissPermanent();
    } else {
      handleDismissSession();
    }
  };

  if (!open) return null;

  const docsUrl = `${env.NEXT_PUBLIC_DOCS_LINK}/guides/flow-editor`;

  return (
    <Card
      className={cn(
        "fixed bottom-16 inset-x-3 sm:inset-x-auto sm:absolute sm:bottom-4 sm:left-3 sm:w-96 z-30 shadow-lg border border-border dark:border-stone-800 bg-[#FCFCFC] dark:bg-[#141211] text-foreground dark:text-stone-100 animate-in fade-in slide-in-from-bottom-2 duration-200",
        className
      )}
    >
      <CardHeader className="p-4 pb-2 flex flex-row items-start justify-between space-y-0 gap-2">
        <div className="flex items-center gap-2">
          <div className="size-8 rounded-md bg-primary/15 text-primary flex items-center justify-center flex-none">
            <HelpCircleIcon className="size-4" />
          </div>
          <div>
            <CardTitle className="text-base font-semibold text-foreground dark:text-stone-100">
              Need Help?
            </CardTitle>
          </div>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="size-7 rounded-md text-muted-foreground hover:text-foreground hover:bg-stone-200/60 dark:text-stone-400 dark:hover:text-stone-100 dark:hover:bg-stone-800/60 -mr-1 -mt-1 flex-none"
          onClick={handleDismissSession}
          title="Close until page reopens"
        >
          <XIcon className="size-4" />
          <span className="sr-only">Close</span>
        </Button>
      </CardHeader>

      <CardContent className="p-4 pt-2 pb-3 space-y-2.5">
        <p className="text-xs text-muted-foreground dark:text-stone-400 leading-relaxed">
          Learn how to use the editor and learn about its features.
        </p>

        <a
          href={docsUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center justify-between p-2.5 rounded-lg border border-border dark:border-stone-800 bg-white hover:bg-stone-100/80 dark:bg-stone-900/60 dark:hover:bg-stone-900 transition-colors text-xs font-medium text-foreground dark:text-stone-100 group"
        >
          <div className="flex items-center gap-2 min-w-0">
            <BookOpenIcon className="size-4 text-primary flex-none" />
            <span className="truncate">View the docs for the flow editor</span>
          </div>
          <ExternalLinkIcon className="size-3.5 text-muted-foreground group-hover:text-foreground dark:text-stone-400 dark:group-hover:text-stone-100 transition-colors flex-none ml-2" />
        </a>
      </CardContent>

      <CardFooter className="p-4 pt-1 flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Checkbox
            id="flow-help-do-not-show-again"
            checked={doNotShowAgain}
            onCheckedChange={(checked) => setDoNotShowAgain(checked === true)}
            className="border-stone-400 dark:border-stone-600 data-[state=checked]:bg-primary data-[state=checked]:border-primary"
          />
          <label
            htmlFor="flow-help-do-not-show-again"
            className="text-xs text-muted-foreground dark:text-stone-400 select-none cursor-pointer leading-none hover:text-foreground dark:hover:text-stone-200 transition-colors"
          >
            Do not show again
          </label>
        </div>
        <Button
          variant="default"
          size="sm"
          className="text-xs h-8 px-4"
          onClick={handleOkClick}
        >
          OK
        </Button>
      </CardFooter>
    </Card>
  );
}
