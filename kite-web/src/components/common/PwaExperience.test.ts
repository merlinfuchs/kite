import { describe, expect, it } from "vitest";
import {
  isAllowedPwaRoute,
  isPwaDashboardRoute,
  isFlowEditorRoute,
  isPwaMarketplaceRoute,
  isOpeningBotRoute,
  isPwaTransitionRoute,
} from "./PwaExperience";

describe("isAllowedPwaRoute", () => {
  it.each([
    "/login",
    "/login?returnTo=%2Fapps",
    "/apps",
    "/apps/abc/commands",
    "/marketplace",
    "/marketplace/categories/ai",
    "/marketplace/template-id",
    "/bundles/legacy-template-id",
  ])("allows %s", (route) => {
    expect(isAllowedPwaRoute(route)).toBe(true);
  });

  it.each(["/", "/install", "/pricing", "/tools", "/apps-extra"])(
    "blocks %s",
    (route) => {
      expect(isAllowedPwaRoute(route)).toBe(false);
    }
  );
});

describe("isOpeningBotRoute", () => {
  it.each(["/apps/abc", "/apps/abc?tab=overview", "/apps/abc/"])(
    "shows the bot loading screen for %s",
    (route) => {
      expect(isOpeningBotRoute(route)).toBe(true);
    }
  );

  it.each(["/apps", "/apps/abc/commands", "/apps-extra/abc"])(
    "does not show the bot loading screen for %s",
    (route) => {
      expect(isOpeningBotRoute(route)).toBe(false);
    }
  );
});

describe("isFlowEditorRoute", () => {
  it.each([
    "/apps/[appId]/commands/[cmdId]",
    "/apps/[appId]/events/[eventId]",
    "/apps/abc/commands/123",
    "/apps/abc/events/123?tab=flow",
  ])("matches %s", (pathname) => {
    expect(isFlowEditorRoute(pathname)).toBe(true);
  });

  it.each([
    "/apps/[appId]",
    "/apps/[appId]/commands",
    "/apps/[appId]/settings",
  ])("does not match %s", (pathname) => {
    expect(isFlowEditorRoute(pathname)).toBe(false);
  });
});

describe("isPwaTransitionRoute", () => {
  it.each(["/apps", "/apps/abc", "/apps/abc/commands/123"])(
    "matches %s",
    (route) => {
      expect(isPwaTransitionRoute(route)).toBe(true);
    }
  );

  it.each([
    "/apps/abc/settings",
    "/apps/abc/commands",
    "/apps/abc/events",
    "/apps/abc/marketplace",
    "/marketplace",
    "/login",
    "/install",
    "/pricing",
  ])("does not match %s", (route) => {
    expect(isPwaTransitionRoute(route)).toBe(false);
  });
});

describe("isPwaDashboardRoute", () => {
  it.each([
    "/apps",
    "/apps/abc",
    "/apps/abc/commands",
    "/apps/abc/events",
    "/apps/abc/schedules",
    "/apps/abc/webhooks",
  ])("matches %s", (route) => {
    expect(isPwaDashboardRoute(route)).toBe(true);
  });
});

describe("isPwaMarketplaceRoute", () => {
  it.each([
    "/marketplace",
    "/marketplace/categories/ai",
    "/apps/abc/marketplace",
    "/apps/abc/marketplace/template",
  ])("matches %s", (route) => {
    expect(isPwaMarketplaceRoute(route)).toBe(true);
  });
});
