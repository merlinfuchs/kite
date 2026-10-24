import { useMemo } from "react";
import { useBundleQuery, useBundlesQuery } from "../api/queries";
import { Template, templateFromBundle } from "../flow/templates";

// useTemplates are the official templates of the service, or undefined while
// they load.
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
// undefined while it loads.
export function useTemplate(id: string): Template | null | undefined {
  const { data } = useBundleQuery(id);
  return useMemo(() => {
    if (!data) return undefined;
    if (!data.success) return null;
    return templateFromBundle(data.data) ?? null;
  }, [data]);
}
