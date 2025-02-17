import { forwardRef, useCallback, useMemo, useRef } from "react";
import { Input } from "../ui/input";

// See https://akashhamirwasia.com/blog/building-highlighted-input-field-in-react/

const REGEX = /({{[a-z0-9_.]+}})/g;

const PlaceholderInput = forwardRef<
  HTMLInputElement,
  {
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
  }
>(({ value, onChange, placeholder }, ref) => {
  const renderRef = useRef<HTMLDivElement>(null);

  const syncScroll = useCallback((e: any) => {
    if (renderRef.current) {
      renderRef.current.scrollTop = e.target.scrollTop;
      renderRef.current.scrollLeft = e.target.scrollLeft;
    }
  }, []);

  const parts = useMemo(() => {
    return value.split(REGEX).map((word, i) => {
      if (word.match(REGEX) !== null) {
        return (
          <span
            key={i}
            className="bg-blue-500 rounded-[3px] bg-opacity-30 pr-[8px] py-[2px] -ml-[3px]"
          >
            {word}
          </span>
        );
      } else {
        return <span key={i}>{word}</span>;
      }
    });
  }, [value]);

  return (
    <div className="relative h-10 w-full">
      <Input
        onScroll={syncScroll}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="bg-transparent absolute inset-0 z-10"
        ref={ref}
        placeholder={placeholder}
      />
      <div
        ref={renderRef}
        className="absolute inset-0 whitespace-pre overflow-x-auto select-none scroll px-3 py-2 text-sm flex items-center text-transparent"
        style={{
          scrollbarWidth: "none",
        }}
      >
        {parts}
      </div>
    </div>
  );
});

PlaceholderInput.displayName = "PlaceholderInput";
export default PlaceholderInput;
