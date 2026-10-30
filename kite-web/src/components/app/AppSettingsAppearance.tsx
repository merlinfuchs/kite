import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  useAppStateUserUpdateMutation,
  useAppUpdateMutation,
} from "@/lib/api/mutations";
import { useAppStateUserQuery } from "@/lib/api/queries";
import { setValidationErrors } from "@/lib/form";
import { useApp, useAppFeature, useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import {
  maxProfileImageSize,
  profileImageTypes,
  readDataUrl,
} from "@/lib/profileImage";
import { cn } from "@/lib/utils";
import { StateUserUpdateRequest } from "@/lib/types/wire.gen";
import { ExternalLinkIcon } from "lucide-react";
import Link from "next/link";
import { ChangeEvent, useCallback, useEffect, useRef } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { Input } from "../ui/input";
import { Textarea } from "../ui/textarea";

interface FormFields {
  name: string;
  description: string;
  enabled: boolean;
}

export default function AppSettingsAppearance() {
  const app = useApp();

  const form = useForm<FormFields>({
    defaultValues: {
      name: "",
      description: "",
      enabled: false,
    },
  });

  useEffect(() => {
    if (app) {
      form.reset({
        name: app.name,
        description: app.description || "",
        enabled: app.enabled,
      });
    }
  }, [app, form]);

  const updateMutation = useAppUpdateMutation(useAppId());

  const onSubmit = useCallback(
    (data: FormFields) => {
      updateMutation.mutate(
        {
          name: data.name,
          description: data.description || null,
          enabled: data.enabled,
        },
        {
          onSuccess(res) {
            if (res.success) {
              toast.success("Settings saved!");
            } else {
              if (res.error.code === "validation_failed") {
                setValidationErrors(form, res.error.data);
              } else {
                toast.error(
                  `Failed to update app: ${res.error.message} (${res.error.code})`
                );
              }
            }
          },
        }
      );
    },
    [form, updateMutation]
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Appearance</CardTitle>
        <CardDescription>
          Configure how your app appears to users in Discord and Kite.
        </CardDescription>
      </CardHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
          <CardContent className="space-y-5">
            <BotImages />
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>

          <CardFooter className="flex flex-wrap border-t px-6 py-4 gap-3">
            <Button type="submit">Save settings</Button>
            <Button
              variant="outline"
              type="button"
              className="flex gap-2"
              asChild
            >
              <Link
                href={`https://discord.com/developers/applications/${app?.discord_id}`}
                target="_blank"
              >
                <div>Manage on Discord</div>
                <ExternalLinkIcon className="w-4 h-4" />
              </Link>
            </Button>
          </CardFooter>
        </form>
      </Form>
    </Card>
  );
}

// The images are saved as soon as one is picked, as they're set on the bot
// directly instead of with the other settings.
function BotImages() {
  const appId = useAppId();
  const user = useResponseData(useAppStateUserQuery(appId));
  const animated = !!useAppFeature((f) => f.animated_profile);
  const mutation = useAppStateUserUpdateMutation(appId);

  async function update(req: StateUserUpdateRequest) {
    const res = await mutation.mutateAsync(req);
    if (res.success) {
      toast.success("Profile updated");
    } else {
      toast.error(
        `Failed to update profile: ${res.error.message} (${res.error.code})`
      );
    }
  }

  function onUpload(field: "avatar" | "banner") {
    return async (file: File) => {
      if (file.size > maxProfileImageSize) {
        toast.error("Images can be at most 5 MB");
        return;
      }
      if (file.type === "image/gif" && !animated) {
        toast.error("Your plan doesn't include animated avatars and banners.");
        return;
      }
      update({ [field]: await readDataUrl(file) });
    };
  }

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium">Avatar & Banner</div>
      <div className="flex flex-col sm:flex-row gap-5">
        <BotImage
          label="Avatar"
          url={user?.avatar_url}
          className="h-20 w-20 shrink-0 rounded-full"
          disabled={mutation.isPending}
          onUpload={onUpload("avatar")}
        />
        <BotImage
          label="Banner"
          url={user?.banner_url}
          className="h-20 min-w-0 flex-1 sm:w-56 sm:flex-none rounded-md"
          disabled={mutation.isPending}
          onUpload={onUpload("banner")}
          onRemove={() => update({ banner: "" })}
        />
      </div>
      <p className="text-sm text-muted-foreground">
        PNG, JPEG, GIF or WebP up to 5 MB.{" "}
        {animated ? (
          "Animated GIFs and WebPs move in Discord."
        ) : (
          <>
            Animated ones need a{" "}
            <Link
              href={`/apps/${appId}/premium`}
              className="text-primary hover:underline"
            >
              paid plan
            </Link>
            .
          </>
        )}
      </p>
    </div>
  );
}

function BotImage({
  label,
  url,
  className,
  disabled,
  onUpload,
  onRemove,
}: {
  label: string;
  url?: string;
  className: string;
  disabled: boolean;
  onUpload: (file: File) => void;
  onRemove?: () => void;
}) {
  const input = useRef<HTMLInputElement>(null);

  function onChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (file) onUpload(file);
  }

  return (
    <div className="flex items-center gap-3 w-full sm:w-auto">
      <div
        className={cn("bg-muted bg-cover bg-center border", className)}
        style={url ? { backgroundImage: `url(${url})` } : undefined}
        role="img"
        aria-label={label}
      />
      <div className="flex flex-col gap-2">
        <input
          ref={input}
          type="file"
          accept={profileImageTypes}
          className="hidden"
          onChange={onChange}
        />
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={disabled}
          onClick={() => input.current?.click()}
        >
          Change {label.toLowerCase()}
        </Button>
        {onRemove && url && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={disabled}
            onClick={onRemove}
          >
            Remove
          </Button>
        )}
      </div>
    </div>
  );
}
