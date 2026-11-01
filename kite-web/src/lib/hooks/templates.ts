import { useMemo, useSyncExternalStore } from "react";
import {
  useBundleQuery,
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

// useStarredTemplateIds are the IDs of the templates the user starred, or
// undefined while they load and without logging in.
export function useStarredTemplateIds(): string[] | undefined {
  const { data: user } = useUserQuery();
  return useResponseData(useStarredBundlesQuery(!!user?.success));
}

const hideAITemplatesKey = "kite-hide-ai-marketplace";
const hideAITemplatesListeners = new Set<() => void>();

function subscribeToHideAITemplates(listener: () => void) {
  hideAITemplatesListeners.add(listener);
  window.addEventListener("storage", listener);
  return () => {
    hideAITemplatesListeners.delete(listener);
    window.removeEventListener("storage", listener);
  };
}

function getHideAITemplates() {
  return localStorage.getItem(hideAITemplatesKey) === "true";
}

export function useHideAITemplates() {
  return useSyncExternalStore(
    subscribeToHideAITemplates,
    getHideAITemplates,
    () => false
  );
}

export function setHideAITemplates(hidden: boolean) {
  localStorage.setItem(hideAITemplatesKey, String(hidden));
  hideAITemplatesListeners.forEach((listener) => listener());
}
