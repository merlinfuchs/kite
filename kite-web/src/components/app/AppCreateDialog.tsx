import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { ReactNode, useEffect, useRef, useState } from "react";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { useForm } from "react-hook-form";
import { useAppCreateMutation } from "@/lib/api/mutations";
import { toast } from "sonner";
import { setValidationErrors } from "@/lib/form";
import LoadingButton from "../common/LoadingButton";
import { CircleHelpIcon, ExternalLinkIcon } from "lucide-react";
import env from "@/lib/env/client";
import { useRouter } from "next/router";
import {
  checkTokenFormat,
  WrongTokenKind,
  wrongTokenMessages,
} from "@/lib/discord/token";
import { Button } from "../ui/button";
import DeveloperPortalMock from "./DeveloperPortalMock";
import { cn } from "@/lib/utils";
import { useTrack } from "@/lib/analytics";
import { loginUrl } from "@/lib/api/client";

interface FormFields {
  discord_token: string;
}

const steps = ["Create the bot", "Copy its token", "Add it to Kite"];

const portalUrl = "https://discord.com/developers/applications";

// Back from the Developer Portal after at least this long, the user probably
// did what the step asked.
const portalReturnMs = 3000;

interface WrongToken {
  kind: WrongTokenKind;
  // The application ID, if that's what was pasted, to link to its Bot page.
  applicationId?: string;
}

