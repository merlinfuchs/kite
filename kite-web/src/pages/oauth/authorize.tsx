import BaseLayout from "@/components/common/BaseLayout";
import LoadingButton from "@/components/common/LoadingButton";
import AccountPageLayout from "@/components/app/AccountPageLayout";
import APITokenAccessFields, {
  AccessFields,
  accessScopes,
} from "@/components/app/APITokenAccessFields";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Form, FormControl, FormField, FormItem } from "@/components/ui/form";
import { Skeleton } from "@/components/ui/skeleton";
import { useOAuthAuthorizeMutation } from "@/lib/api/mutations";
import { useOAuthAuthorizeQuery } from "@/lib/api/queries";
import { loginUrlTo } from "@/lib/api/client";
import { setValidationErrors } from "@/lib/form";
import { hasAccess } from "@/lib/apiTokens";
import { isVerifiedClientDomain } from "@/lib/oauth";
import { OAuthAuthorizeGetResponse } from "@/lib/types/wire.gen";
import { BadgeCheckIcon, TriangleAlertIcon } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/router";
import { ReactNode, useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

// The consent page of OAuth clients like Claude, which the API's metadata
// sends users to. Its URL has the client's authorization request, which the
// API checks, and approving sends the user back to the client with a code.
export default function OAuthAuthorizePage() {
  const router = useRouter();
  const query = router.isReady ? queryString(router.asPath) : null;
  const res = useOAuthAuthorizeQuery(query).data;

  useEffect(() => {
    if (res && !res.success && res.error.code === "unauthorized") {
      window.location.href = loginUrlTo(router.asPath);
    }
  }, [res, router.asPath]);

  return (
    <BaseLayout title="Connect to Kite" noindex>
      <AccountPageLayout showAccount={!!res?.success}>
        {!res || (!res.success && res.error.code === "unauthorized") ? (
          <Skeleton className="h-96" />
        ) : !res.success ? (
          <Message title="This connection link doesn't work">
            {res.error.message}. Go back to the app you came from and try
            connecting again.
          </Message>
        ) : (
          <ConsentForm request={res.data} query={query ?? ""} />
        )}
      </AccountPageLayout>
    </BaseLayout>
  );
}

function queryString(path: string) {
  const i = path.indexOf("?");
  return i === -1 ? "" : path.slice(i + 1);
}

function Message({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div>
      <h1 className="text-lg font-semibold md:text-2xl mb-1">{title}</h1>
      <p className="text-muted-foreground text-sm">{children}</p>
    </div>
  );
}

interface FormFields extends AccessFields {
  understood: boolean;
}

function ConsentForm({
  request,
  query,
}: {
  request: OAuthAuthorizeGetResponse;
  // The client's authorization request, which the API checks again.
  query: string;
}) {
  const { client, scopes } = request;
  const verified = isVerifiedClientDomain(client.domain);
  const mutation = useOAuthAuthorizeMutation(query);
  const [redirecting, setRedirecting] = useState(false);
  const [limitReached, setLimitReached] = useState<string | null>(null);

  const form = useForm<FormFields>({
    defaultValues: {
      scope: hasAccess(scopes, "kite", "write") ? "kite:write" : "kite:read",
      discordScope: "none",
      // Picking the apps is the default, so a client only gets the ones it
      // needs. All apps is one click away.
      apps: "some",
      appIds: [],
      understood: false,
    },
  });
  const understood = form.watch("understood");

  function submit(approved: boolean, data?: FormFields) {
    if (mutation.isPending) return;

    mutation.mutate(
      {
        approved,
        scopes: data ? accessScopes(data) : [],
        app_ids: !data || data.apps === "all" ? null : data.appIds,
      },
      {
        onSuccess(res) {
          if (res.success) {
            setRedirecting(true);
            window.location.href = res.data.redirect_url;
          } else if (res.error.code === "validation_failed") {
            setValidationErrors(form, res.error.data, {
              scopes: "scope",
              app_ids: "appIds",
            });
          } else if (res.error.code === "resource_limit") {
            setLimitReached(res.error.message);
          } else {
            toast.error(
              `Failed to connect: ${res.error.message} (${res.error.code})`
            );
          }
        },
      }
    );
  }

  if (redirecting) {
    return (
      <Message title={`Taking you back to ${client.name}`}>
        You can close this page if it doesn&apos;t happen on its own.
      </Message>
    );
  }

  return (
    <div className="flex flex-col gap-5">
      <div>
        <h1 className="text-lg font-semibold md:text-2xl mb-1">
          Connect {client.name} to Kite
        </h1>
        <p className="text-muted-foreground text-sm">
          {client.name} wants to use your Kite apps. Pick what it can do, you
          can disconnect it anytime under API tokens.
        </p>
        <div className="mt-3 flex items-center gap-1.5 text-sm">
          <span className="text-muted-foreground">From</span>
          <span className="font-medium">{client.domain}</span>
          {verified && (
            <BadgeCheckIcon
              className="size-4 text-primary"
              aria-label="Verified by Kite"
            />
          )}
        </div>
      </div>

      {!verified && (
        <Alert variant="destructive">
          <TriangleAlertIcon className="size-4" />
          <AlertTitle>Kite doesn&apos;t know this app</AlertTitle>
          <AlertDescription>
            Anyone can give an app any name, its domain {client.domain} is what
            says who made it. Only connect it if you just started connecting it
            yourself and trust {client.domain}.
          </AlertDescription>
        </Alert>
      )}

      {limitReached && (
        <Alert variant="destructive">
          <TriangleAlertIcon className="size-4" />
          <AlertTitle>Too many tokens and connections</AlertTitle>
          <AlertDescription>
            {limitReached}.{" "}
            <Link href="/apps?api-tokens" className="underline">
              Open API tokens
            </Link>
          </AlertDescription>
        </Alert>
      )}

      <Form {...form}>
        <form
          onSubmit={form.handleSubmit((data) => submit(true, data))}
          className="grid gap-3"
        >
          <APITokenAccessFields scopes={scopes} />
          {!verified && (
            <FormField
              control={form.control}
              name="understood"
              render={({ field }) => (
                <FormItem>
                  <label className="flex items-start gap-2 text-sm">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={(checked) =>
                          field.onChange(checked === true)
                        }
                        className="mt-0.5"
                      />
                    </FormControl>
                    I trust {client.domain} with my bots, it can do anything
                    with the access I pick here.
                  </label>
                </FormItem>
              )}
            />
          )}
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              disabled={mutation.isPending}
              onClick={() => submit(false)}
            >
              Cancel
            </Button>
            <LoadingButton
              type="submit"
              loading={mutation.isPending}
              disabled={!verified && !understood}
            >
              Connect
            </LoadingButton>
          </div>
        </form>
      </Form>
    </div>
  );
}
