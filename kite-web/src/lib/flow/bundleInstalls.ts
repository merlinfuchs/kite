import { BundleInstall } from "../types/wire.gen";

// bundleInstallHref is the page of the template or bundle an app added, where
// it can be updated or removed. Official templates have theirs in the
// marketplace.
export function bundleInstallHref(install: BundleInstall, appId: string) {
  return install.official
    ? {
        pathname: "/apps/[appId]/marketplace/[templateId]",
        query: { appId, templateId: install.bundle_id },
      }
    : {
        pathname: "/apps/[appId]/bundles/[bundleId]",
        query: { appId, bundleId: install.bundle_id },
      };
}

// hasUpdate is whether the app can update what it added to a newer version.
export function hasUpdate(install: BundleInstall) {
  return (
    install.available_version !== null &&
    install.available_version > install.bundle_version
  );
}
