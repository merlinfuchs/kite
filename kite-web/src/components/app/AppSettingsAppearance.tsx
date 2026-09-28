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
  useAppAvatarDeleteMutation,
  useAppAvatarUpdateMutation,
  useAppUpdateMutation,
} from "@/lib/api/mutations";
import { useAppAvatarQuery } from "@/lib/api/queries";
import { setValidationErrors } from "@/lib/form";
import { useApp, useResponseData } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import {
  ExternalLinkIcon,
  Loader2Icon,
  Trash2Icon,
  UploadIcon,
} from "lucide-react";
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
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Textarea } from "../ui/textarea";

// Must match the limits in the service's HandleAppAvatarUpdate.
const maxAvatarSize = 5 * 1024 * 1024;
const avatarTypes = ["image/png", "image/jpeg", "image/gif"];

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

  const appId = useAppId();
  const updateMutation = useAppUpdateMutation(appId);

  const avatar = useResponseData(useAppAvatarQuery(appId));
  const avatarMutation = useAppAvatarUpdateMutation(appId);
  const avatarDeleteMutation = useAppAvatarDeleteMutation(appId);
  const avatarInputRef = useRef<HTMLInputElement>(null);

  const avatarBusy = avatarMutation.isPending || avatarDeleteMutation.isPending;

  const onAvatarDelete = useCallback(() => {
    avatarDeleteMutation.mutate(undefined, {
      onSuccess(res) {
        if (res.success) {
          toast.success("Avatar removed!");
        } else {
          toast.error(
            `Failed to remove avatar: ${res.error.message} (${res.error.code})`
          );
        }
      },
    });
  }, [avatarDeleteMutation]);

  const onAvatarChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      // Reset so picking the same file again still fires onChange.
      e.target.value = "";
      if (!file) return;

      if (!avatarTypes.includes(file.type)) {
        toast.error("Avatar must be a PNG, JPEG or GIF image");
        return;
      }
      if (file.size > maxAvatarSize) {
        toast.error("Avatar must be smaller than 5 MB");
        return;
      }

      const reader = new FileReader();
      reader.onerror = () => toast.error("Failed to read image");
      reader.onload = () => {
        // Strip the "data:<type>;base64," prefix, the service sniffs the type itself.
        const base64 = (reader.result as string).split(",")[1];

        avatarMutation.mutate(
          { avatar: base64 },
          {
            onSuccess(res) {
              if (res.success) {
                toast.success("Avatar updated!");
              } else {
                toast.error(
                  `Failed to update avatar: ${res.error.message} (${res.error.code})`
                );
              }
            },
          }
        );
      };
      reader.readAsDataURL(file);
    },
    [avatarMutation]
  );

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
            <div className="flex items-center gap-4">
              <Avatar className="h-16 w-16">
                {avatar?.avatar_url && (
                  <AvatarImage src={avatar.avatar_url} alt="Bot avatar" />
                )}
                <AvatarFallback>
                  {app?.name?.slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="space-y-1">
                <div className="flex gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    className="flex gap-2"
                    disabled={avatarBusy}
                    onClick={() => avatarInputRef.current?.click()}
                  >
                    {avatarMutation.isPending ? (
                      <Loader2Icon className="w-4 h-4 animate-spin" />
                    ) : (
                      <UploadIcon className="w-4 h-4" />
                    )}
                    <div>Change avatar</div>
                  </Button>
                  {avatar?.avatar_url && (
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      title="Remove avatar"
                      aria-label="Remove avatar"
                      disabled={avatarBusy}
                      onClick={onAvatarDelete}
                    >
                      {avatarDeleteMutation.isPending ? (
                        <Loader2Icon className="w-4 h-4 animate-spin" />
                      ) : (
                        <Trash2Icon className="w-4 h-4" />
                      )}
                    </Button>
                  )}
                </div>
                <p className="text-muted-foreground text-xs">
                  PNG, JPEG or GIF, up to 5 MB. Discord only allows a few
                  changes per hour.
                </p>
              </div>
              <input
                ref={avatarInputRef}
                type="file"
                accept={avatarTypes.join(",")}
                className="hidden"
                onChange={onAvatarChange}
              />
            </div>
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
