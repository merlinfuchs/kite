import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  CheckIcon,
  ChevronUpIcon,
  ClipboardPasteIcon,
  CopyIcon,
  CopyPlusIcon,
  FocusIcon,
  GitForkIcon,
  ListTreeIcon,
  MessageSquareWarningIcon,
  MoveIcon,
  PencilIcon,
  PlusIcon,
  Redo2Icon,
  ScissorsIcon,
  Trash2Icon,
  Undo2Icon,
  UnlinkIcon,
  XIcon,
} from "lucide-react";
import { useEffect, useLayoutEffect, useRef, useState } from "react";

interface Props {
  topLevelCount: number;
  actionableCount: number;
  canDisconnect: boolean;
  hasClipboard: boolean;
  isMoveMode: boolean;
  hidden?: boolean;
  onAddBlock: () => void;
  onEditNode: () => void;
  onDeselectNode: () => void;
  onStartMove: () => void;
  onStopMove: () => void;
  onCopy: () => void;
  onCut: () => void;
  onDuplicate: () => void;
  onDuplicateWithConnections: () => void;
  onDisconnect: () => void;
  onDelete: () => void;
  onPaste: () => void;
  onOpenLogs: () => void;
  onCenterFlow?: () => void;
  onUndo?: () => void;
  onRedo?: () => void;
  onFormat?: () => void;
  canUndo?: boolean;
  canRedo?: boolean;
  logCount?: number;
}

function buildConnectedMenuPath(
  chipW: number,
  chipH: number,
  menuW: number,
  menuH: number,
  progress: number
): string {
  const s = 0.5;
  const totalH = menuH + chipH;
  const xL = s;
  const yT = s;
  const xC = chipW - s;
  const xR = Math.max(xC, menuW - s);
  const yM = menuH;
  const yB = totalH - s;

  const rC = Math.max(0, Math.min(chipH / 2 - s, chipW / 2 - s));
  const rTopTarget = 12;
  const rTl = rC + (rTopTarget - rC) * progress;

  const dxStep = Math.max(0, xR - xC);
  if (dxStep <= 0.05) {
    const rTr = rTl;
    return [
      `M ${xL + rTl} ${yT}`,
      `L ${xC - rTr} ${yT}`,
      `A ${rTr} ${rTr} 0 0 1 ${xC} ${yT + rTr}`,
      `L ${xC} ${yB - rC}`,
      `A ${rC} ${rC} 0 0 1 ${xC - rC} ${yB}`,
      `L ${xL + rC} ${yB}`,
      `A ${rC} ${rC} 0 0 1 ${xL} ${yB - rC}`,
      `L ${xL} ${yT + rTl}`,
      `A ${rTl} ${rTl} 0 0 1 ${xL + rTl} ${yT}`,
      "Z",
    ].join(" ");
  }

  const dyMenu = Math.max(0, yM - yT);
  const dyChipRight = Math.max(0, yB - rC - yM);

  const rTr = Math.max(0, Math.min(12, dyMenu / 2));
  const rBr = Math.max(0, Math.min(12, dyMenu - rTr, dxStep / 2));
  const rF = Math.max(0, Math.min(12, dxStep - rBr, dyChipRight));

  return [
    `M ${xL + rTl} ${yT}`,
    `L ${xR - rTr} ${yT}`,
    rTr > 0 ? `A ${rTr} ${rTr} 0 0 1 ${xR} ${yT + rTr}` : `L ${xR} ${yT}`,
    `L ${xR} ${yM - rBr}`,
    rBr > 0 ? `A ${rBr} ${rBr} 0 0 1 ${xR - rBr} ${yM}` : `L ${xR} ${yM}`,
    `L ${xC + rF} ${yM}`,
    rF > 0 ? `A ${rF} ${rF} 0 0 0 ${xC} ${yM + rF}` : `L ${xC} ${yM}`,
    `L ${xC} ${yB - rC}`,
    `A ${rC} ${rC} 0 0 1 ${xC - rC} ${yB}`,
    `L ${xL + rC} ${yB}`,
    `A ${rC} ${rC} 0 0 1 ${xL} ${yB - rC}`,
    `L ${xL} ${yT + rTl}`,
    `A ${rTl} ${rTl} 0 0 1 ${xL + rTl} ${yT}`,
    "Z",
  ].join(" ");
}

