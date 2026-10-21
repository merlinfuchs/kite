import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { FlowContextType } from "@/lib/flow/context";
import {
  EditorSettingsStore,
  KeyboardMovementKeybinds,
  ReturnToCanvasBehavior,
  useEditorSettings,
} from "@/lib/hooks/editorSettings";
import { useHookedTheme } from "@/lib/hooks/theme";
import { useIsMobile } from "@/lib/hooks/use-mobile";
import { cn } from "@/lib/utils";
import {
  ChevronRightIcon,
  KeyboardIcon,
  LucideIcon,
  MoonStarIcon,
  RotateCcwIcon,
  SettingsIcon,
  SunIcon,
  XIcon,
} from "lucide-react";
import { Fragment, useEffect, useState } from "react";
import { toast } from "sonner";

export type EditorSettingsScope = "flow" | "message";

export interface EditorSettingsContext {
  settings: EditorSettingsStore;
  theme: string | undefined;
  setTheme: (theme: string) => void;
  editor: EditorSettingsScope;
  flowContext?: FlowContextType;
  isMobile: boolean;
}

export interface EditorSettingVisibility {
  editors?: EditorSettingsScope[];
  flowContexts?: FlowContextType[];
  hideOnMobile?: boolean;
  hideOnDesktop?: boolean;
  isVisible?: (ctx: EditorSettingsContext) => boolean;
}

export function isEditorSettingVisible(
  target: EditorSettingVisibility,
  ctx: EditorSettingsContext
): boolean {
  if (target.editors && !target.editors.includes(ctx.editor)) {
    return false;
  }
  if (
    target.flowContexts &&
    (!ctx.flowContext || !target.flowContexts.includes(ctx.flowContext))
  ) {
    return false;
  }
  if (target.hideOnMobile && ctx.isMobile) {
    return false;
  }
  if (target.hideOnDesktop && !ctx.isMobile) {
    return false;
  }
  if (target.isVisible && !target.isVisible(ctx)) {
    return false;
  }
  return true;
}

export type EditorSettingItem = EditorSettingVisibility &
  (
    | {
        id: string;
        type: "segmented";
        label: string;
        description?: string;
        options: { value: string; label: string; icon?: LucideIcon }[];
        getValue: (ctx: EditorSettingsContext) => string;
        setValue: (ctx: EditorSettingsContext, value: string) => void;
      }
    | {
        id: string;
        type: "toggle";
        label: string;
        description?: string;
        getValue: (ctx: EditorSettingsContext) => boolean;
        setValue: (ctx: EditorSettingsContext, value: boolean) => void;
        subPageLink?: {
          pageId: string;
          label: string;
          icon?: LucideIcon;
        };
      }
    | {
        id: string;
        type: "select";
        label: string;
        description?: string;
        options: { value: string; label: string }[];
        getValue: (ctx: EditorSettingsContext) => string;
        setValue: (ctx: EditorSettingsContext, value: string) => void;
      }
    | {
        id: string;
        type: "keybind-list";
        label?: string;
        description?: string;
        items: {
          action: keyof KeyboardMovementKeybinds;
          label: string;
          isDisabled?: (ctx: EditorSettingsContext) => boolean;
        }[];
      }
  );

export interface EditorSettingsPage extends EditorSettingVisibility {
  id: string;
  title: string;
  description: string;
  items: EditorSettingItem[];
  onReset?: (ctx: EditorSettingsContext) => void;
}

