import { describe, it, expect, beforeEach, vi } from "vitest";
import { useSidebarSupportStore, THREE_DAYS_MS } from "./sidebarSupport";

describe("useSidebarSupportStore", () => {
  beforeEach(() => {
    const storage: Record<string, string> = {};
    vi.stubGlobal("localStorage", {
      getItem: (key: string) => storage[key] ?? null,
      setItem: (key: string, value: string) => {
        storage[key] = value;
      },
      removeItem: (key: string) => {
        delete storage[key];
      },
    });

    useSidebarSupportStore.setState({ dismissedUntil: null });
    vi.useRealTimers();
  });

  it("is not dismissed initially", () => {
    const state = useSidebarSupportStore.getState();
    expect(state.dismissedUntil).toBeNull();
    expect(state.isDismissed()).toBe(false);
  });

  it("dismisses for 3 days when dismiss() is called", () => {
    const now = 1700000000000;
    vi.useFakeTimers();
    vi.setSystemTime(now);

    useSidebarSupportStore.getState().dismiss();

    const state = useSidebarSupportStore.getState();
    expect(state.dismissedUntil).toBe(now + THREE_DAYS_MS);
    expect(state.isDismissed()).toBe(true);

    // Fast-forward 2 days: still dismissed
    vi.advanceTimersByTime(2 * 24 * 60 * 60 * 1000);
    expect(useSidebarSupportStore.getState().isDismissed()).toBe(true);

    // Fast-forward past 3 days: no longer dismissed
    vi.advanceTimersByTime(1 * 24 * 60 * 60 * 1000 + 1);
    expect(useSidebarSupportStore.getState().isDismissed()).toBe(false);
  });
});