export default function FlowMobileBottomBar({
  topLevelCount,
  actionableCount,
  canDisconnect,
  hasClipboard,
  isMoveMode,
  hidden,
  onAddBlock,
  onEditNode,
  onDeselectNode,
  onStartMove,
  onStopMove,
  onCopy,
  onCut,
  onDuplicate,
  onDuplicateWithConnections,
  onDisconnect,
  onDelete,
  onPaste,
  onOpenLogs,
  onCenterFlow,
  onUndo,
  onRedo,
  onFormat,
  canUndo = false,
  canRedo = false,
  logCount = 0,
}: Props) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [animProgress, setAnimProgress] = useState(0);
  const animProgressRef = useRef(0);
  const chipRef = useRef<HTMLDivElement>(null);
  const menuMeasureRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState({
    chipW: 132,
    chipH: 36,
    menuW: 240,
    menuH: 244,
  });

  const showChip = topLevelCount > 0;
  const isMulti = topLevelCount > 1;
  const actionSuffix = isMulti
    ? ` (${actionableCount > 0 ? actionableCount : topLevelCount})`
    : "";
  const totalSuffix = isMulti ? ` (${topLevelCount})` : "";
  const disableActionable = actionableCount === 0;

  useEffect(() => {
    if (!showChip || isMoveMode || hidden) {
      setMenuOpen(false);
      animProgressRef.current = 0;
      setAnimProgress(0);
    }
  }, [showChip, isMoveMode, hidden]);

  useEffect(() => {
    const startVal = animProgressRef.current;
    const targetVal = menuOpen ? 1 : 0;
    if (startVal === targetVal) return;

    const duration = menuOpen ? 210 : 150;
    const startTime = performance.now();
    let rafId = 0;

    const tick = (now: number) => {
      const elapsed = now - startTime;
      const t = Math.min(1, elapsed / duration);
      // Cubic ease-out when unfolding, ease-in-out when folding back
      const eased = menuOpen
        ? 1 - Math.pow(1 - t, 3)
        : t < 0.5
        ? 4 * t * t * t
        : 1 - Math.pow(-2 * t + 2, 3) / 2;

      const nextVal = startVal + (targetVal - startVal) * eased;
      animProgressRef.current = nextVal;
      setAnimProgress(nextVal);

      if (t < 1) {
        rafId = requestAnimationFrame(tick);
      }
    };

    rafId = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(rafId);
  }, [menuOpen]);

  const isMenuRendered = menuOpen || animProgress > 0.002;

  useLayoutEffect(() => {
    if (!isMenuRendered) return;
    const chipEl = chipRef.current;
    const menuEl = menuMeasureRef.current;
    if (!chipEl || !menuEl) return;

    const updateSizes = () => {
      const cW = chipEl.offsetWidth;
      const cH = chipEl.offsetHeight;
      const mW = Math.max(cW + 32, menuEl.offsetWidth);
      const mH = menuEl.offsetHeight;
      setDimensions((prev) =>
        prev.chipW === cW &&
        prev.chipH === cH &&
        prev.menuW === mW &&
        prev.menuH === mH
          ? prev
          : { chipW: cW, chipH: cH, menuW: mW, menuH: mH }
      );
    };

    updateSizes();
    const ro = new ResizeObserver(updateSizes);
    ro.observe(chipEl);
    ro.observe(menuEl);
    return () => ro.disconnect();
  }, [isMenuRendered, topLevelCount, actionableCount]);

  useEffect(() => {
    if (!menuOpen) return;

    const handlePointerDown = (e: PointerEvent) => {
      if (chipRef.current && !chipRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setMenuOpen(false);
      }
    };

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [menuOpen]);

  const currentMenuH = dimensions.menuH * animProgress;
  // Let the menu rise above 24px before expanding rightward so the concave fillet never pinches
  const minHForWidth = Math.min(26, dimensions.menuH * 0.2);
  const widthFactor =
    currentMenuH <= minHForWidth
      ? 0
      : Math.min(
          1,
          (currentMenuH - minHForWidth) /
            Math.max(1, dimensions.menuH - minHForWidth)
        );
  const currentMenuW =
    dimensions.chipW + (dimensions.menuW - dimensions.chipW) * widthFactor;
  const totalSvgH = currentMenuH + dimensions.chipH;
  const svgPath = isMenuRendered
    ? buildConnectedMenuPath(
        dimensions.chipW,
        dimensions.chipH,
        currentMenuW,
        currentMenuH,
        animProgress
      )
    : "";

  return (
    <div
      className={cn(
        "fixed bottom-0 inset-x-0 z-30 md:hidden bg-background/95 backdrop-blur-md border-t border-border px-2.5 sm:px-3 py-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] flex items-center justify-between gap-1.5 shadow-lg select-none",
        hidden && "hidden"
      )}
    >
      <div className="flex items-center gap-1.5 min-w-0">
        <Button
          size="sm"
          className="gap-1.5 h-9 rounded-full px-3 flex-none font-medium shadow-sm"
          onClick={onAddBlock}
        >
          <PlusIcon className="size-4" />
          <span className={showChip ? "hidden sm:inline" : ""}>Add Block</span>
        </Button>

        {showChip && isMoveMode && (
          <div className="flex items-center gap-1 bg-primary text-primary-foreground rounded-full pl-2.5 pr-1 h-9 shadow-sm min-w-0">
            <button
              type="button"
              className="flex items-center gap-1.5 text-xs font-medium truncate pr-1"
              onClick={onStopMove}
              title="Finish moving blocks"
            >
              <MoveIcon className="size-3.5 flex-none" />
              <span className="truncate">Moving ({topLevelCount})</span>
            </button>
            <button
              type="button"
              className="p-1 bg-primary-foreground/20 hover:bg-primary-foreground/30 rounded-full flex-none"
              onClick={onStopMove}
              title="Done moving"
            >
              <CheckIcon className="size-3.5" />
            </button>
          </div>
        )}

        {showChip && !isMoveMode && (
          <div
            ref={chipRef}
            className={cn(
              "relative flex items-center gap-0.5 pl-2.5 pr-1 h-9 min-w-0",
              !isMenuRendered &&
                "bg-muted/80 rounded-full ring-1 ring-inset ring-primary/40"
            )}
          >
            {isMenuRendered && (
              <>
                {/* Single unified SVG drawing both the menu and the chip as one seamless outline */}
                <svg
                  width={currentMenuW}
                  height={totalSvgH}
                  viewBox={`0 0 ${currentMenuW} ${totalSvgH}`}
                  className="absolute left-0 bottom-0 z-20 pointer-events-none overflow-visible drop-shadow-[0_-6px_18px_rgba(0,0,0,0.12)]"
                  aria-hidden="true"
                >
                  <path
                    d={svgPath}
                    className="fill-popover stroke-primary/40"
                    strokeWidth="1"
                  />
                </svg>

                {/* Clipping viewport that unfolds upward from the top of the chip */}
                <div
                  style={{
                    width: currentMenuW,
                    height: currentMenuH,
                  }}
                  className="absolute bottom-full left-0 z-30 overflow-hidden"
                >
                  <div
                    ref={menuMeasureRef}
                    role="menu"
                    style={{
                      opacity: Math.max(
                        0,
                        Math.min(1, (animProgress - 0.1) / 0.75)
                      ),
                      transform: `translateY(${(1 - animProgress) * 8}px)`,
                    }}
                    className="w-max min-w-[13.5rem] p-1.5 pb-1 text-popover-foreground select-none"
                  >
                    {topLevelCount === 1 && (
                      <>
                        <button
                          type="button"
                          role="menuitem"
                          onClick={() => {
                            setMenuOpen(false);
                            onEditNode();
                          }}
                          className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-left whitespace-nowrap outline-none transition-colors hover:bg-accent hover:text-accent-foreground active:bg-accent cursor-pointer"
                        >
                          <PencilIcon className="size-4 flex-none text-muted-foreground" />
                          <span>Edit settings</span>
                        </button>
                        <div className="-mx-1 my-1 h-px bg-muted" />
                      </>
                    )}

                    {topLevelCount > 1 && (
                      <button
                        type="button"
                        role="menuitem"
                        onClick={() => {
                          setMenuOpen(false);
                          onStartMove();
                        }}
                        className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-left whitespace-nowrap outline-none transition-colors hover:bg-accent hover:text-accent-foreground active:bg-accent cursor-pointer"
                      >
                        <MoveIcon className="size-4 flex-none text-muted-foreground" />
                        <span>Move{totalSuffix}</span>
                      </button>
                    )}

                    <button
                      type="button"
                      role="menuitem"
                      disabled={disableActionable}
                      onClick={() => {
                        setMenuOpen(false);
                        onCopy();
                      }}
                      className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-left whitespace-nowrap outline-none transition-colors hover:bg-accent hover:text-accent-foreground active:bg-accent cursor-pointer disabled:pointer-events-none disabled:opacity-50"
                    >
                      <CopyIcon className="size-4 flex-none text-muted-foreground" />
                      <span>Copy{actionSuffix}</span>
                    </button>

                    <button
                      type="button"
                      role="menuitem"
                      disabled={disableActionable}
                      onClick={() => {
                        setMenuOpen(false);
                        onCut();
                      }}
                      className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-left whitespace-nowrap outline-none transition-colors hover:bg-accent hover:text-accent-foreground active:bg-accent cursor-pointer disabled:pointer-events-none disabled:opacity-50"
                    >
                      <ScissorsIcon className="size-4 flex-none text-muted-foreground" />
                      <span>Cut{actionSuffix}</span>
                    </button>

                    <button
                      type="button"
                      role="menuitem"
                      disabled={disableActionable}
                      onClick={() => {
                        setMenuOpen(false);
                        onDuplicate();
                      }}
                      className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-left whitespace-nowrap outline-none transition-colors hover:bg-accent hover:text-accent-foreground active:bg-accent cursor-pointer disabled:pointer-events-none disabled:opacity-50"
                    >
                      <CopyPlusIcon className="size-4 flex-none text-muted-foreground" />
                      <span>Duplicate{actionSuffix}</span>
                    </button>

                    <button
                      type="button"
                      role="menuitem"
                      disabled={disableActionable}
                      onClick={() => {
                        setMenuOpen(false);
                        onDuplicateWithConnections();
                      }}
                      className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-left whitespace-nowrap outline-none transition-colors hover:bg-accent hover:text-accent-foreground active:bg-accent cursor-pointer disabled:pointer-events-none disabled:opacity-50"
                    >
                      <GitForkIcon className="size-4 flex-none text-muted-foreground" />
                      <span>Duplicate with connections{actionSuffix}</span>
                    </button>

                    <button
                      type="button"
                      role="menuitem"
                      disabled={!canDisconnect}
                      onClick={() => {
                        setMenuOpen(false);
                        onDisconnect();
                      }}
                      className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-left whitespace-nowrap outline-none transition-colors hover:bg-accent hover:text-accent-foreground active:bg-accent cursor-pointer disabled:pointer-events-none disabled:opacity-50"
                    >
                      <UnlinkIcon className="size-4 flex-none text-muted-foreground" />
                      <span>Disconnect{totalSuffix}</span>
                    </button>

                    <div className="-mx-1 my-1 h-px bg-muted" />

                    <button
                      type="button"
                      role="menuitem"
                      disabled={disableActionable}
                      onClick={() => {
                        setMenuOpen(false);
                        onDelete();
                      }}
                      className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-left whitespace-nowrap text-destructive outline-none transition-colors hover:bg-accent hover:text-destructive active:bg-accent cursor-pointer disabled:pointer-events-none disabled:opacity-50"
                    >
                      <Trash2Icon className="size-4 flex-none" />
                      <span>Delete{actionSuffix}</span>
                    </button>
                  </div>
                </div>
              </>
            )}

            <button
              type="button"
              className="relative z-30 flex items-center gap-1 text-xs font-medium text-foreground truncate py-1 outline-none"
              onClick={() => setMenuOpen((prev) => !prev)}
              title="Selected block options"
              aria-expanded={menuOpen}
            >
              <span className="truncate">Selected ({topLevelCount})</span>
              <ChevronUpIcon
                className={cn(
                  "size-3.5 text-muted-foreground flex-none transition-transform duration-150",
                  menuOpen && "rotate-180"
                )}
              />
            </button>

            <button
              type="button"
              className="relative z-30 p-1 text-muted-foreground hover:text-foreground rounded-full flex-none"
              onClick={() => {
                setMenuOpen(false);
                onDeselectNode();
              }}
              title="Clear selection"
            >
              <XIcon className="size-3.5" />
            </button>
          </div>
        )}
      </div>

      <div className="flex items-center gap-0 sm:gap-0.5 bg-muted/60 p-0.5 rounded-full border border-border/50 flex-none">
        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full disabled:opacity-30 text-muted-foreground hover:text-destructive"
          onClick={onDelete}
          disabled={disableActionable}
          title="Delete selected"
        >
          <Trash2Icon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full disabled:opacity-30"
          onClick={onPaste}
          disabled={!hasClipboard}
          title="Paste"
        >
          <ClipboardPasteIcon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full disabled:opacity-30"
          onClick={onUndo}
          disabled={!canUndo}
          title="Undo"
        >
          <Undo2Icon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full disabled:opacity-30"
          onClick={onRedo}
          disabled={!canRedo}
          title="Redo"
        >
          <Redo2Icon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full"
          onClick={onCenterFlow}
          title="Center flow"
        >
          <FocusIcon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full"
          onClick={onFormat}
          title="Auto arrange layout"
        >
          <ListTreeIcon className="size-3.5 sm:size-4" />
        </Button>

        <Button
          variant="ghost"
          size="icon"
          className="size-7 sm:size-8 rounded-full relative"
          onClick={onOpenLogs}
          title="Logs"
        >
          <MessageSquareWarningIcon className="size-3.5 sm:size-4" />
          {logCount > 0 && (
            <span className="absolute -top-1 -right-1 size-3.5 bg-primary text-primary-foreground text-[9px] font-bold rounded-full flex items-center justify-center">
              {logCount > 9 ? "9+" : logCount}
            </span>
          )}
        </Button>
      </div>
    </div>
  );
}
