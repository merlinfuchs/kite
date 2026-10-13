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
import { CircleHelpIcon } from "lucide-react";
import env from "@/lib/env/client";
import { useRouter } from "next/router";
import { checkTokenFormat } from "@/lib/discord/token";

interface FormFields {
  discord_token: string;
}

export default function AppCreateDialog({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const router = useRouter();

  const createMutation = useAppCreateMutation();
  const form = useForm<FormFields>({
    defaultValues: {
      discord_token: "",
    },
  });

  function onSubmit(data: FormFields) {
    if (createMutation.isPending) return;

    const token = data.discord_token.trim().replace(/^Bot\s+/i, "");
    const formatError = checkTokenFormat(token);
    if (formatError) {
      form.setError("discord_token", { message: formatError });
      return;
    }

    createMutation.mutate(
      { discord_token: token },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success("App created!");
            setOpen(false);
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
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
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
        <ol className="list-decimal pl-5 space-y-2 text-sm">
          <li>
            Open the{" "}
            <a
              href="https://discord.com/developers/applications"
              target="_blank"
              rel="noreferrer"
              className="text-primary hover:underline"
            >
              Discord Developer Portal
            </a>{" "}
            and click on <strong>New Application</strong>. The name you choose
            is the name of your bot.
          </li>
          <li>
            Click on <strong>Bot</strong> in the sidebar, then on{" "}
            <strong>Reset Token</strong>, and copy the token.
          </li>
        </ol>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            <FormField
              control={form.control}
              name="discord_token"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Bot Token</FormLabel>
                  <FormControl>
                    <Input type="password" autoComplete="off" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <LoadingButton type="submit" loading={createMutation.isPending}>
                Create app
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
