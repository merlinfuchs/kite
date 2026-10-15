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
import { ReactNode, useState } from "react";
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
import { checkTokenFormat } from "@/lib/discord/token";
import { Button } from "../ui/button";
import DeveloperPortalMock from "./DeveloperPortalMock";
import { cn } from "@/lib/utils";
import { useTrack } from "@/lib/analytics";
import { loginUrl } from "@/lib/api/client";

interface FormFields {
  discord_token: string;
}

const steps = ["Create the bot", "Copy its token", "Add it to Kite"];

export default function AppCreateDialog({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState(0);
  const [skippedGuide, setSkippedGuide] = useState(false);
  const router = useRouter();
  const track = useTrack();

  function goToStep(next: number) {
    const skipped = step === 0 && next === 2;
    if (skipped) setSkippedGuide(true);
    setStep(next);
    track("app_create_step", { step: next + 1, skipped });
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
      form.reset();
    }
  }

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    const token = data.discord_token.trim().replace(/^Bot\s+/i, "");
    if (!token) {
      form.setError("discord_token", {
        message: "Paste your bot token first.",
      });
      track("app_create_failed", { code: "empty_token" });
      return;
    }

    const formatError = checkTokenFormat(token);
    if (formatError) {
      form.setError("discord_token", { message: formatError });
      track("app_create_failed", { code: "token_format" });
      return;
    }

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
                    href="https://discord.com/developers/applications"
                    target="_blank"
                    rel="noreferrer"
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
                      />
                    </FormControl>
                    <FormDescription>
                      Under Bot in the Discord Developer Portal.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
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
