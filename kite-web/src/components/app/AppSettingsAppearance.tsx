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
import {
  useApp,
  useAppFeature,
  useFeaturePlan,
  useResponseData,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { profileImageTypes, readProfileImage } from "@/lib/profileImage";
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
import BotProfilePreview from "./BotProfilePreview";

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

  const user = useResponseData(useAppStateUserQuery(useAppId()));
  const name = form.watch("name");
  const description = form.watch("description");

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
          <CardContent className="grid gap-6 md:grid-cols-[1fr_auto]">
            <BotProfilePreview
              className="md:order-last md:w-80 md:self-start"
              name={name || app?.name || ""}
              username={user?.username}
              avatarUrl={user?.avatar_url}
              bannerUrl={user?.banner_url}
              about={description}
            />
            <div className="space-y-5 min-w-0">
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
            </div>
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
  const plan = useFeaturePlan("animated_profile");
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
      const image = await readProfileImage(file);
      if (image) update({ [field]: image });
    };
  }

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium">Avatar & Banner</div>
      <div className="flex flex-wrap gap-2">
        <ImageButton
          label="Change avatar"
          disabled={mutation.isPending}
          onUpload={onUpload("avatar")}
        />
        <ImageButton
          label="Change banner"
          disabled={mutation.isPending}
          onUpload={onUpload("banner")}
        />
        {user?.banner_url && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={mutation.isPending}
            onClick={() => update({ banner: "" })}
          >
            Remove banner
          </Button>
        )}
      </div>
      <p className="text-sm text-muted-foreground">
        PNG, JPEG, GIF or WebP up to 5 MB.{" "}
        {animated ? (
          "Animated GIFs and WebPs move in Discord."
        ) : (
          <>
            Animated ones need{" "}
            <Link
              href={`/apps/${appId}/premium`}
              className="text-primary hover:underline"
            >
              {plan ? `Kite ${plan.title}` : "a paid plan"}
            </Link>
            .
          </>
        )}
      </p>
    </div>
  );
}

function ImageButton({
  label,
  disabled,
  onUpload,
}: {
  label: string;
  disabled: boolean;
  onUpload: (file: File) => void;
}) {
  const input = useRef<HTMLInputElement>(null);

  function onChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (file) onUpload(file);
  }

  return (
    <>
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
        {label}
      </Button>
    </>
  );
}