export const editorSettingsPages: Record<string, EditorSettingsPage> = {
  root: {
    id: "root",
    title: "Settings",
    description: "Configure your editor appearance and canvas preferences.",
    items: [
      {
        id: "theme",
        type: "segmented",
        label: "Theme",
        options: [
          { value: "light", label: "Light", icon: SunIcon },
          { value: "dark", label: "Dark", icon: MoonStarIcon },
        ],
        getValue: (ctx) => (ctx.theme === "dark" ? "dark" : "light"),
        setValue: (ctx, value) => ctx.setTheme(value),
      },
      {
        id: "show-share",
        type: "toggle",
        label: "Show Share button",
        description: "Display the share button in the editor top bar.",
        editors: ["flow"],
        getValue: (ctx) => ctx.settings.showShare,
        setValue: (ctx, value) => ctx.settings.setShowShare(value),
      },
      {
        id: "keyboard-movement",
        type: "toggle",
        label: "Keyboard movement",
        description: "Pan the canvas using keyboard shortcuts.",
        hideOnMobile: true,
        getValue: (ctx) => ctx.settings.keyboardMovement,
        setValue: (ctx, value) => ctx.settings.setKeyboardMovement(value),
        subPageLink: {
          pageId: "keyboard",
          label: "Keyboard movement settings",
          icon: KeyboardIcon,
        },
      },
      {
        id: "return-to-canvas",
        type: "select",
        label: "Return to canvas behavior",
        description: "Camera position when returning to a canvas.",
        options: [
          { value: "root", label: "Return to command root" },
          { value: "previous", label: "Return to previous canvas location" },
          { value: "center", label: "Return to canvas center" },
        ],
        getValue: (ctx) => ctx.settings.returnToCanvas,
        setValue: (ctx, value) =>
          ctx.settings.setReturnToCanvas(value as ReturnToCanvasBehavior),
      },
    ],
  },
  keyboard: {
    id: "keyboard",
    title: "Keyboard Movement",
    description: "Customize movement keybinds and fast panning behavior.",
    hideOnMobile: true,
    onReset: (ctx) => ctx.settings.resetKeybinds(),
    items: [
      {
        id: "direction-keybinds",
        type: "keybind-list",
        label: "Direction keybinds",
        description: "Click a keybind and press any key to reassign it.",
        items: [
          { action: "up", label: "Move up" },
          { action: "down", label: "Move down" },
          { action: "left", label: "Move left" },
          { action: "right", label: "Move right" },
        ],
      },
      {
        id: "fast-movement",
        type: "toggle",
        label: "Fast movement",
        description: "Hold the bound key while moving to pan faster.",
        getValue: (ctx) => ctx.settings.fastMovement,
        setValue: (ctx, value) => ctx.settings.setFastMovement(value),
      },
      {
        id: "fast-movement-keybind",
        type: "keybind-list",
        items: [
          {
            action: "fast",
            label: "Fast movement key",
            isDisabled: (ctx) => !ctx.settings.fastMovement,
          },
        ],
      },
    ],
  },
};

const keybindActionLabels: Record<keyof KeyboardMovementKeybinds, string> = {
  up: "Move up",
  down: "Move down",
  left: "Move left",
  right: "Move right",
  fast: "Fast movement key",
};

function formatKeyName(key: string): string {
  switch (key.toLowerCase()) {
    case " ":
      return "Space";
    case "arrowup":
      return "↑ Up";
    case "arrowdown":
      return "↓ Down";
    case "arrowleft":
      return "← Left";
    case "arrowright":
      return "→ Right";
    case "shift":
      return "Shift";
    case "control":
      return "Ctrl";
    case "alt":
      return "Alt";
    case "tab":
      return "Tab";
    default:
      return key.length === 1
        ? key.toUpperCase()
        : key.charAt(0).toUpperCase() + key.slice(1);
  }
}

type SweepTransition = {
  fromPageId: string;
  fromDepth: number;
  direction: "forward" | "backward";
};

interface EditorSettingsMenuProps {
  editor?: EditorSettingsScope;
  flowContext?: FlowContextType;
}

