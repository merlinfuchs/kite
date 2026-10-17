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
  useAppProfileUpdateMutation,
  useAppUpdateMutation,
} from "@/lib/api/mutations";
import { useAppProfileQuery } from "@/lib/api/queries";
import { setValidationErrors } from "@/lib/form";
import { useApp } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { cn, readFileAsBase64 } from "@/lib/utils";
import {
  CameraIcon,
  ExternalLinkIcon,
  Loader2Icon,
  Trash2Icon,
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
import { Textarea } from "../ui/textarea";

interface FormFields {
  name: string;
  description: string;
  enabled: boolean;
}

type ProfileImage = "avatar" | "banner";

const imageLabels: Record<ProfileImage, string> = {
  avatar: "Profile picture",
  banner: "Banner",
};

const acceptedImageTypes = [
  "image/png",
  "image/jpeg",
  "image/gif",
  "image/webp",
];

// The API takes the image inside a JSON body, which leaves room for about 5 MB.
const maxImageSize = 5 * 1024 * 1024;

const defaultAvatarUrl = "https://cdn.discordapp.com/embed/avatars/0.png";

export default function AppSettingsAppearance() {
  const app = useApp();
  const appId = useAppId();

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

  const updateMutation = useAppUpdateMutation(appId);

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

  const profileQuery = useAppProfileQuery(appId);
  const profile = profileQuery.data?.success
    ? profileQuery.data.data
    : undefined;

  const profileMutation = useAppProfileUpdateMutation(appId);

  const avatarInputRef = useRef<HTMLInputElement>(null);
  const bannerInputRef = useRef<HTMLInputElement>(null);

  // An empty value removes the image.
  const updateImage = useCallback(
    (image: ProfileImage, value: string) => {
      profileMutation.mutate(
        { [image]: value },
        {
          onSuccess(res) {
            if (res.success) {
              toast.success(
                `${imageLabels[image]} ${value ? "updated" : "removed"}!`
              );
            } else if (res.error.code === "validation_failed") {
              const reason = res.error.data?.[image] ?? res.error.message;
              toast.error(`Failed to update ${image}: ${reason}`);
            } else {
              toast.error(
                `Failed to update ${image}: ${res.error.message} (${res.error.code})`
              );
            }
          },
          onError(err) {
            toast.error(`Failed to update ${image}: ${err.message}`);
          },
        }
      );
    },
    [profileMutation]
  );

  const onFileChange = useCallback(
    async (image: ProfileImage, e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      // Lets the same file be picked again after a failed upload.
      e.target.value = "";
      if (!file) return;

      if (!acceptedImageTypes.includes(file.type)) {
        toast.error("The image must be a PNG, JPEG, GIF or WebP file.");
        return;
      }
      if (file.size > maxImageSize) {
        toast.error("The image is too large, the limit is 5 MB.");
        return;
      }

      try {
        const data = await readFileAsBase64(file);
        updateImage(image, `data:${file.type};base64,${data}`);
      } catch {
        toast.error("Failed to read the image file.");
      }
    },
    [updateImage]
  );

  const imagesDisabled = !profile || profileMutation.isPending;
  const avatarPending =
    profileMutation.isPending &&
    profileMutation.variables?.avatar !== undefined;
  const bannerPending =
    profileMutation.isPending &&
    profileMutation.variables?.banner !== undefined;

  const previewName = form.watch("name");
  const previewDescription = form.watch("description");

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
          <CardContent className="flex flex-col gap-8 md:flex-row">
            <div className="w-full shrink-0 space-y-3 md:w-72">
              <div className="overflow-hidden rounded-xl border bg-muted/30">
                <div className="relative">
                  <button
                    type="button"
                    aria-label="Change banner"
                    disabled={imagesDisabled}
                    onClick={() => bannerInputRef.current?.click()}
                    className="group relative block h-24 w-full bg-muted bg-cover bg-center focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring disabled:cursor-not-allowed"
                    style={
                      profile?.banner_url
                        ? {
                            backgroundImage: `url(${profile.banner_url}?size=1024)`,
                          }
                        : undefined
                    }
                  >
                    <ImageOverlay pending={bannerPending}>
                      <CameraIcon className="h-4 w-4" />
                      <span>
                        {profile?.banner_url ? "Change banner" : "Add banner"}
                      </span>
                    </ImageOverlay>
                  </button>
                  {profile?.banner_url && (
                    <RemoveImageButton
                      label="Remove banner"
                      disabled={imagesDisabled}
                      onClick={() => updateImage("banner", "")}
                      className="absolute right-2 top-2"
                    />
                  )}

                  <div className="absolute -bottom-10 left-4">
                    <button
                      type="button"
                      aria-label="Change profile picture"
                      disabled={imagesDisabled}
                      onClick={() => avatarInputRef.current?.click()}
                      className="group relative block h-20 w-20 overflow-hidden rounded-full border-4 border-card bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed"
                    >
                      <img
                        src={
                          profile?.avatar_url
                            ? `${profile.avatar_url}?size=256`
                            : defaultAvatarUrl
                        }
                        alt=""
                        className="h-full w-full object-cover"
                      />
                      <ImageOverlay pending={avatarPending}>
                        <CameraIcon className="h-5 w-5" />
                      </ImageOverlay>
                    </button>
                    {profile?.avatar_url && (
                      <RemoveImageButton
                        label="Remove profile picture"
                        disabled={imagesDisabled}
                        onClick={() => updateImage("avatar", "")}
                        className="absolute -right-1 bottom-0"
                      />
                    )}
                  </div>
                </div>

                <div className="min-h-[5.5rem] px-4 pb-4 pt-12">
                  <div className="truncate font-semibold">
                    {previewName || "Your app"}
                  </div>
                  {previewDescription && (
                    <p className="mt-1 line-clamp-3 whitespace-pre-line break-words text-sm text-muted-foreground">
                      {previewDescription}
                    </p>
                  )}
                </div>
              </div>
              <p className="text-[0.8rem] text-muted-foreground">
                Click the picture or banner to change it. Images apply right
                away, up to 5 MB each.
              </p>

              <input
                ref={avatarInputRef}
                type="file"
                accept={acceptedImageTypes.join(",")}
                className="hidden"
                onChange={(e) => onFileChange("avatar", e)}
              />
              <input
                ref={bannerInputRef}
                type="file"
                accept={acceptedImageTypes.join(",")}
                className="hidden"
                onChange={(e) => onFileChange("banner", e)}
              />
            </div>

            <div className="min-w-0 flex-1 space-y-5">
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
                      <Textarea className="min-h-[7.5rem]" {...field} />
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

// ImageOverlay dims an image button on hover and focus to show it can be
// changed, and stays visible with a spinner while the image is being saved.
function ImageOverlay({
  pending,
  children,
}: {
  pending: boolean;
  children: React.ReactNode;
}) {
  return (
    <span
      className={cn(
        "absolute inset-0 flex items-center justify-center gap-2 bg-black/60 text-sm font-medium text-white transition-opacity",
        pending
          ? "opacity-100"
          : "opacity-0 group-hover:opacity-100 group-focus-visible:opacity-100 group-disabled:opacity-0"
      )}
    >
      {pending ? <Loader2Icon className="h-5 w-5 animate-spin" /> : children}
    </span>
  );
}

function RemoveImageButton({
  label,
  disabled,
  onClick,
  className,
}: {
  label: string;
  disabled: boolean;
  onClick: () => void;
  className?: string;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      disabled={disabled}
      onClick={onClick}
      className={cn(
        "flex h-7 w-7 items-center justify-center rounded-full border bg-card text-muted-foreground shadow-sm transition-colors hover:text-destructive focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
        className
      )}
    >
      <Trash2Icon className="h-3.5 w-3.5" />
    </button>
  );
}
