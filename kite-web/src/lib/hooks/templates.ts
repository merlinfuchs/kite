import { useMemo } from "react";
import {
  useBundleQuery,
  useBundlesByIdQueries,
  useBundlesQuery,
  useStarredBundlesQuery,
  useUserQuery,
} from "../api/queries";
import { Template, templateFromBundle } from "../flow/templates";
import { BundleGetResponse } from "../types/wire.gen";
import { useResponseData } from "./api";

// useTemplates are the official templates of the service and the bundles
// users listed in the marketplace, or undefined while they load.
export function useTemplates(): Template[] | undefined {
  const { data } = useBundlesQuery();
  return useMemo(() => {
    if (!data?.success) return undefined;
    return data.data.flatMap((bundle) => {
      const template = bundle && templateFromBundle(bundle);
      return template ? [template] : [];
    });
  }, [data]);
}

// useTemplate is the template of the bundle, null if there is none, or
// undefined while it loads. The placeholder is shown until then.
export function useTemplate(
  id: string,
  placeholder?: BundleGetResponse | null
): Template | null | undefined {
  const { data } = useBundleQuery(id, placeholder);
  return useMemo(() => {
    if (!data) return undefined;
    if (!data.success) return null;
    return templateFromBundle(data.data) ?? null;
  }, [data]);
}

// useFullTemplates are the templates with these IDs with their flows, which
// the list of useTemplates leaves out, or undefined while one loads.
export function useFullTemplates(ids: string[]): Template[] | undefined {
  const results = useBundlesByIdQueries(ids);
  // Changes when a template is added or loads again.
  const key = results.map((r, i) => `${ids[i]}@${r.dataUpdatedAt}`).join(",");
  return useMemo(() => {
    if (results.some((r) => !r.data)) return undefined;
    return results.flatMap((r) => {
      const template = r.data?.success
        ? templateFromBundle(r.data.data)
        : undefined;
      return template ? [template] : [];
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key]);
}

// useStarredTemplateIds are the IDs of the templates the user starred, or
// undefined while they load and without logging in.
export function useStarredTemplateIds(): string[] | undefined {
  const { data: user } = useUserQuery();
  return useResponseData(useStarredBundlesQuery(!!user?.success));
}