export default function EditorSettingsMenu({
  editor = "flow",
  flowContext,
}: EditorSettingsMenuProps = {}) {
  const { theme, setTheme } = useHookedTheme();
  const settings = useEditorSettings();
  const isMobile = useIsMobile();

  const [open, setOpen] = useState(false);
  const [pageStack, setPageStack] = useState<string[]>(["root"]);
  const [transition, setTransition] = useState<SweepTransition | null>(null);
  const [listeningAction, setListeningAction] = useState<
    keyof KeyboardMovementKeybinds | null
  >(null);

  const currentPageId = pageStack[pageStack.length - 1] ?? "root";
  const currentPage =
    editorSettingsPages[currentPageId] ?? editorSettingsPages.root;

  const ctx: EditorSettingsContext = {
    settings,
    theme,
    setTheme,
    editor,
    flowContext,
    isMobile,
  };

  useEffect(() => {
    if (
      pageStack.length > 1 &&
      !isEditorSettingVisible(currentPage, {
        settings,
        theme,
        setTheme,
        editor,
        flowContext,
        isMobile,
      })
    ) {
      setPageStack(["root"]);
      setTransition(null);
      setListeningAction(null);
    }
  }, [
    pageStack.length,
    currentPage,
    settings,
    theme,
    setTheme,
    editor,
    flowContext,
    isMobile,
  ]);

  useEffect(() => {
    if (!transition) return;
    const timer = setTimeout(() => {
      setTransition(null);
    }, 200);
    return () => clearTimeout(timer);
  }, [transition]);

  const handleOpenChange = (nextOpen: boolean) => {
    setOpen(nextOpen);
    if (!nextOpen) {
      setPageStack(["root"]);
      setTransition(null);
      setListeningAction(null);
    }
  };

  const navigateToPage = (pageId: string) => {
    const targetPage = editorSettingsPages[pageId];
    if (
      !targetPage ||
      pageId === currentPageId ||
      !isEditorSettingVisible(targetPage, ctx)
    ) {
      return;
    }
    setListeningAction(null);
    setTransition({
      fromPageId: currentPageId,
      fromDepth: pageStack.length,
      direction: "forward",
    });
    setPageStack((prev) => [...prev, pageId]);
  };

  const navigateToStackIndex = (index: number) => {
    const targetPageId = pageStack[index];
    if (!targetPageId || targetPageId === currentPageId) return;
    setListeningAction(null);
    setTransition({
      fromPageId: currentPageId,
      fromDepth: pageStack.length,
      direction: "backward",
    });
    setPageStack((prev) => prev.slice(0, index + 1));
  };

  useEffect(() => {
    if (!listeningAction) return;

    const onKeyDown = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();

      if (e.key === "Escape") {
        setListeningAction(null);
        return;
      }

      const normalized = e.key.toLowerCase();
      const conflict = (
        Object.entries(settings.keybinds) as [
          keyof KeyboardMovementKeybinds,
          string
        ][]
      ).find(
        ([action, boundKey]) =>
          action !== listeningAction && boundKey.toLowerCase() === normalized
      );

      if (conflict) {
        toast.error(
          `"${formatKeyName(normalized)}" is already bound to "${
            keybindActionLabels[conflict[0]]
          }".`
        );
        setListeningAction(null);
        return;
      }

      settings.setKeybind(listeningAction, normalized);
      setListeningAction(null);
    };

    window.addEventListener("keydown", onKeyDown, true);
    return () => window.removeEventListener("keydown", onKeyDown, true);
  }, [listeningAction, settings]);

  const renderPageContent = (
    page: EditorSettingsPage,
    depth: number,
    interactive: boolean
  ) => {
    const visibleItems = page.items.filter((item) =>
      isEditorSettingVisible(item, ctx)
    );

    return (
      <div className="h-full flex flex-col">
        <div className="flex-1 overflow-y-auto px-6 divide-y divide-border">
          {visibleItems.map((item) => {
            if (item.type === "segmented") {
              const currentValue = item.getValue(ctx);
              return (
                <div key={item.id} className="py-4 space-y-2">
                  <div className="space-y-0.5">
                    <Label className="text-sm font-medium">{item.label}</Label>
                    {item.description && (
                      <div className="text-xs text-muted-foreground">
                        {item.description}
                      </div>
                    )}
                  </div>
                  <div
                    className="grid gap-1 bg-muted p-1 rounded-md"
                    style={{
                      gridTemplateColumns: `repeat(${item.options.length}, minmax(0, 1fr))`,
                    }}
                  >
                    {item.options.map((opt) => {
                      const Icon = opt.icon;
                      const active = currentValue === opt.value;
                      return (
                        <button
                          key={opt.value}
                          type="button"
                          onClick={() =>
                            interactive && item.setValue(ctx, opt.value)
                          }
                          className={cn(
                            "flex items-center justify-center gap-1.5 py-1.5 text-xs font-medium rounded-sm transition-colors",
                            active
                              ? "bg-background text-foreground shadow-sm"
                              : "text-muted-foreground hover:text-foreground"
                          )}
                        >
                          {Icon && <Icon className="size-3.5" />}
                          <span>{opt.label}</span>
                        </button>
                      );
                    })}
                  </div>
                </div>
              );
            }

            if (item.type === "toggle") {
              const checked = item.getValue(ctx);
              const subPage = item.subPageLink
                ? editorSettingsPages[item.subPageLink.pageId]
                : undefined;
              const showSubPageLink =
                item.subPageLink &&
                subPage &&
                isEditorSettingVisible(subPage, ctx);
              const SubIcon = item.subPageLink?.icon;
              return (
                <div key={item.id} className="py-4 space-y-3">
                  <div className="flex items-center justify-between gap-4">
                    <div className="space-y-0.5">
                      <Label
                        htmlFor={interactive ? `setting-${item.id}` : undefined}
                        className="text-sm font-medium cursor-pointer"
                      >
                        {item.label}
                      </Label>
                      {item.description && (
                        <div className="text-xs text-muted-foreground">
                          {item.description}
                        </div>
                      )}
                    </div>
                    <Switch
                      id={interactive ? `setting-${item.id}` : undefined}
                      checked={checked}
                      onCheckedChange={(val) =>
                        interactive && item.setValue(ctx, val)
                      }
                    />
                  </div>
                  {showSubPageLink && item.subPageLink && (
                    <button
                      type="button"
                      onClick={() =>
                        interactive && navigateToPage(item.subPageLink!.pageId)
                      }
                      className="w-full flex items-center justify-between px-3 py-2 rounded-md border bg-muted/40 hover:bg-muted text-xs font-medium text-foreground transition-colors"
                    >
                      <span className="flex items-center gap-2">
                        {SubIcon && (
                          <SubIcon className="size-3.5 text-muted-foreground" />
                        )}
                        {item.subPageLink.label}
                      </span>
                      <ChevronRightIcon className="size-4 text-muted-foreground" />
                    </button>
                  )}
                </div>
              );
            }

            if (item.type === "select") {
              const value = item.getValue(ctx);
              return (
                <div key={item.id} className="py-4 space-y-2">
                  <div className="space-y-0.5">
                    <Label
                      htmlFor={interactive ? `setting-${item.id}` : undefined}
                      className="text-sm font-medium"
                    >
                      {item.label}
                    </Label>
                    {item.description && (
                      <div className="text-xs text-muted-foreground">
                        {item.description}
                      </div>
                    )}
                  </div>
                  <Select
                    value={value}
                    onValueChange={(val) =>
                      interactive && item.setValue(ctx, val)
                    }
                  >
                    <SelectTrigger
                      id={interactive ? `setting-${item.id}` : undefined}
                      className="h-9"
                    >
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {item.options.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              );
            }

            if (item.type === "keybind-list") {
              return (
                <div key={item.id} className="py-4 space-y-3">
                  {(item.label || item.description) && (
                    <div className="space-y-0.5">
                      {item.label && (
                        <Label className="text-sm font-medium">
                          {item.label}
                        </Label>
                      )}
                      {item.description && (
                        <div className="text-xs text-muted-foreground">
                          {item.description}
                        </div>
                      )}
                    </div>
                  )}
                  <div className="space-y-2">
                    {item.items.map((kb) => {
                      const isListening =
                        interactive && listeningAction === kb.action;
                      const disabled = kb.isDisabled?.(ctx) ?? false;
                      return (
                        <div
                          key={kb.action}
                          className="flex items-center justify-between gap-4"
                        >
                          <span className="text-sm text-foreground">
                            {kb.label}
                          </span>
                          <Button
                            type="button"
                            variant={isListening ? "default" : "outline"}
                            size="sm"
                            disabled={disabled}
                            className="h-8 min-w-24 font-mono text-xs"
                            onClick={() =>
                              interactive &&
                              setListeningAction(isListening ? null : kb.action)
                            }
                          >
                            {isListening
                              ? "Press key..."
                              : formatKeyName(settings.keybinds[kb.action])}
                          </Button>
                        </div>
                      );
                    })}
                  </div>
                </div>
              );
            }

            return null;
          })}
        </div>

        {(depth > 1 || page.onReset) && (
          <div className="flex-none flex items-center justify-between px-6 py-3 border-t border-border bg-muted/20">
            {depth > 1 ? (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => interactive && navigateToStackIndex(depth - 2)}
              >
                Back
              </Button>
            ) : (
              <div />
            )}
            {page.onReset && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="gap-1.5"
                onClick={() => {
                  if (!interactive) return;
                  setListeningAction(null);
                  page.onReset?.(ctx);
                }}
              >
                <RotateCcwIcon className="size-3.5" />
                Reset defaults
              </Button>
            )}
          </div>
        )}
      </div>
    );
  };

  const outgoingPage = transition
    ? editorSettingsPages[transition.fromPageId]
    : null;

  const triggerButton = (
    <button
      type="button"
      className="p-1.5 md:p-0 text-foreground/80 hover:text-foreground rounded-md hover:bg-muted md:hover:bg-transparent cursor-pointer"
      title="Settings"
      aria-label="Settings"
    >
      <SettingsIcon className="w-5 h-5 md:w-6 md:h-6" />
    </button>
  );

  const breadcrumbContent = (
    <Breadcrumb>
      <BreadcrumbList className="text-base font-semibold">
        {pageStack.map((pageId, index) => {
          const page = editorSettingsPages[pageId];
          if (!page) return null;
          const isLast = index === pageStack.length - 1;

          return (
            <Fragment key={pageId}>
              {index > 0 && <BreadcrumbSeparator />}
              <BreadcrumbItem>
                {isLast ? (
                  <BreadcrumbPage className="font-semibold">
                    {page.title}
                  </BreadcrumbPage>
                ) : (
                  <BreadcrumbLink asChild>
                    <button
                      type="button"
                      onClick={() => navigateToStackIndex(index)}
                      className="cursor-pointer"
                    >
                      {page.title}
                    </button>
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>
            </Fragment>
          );
        })}
      </BreadcrumbList>
    </Breadcrumb>
  );

  const settingsBody = (
    <div className="relative flex-1 min-h-0 overflow-hidden">
      {outgoingPage && transition && (
        <div
          key={`out-${transition.fromPageId}`}
          aria-hidden="true"
          className={cn(
            "absolute inset-0 pointer-events-none duration-200 ease-out fill-mode-forwards animate-out fade-out-0",
            transition.direction === "forward"
              ? "slide-out-to-left-full"
              : "slide-out-to-right-full"
          )}
        >
          {renderPageContent(outgoingPage, transition.fromDepth, false)}
        </div>
      )}
      <div
        key={`in-${currentPageId}`}
        className={cn(
          "h-full",
          transition && "duration-200 ease-out animate-in fade-in-0",
          transition?.direction === "forward" && "slide-in-from-right-full",
          transition?.direction === "backward" && "slide-in-from-left-full"
        )}
      >
        {renderPageContent(currentPage, pageStack.length, true)}
      </div>
    </div>
  );

  if (isMobile) {
    return (
      <Drawer open={open} onOpenChange={handleOpenChange}>
        <DrawerTrigger asChild>{triggerButton}</DrawerTrigger>
        <DrawerContent className="max-h-[85dvh] h-[85dvh] flex flex-col after:content-[''] after:absolute after:top-full after:left-0 after:right-0 after:h-96 after:bg-background">
          <div className="flex items-start justify-between gap-4 px-6 pt-4 pb-4 border-b border-border flex-none">
            <div className="space-y-1 text-left">
              <DrawerTitle className="text-base">
                {breadcrumbContent}
              </DrawerTitle>
              <DrawerDescription>{currentPage.description}</DrawerDescription>
            </div>
            <DrawerClose asChild>
              <Button
                variant="ghost"
                size="icon"
                className="size-8 -mr-2 -mt-1 rounded-full text-muted-foreground hover:text-foreground flex-none"
              >
                <XIcon className="size-4" />
                <span className="sr-only">Close</span>
              </Button>
            </DrawerClose>
          </div>
          {settingsBody}
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{triggerButton}</DialogTrigger>
      <DialogContent
        className="rounded-lg w-[calc(100%-2rem)] sm:max-w-[460px] h-[440px] max-h-[calc(100dvh-2rem)] flex flex-col gap-0 p-0 overflow-hidden"
        onEscapeKeyDown={(e) => {
          if (listeningAction) {
            e.preventDefault();
            setListeningAction(null);
          }
        }}
      >
        <DialogHeader className="flex-none text-left px-6 pt-6 pb-4 border-b border-border space-y-1">
          <DialogTitle className="text-base">{breadcrumbContent}</DialogTitle>
          <DialogDescription>{currentPage.description}</DialogDescription>
        </DialogHeader>
        {settingsBody}
      </DialogContent>
    </Dialog>
  );
}