export default function AppCreateDialog({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState(0);
  const [skippedGuide, setSkippedGuide] = useState(false);
  const router = useRouter();
  const track = useTrack();

  const [wrongToken, setWrongToken] = useState<WrongToken | null>(null);

  function goToStep(next: number, returned = false) {
    const skipped = step === 0 && next === 2;
    if (skipped) setSkippedGuide(true);
    setStep(next);
    track("app_create_step", { step: next + 1, skipped, returned });
  }

  // When the user comes back from the Developer Portal, show the next step.
  const portalOpenedAt = useRef<number | null>(null);
  function onPortalOpened(from: string) {
    portalOpenedAt.current = Date.now();
    track("app_create_portal_opened", { step: step + 1, from });
  }
  useEffect(() => {
    if (!open) return;
    function onFocus() {
      const openedAt = portalOpenedAt.current;
      portalOpenedAt.current = null;
      if (openedAt && Date.now() - openedAt > portalReturnMs && step < 2) {
        goToStep(step + 1, true);
      }
    }
    window.addEventListener("focus", onFocus);
    return () => window.removeEventListener("focus", onFocus);
  });

  // Checks the token as it's pasted, so a wrong value shows right away.
  function checkWrongToken(value: string) {
    const token = cleanToken(value);
    const kind = checkTokenFormat(token);

    if (kind) {
      // Submitting clears the error, so it's set again every time.
      form.setError("discord_token", { message: wrongTokenMessages[kind] });
      if (kind !== wrongToken?.kind) {
        track("app_create_failed", { code: "token_format", kind });
      }
    } else if (wrongToken) {
      form.clearErrors("discord_token");
    }

    setWrongToken(
      kind
        ? {
            kind,
            applicationId: kind === "application_id" ? token : undefined,
          }
        : null
    );
    return kind;
  }

  const createMutation = useAppCreateMutation();
  const form = useForm<FormFields>({
    defaultValues: {
      discord_token: "",
    },
  });

  function onOpenChange(open: boolean) {
    setOpen(open);
    if (open) {
      track("app_create_opened");
    } else {
      setStep(0);
      setSkippedGuide(false);
      setWrongToken(null);
      portalOpenedAt.current = null;
      form.reset();
    }
  }

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    const token = cleanToken(data.discord_token);
    if (!token) {
      form.setError("discord_token", {
        message: "Paste your bot token first.",
      });
      track("app_create_failed", { code: "empty_token" });
      return;
    }

    if (checkWrongToken(token)) return;

    createMutation.mutate(
      { discord_token: token },
      {
        onSuccess(res) {
          if (!res.success) {
            track("app_create_failed", { code: res.error.code });
          }

          if (res.success) {
            track("app_created", { skipped_guide: skippedGuide });
            toast.success("App created!");
            onOpenChange(false);
            router.push({
              pathname: "/apps/[appId]",
              query: { appId: res.data.id },
            });
          } else if (
            res.error.code === "invalid_discord_token" ||
            res.error.code === "app_exists"
          ) {
            form.setError("discord_token", { message: res.error.message });
          } else if (res.error.code === "validation_failed") {
            setValidationErrors(form, res.error.data);
          } else if (res.error.code === "unauthorized") {
            toast.error("Your login expired. Log in again to create the app.");
            window.location.href = loginUrl;
          } else {
            toast.error(
              `Failed to create app: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            <div className="flex items-center space-x-2">
              <div>Create App</div>
              <a
                href={env.NEXT_PUBLIC_DOCS_LINK + "/guides/getting-started"}
                target="_blank"
              >
                <CircleHelpIcon className="h-5 w-5" />
              </a>
            </div>
          </DialogTitle>
          <DialogDescription>
            Kite runs your bot for you. You only need to create it in Discord
            first, which takes about a minute.
          </DialogDescription>
        </DialogHeader>

        <ol className="flex gap-2">
          {steps.map((title, i) => (
            <li key={title} className="flex-1 space-y-1.5">
              <div
                className={cn(
                  "h-1 rounded-full",
                  i <= step ? "bg-primary" : "bg-muted"
                )}
              />
              <div
                className={cn(
                  "text-xs",
                  i === step ? "font-medium" : "text-muted-foreground"
                )}
              >
                {title}
              </div>
            </li>
          ))}
        </ol>

        {step === 0 && (
          <>
            <div className="space-y-3 text-sm min-w-0">
              <p>
                Open the Discord Developer Portal and click on{" "}
                <strong>New Application</strong>. The name you choose is the
                name of your bot.
              </p>
              <DeveloperPortalMock page="applications" />
            </div>
            <DialogFooter className="gap-2 sm:flex-wrap sm:justify-between sm:space-x-0">
              <Button
                variant="ghost"
                className="px-2"
                onClick={() => goToStep(2)}
              >
                I already have a token
              </Button>
              <div className="flex flex-col-reverse sm:flex-row gap-2">
                <Button variant="outline" asChild>
                  <a
                    href={portalUrl}
                    target="_blank"
                    rel="noreferrer"
                    onClick={() => onPortalOpened("guide")}
                  >
                    Open Developer Portal
                    <ExternalLinkIcon className="h-4 w-4 ml-2" />
                  </a>
                </Button>
                <Button onClick={() => goToStep(1)}>Next</Button>
              </div>
            </DialogFooter>
          </>
        )}

        {step === 1 && (
          <>
            <div className="space-y-3 text-sm min-w-0">
              <p>
                In your new application, click on <strong>Bot</strong> in the
                sidebar. Then click on <strong>Reset Token</strong> and copy the
                token. Discord only shows it once.
              </p>
              <DeveloperPortalMock page="bot" />
            </div>
            <DialogFooter className="gap-2 sm:space-x-0">
              <Button variant="outline" onClick={() => goToStep(0)}>
                Back
              </Button>
              <Button variant="outline" asChild>
                <a
                  href={portalUrl}
                  target="_blank"
                  rel="noreferrer"
                  onClick={() => onPortalOpened("guide")}
                >
                  Open Developer Portal
                  <ExternalLinkIcon className="h-4 w-4 ml-2" />
                </a>
              </Button>
              <Button onClick={() => goToStep(2)}>Next</Button>
            </DialogFooter>
          </>
        )}

        {step === 2 && (
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
              <FormField
                control={form.control}
                name="discord_token"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Bot Token</FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        autoComplete="off"
                        autoFocus
                        {...field}
                        onChange={(e) => {
                          field.onChange(e);
                          checkWrongToken(e.target.value);
                        }}
                      />
                    </FormControl>
                    <FormDescription>
                      Under Bot in the Discord Developer Portal.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {wrongToken && (
                <div className="space-y-3 rounded-lg border bg-muted/50 p-3 text-sm min-w-0">
                  <p>
                    The bot token is under <strong>Bot</strong> in your
                    application. Click on <strong>Reset Token</strong> and copy
                    the token Discord shows.
                  </p>
                  <DeveloperPortalMock page="bot" />
                  <Button variant="outline" size="sm" asChild>
                    <a
                      href={
                        wrongToken.applicationId
                          ? `${portalUrl}/${wrongToken.applicationId}/bot`
                          : portalUrl
                      }
                      target="_blank"
                      rel="noreferrer"
                      onClick={() => onPortalOpened(wrongToken.kind)}
                    >
                      {wrongToken.applicationId
                        ? "Open the Bot page"
                        : "Open Developer Portal"}
                      <ExternalLinkIcon className="h-4 w-4 ml-2" />
                    </a>
                  </Button>
                </div>
              )}
              <DialogFooter className="gap-2 sm:space-x-0">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => goToStep(1)}
                >
                  Back
                </Button>
                <LoadingButton type="submit" loading={createMutation.isPending}>
                  Create app
                </LoadingButton>
              </DialogFooter>
            </form>
          </Form>
        )}
      </DialogContent>
    </Dialog>
  );
}

function cleanToken(value: string) {
  return value.trim().replace(/^Bot\s+/i, "");
}
