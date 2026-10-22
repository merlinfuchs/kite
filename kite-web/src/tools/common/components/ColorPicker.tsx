import { useEffect, useMemo, useState } from "react";
import { HexColorPicker } from "react-colorful";
import { Popover, PopoverTrigger } from "@/components/ui/popover";
import { Input } from "@/components/ui/input";
import { PopoverContent } from "@radix-ui/react-popover";

interface Props {
  value: number | undefined;
  onChange: (newValue: number | undefined) => void;
}

// Anything but a complete color is no color, so the swatch, the preview and
// the message all match what's typed.
function parseHexColor(text: string) {
  return text.length === 6 ? parseInt(text, 16) : undefined;
}

export default function ColorPicker({ value, onChange }: Props) {
  const hexColor = useMemo(() => {
    if (value || value === 0) {
      return value.toString(16).padStart(6, "0");
    } else {
      return "";
    }
  }, [value]);

  // What's typed is kept as text, so a half-typed or half-deleted color isn't
  // turned into a number and padded with zeros on every keystroke.
  const [text, setText] = useState(hexColor);

  // Follow changes from elsewhere, like undo.
  useEffect(() => {
    setText((text) => (parseHexColor(text) === value ? text : hexColor));
  }, [value, hexColor]);

  function setHexColor(newColor: string) {
    const raw = newColor.replace(/[^0-9a-f]/gi, "").substring(0, 6);
    setText(raw);
    onChange(parseHexColor(raw));
  }

  const incomplete = text.length > 0 && text.length < 6;

  const displayColor = hexColor ? "#" + hexColor : "#1f2225";

  return (
    <div>
      <div className="flex space-x-2">
        <div className="flex flex-auto">
          <div className="border rounded-l flex items-center px-2 text-gray-300 select-none">
            #
          </div>
          <Input
            type="text"
            value={text}
            onChange={(e) => setHexColor(e.target.value)}
            placeholder="rrggbb"
            className="rounded-l-none border-l-0"
          />
        </div>
        <Popover>
          <PopoverTrigger>
            <div
              className="w-12 h-full rounded cursor-pointer relative border"
              style={{ backgroundColor: displayColor }}
              role="button"
            />
          </PopoverTrigger>
          <PopoverContent className="z-30">
            <HexColorPicker color={"#" + hexColor} onChange={setHexColor} />
          </PopoverContent>
        </Popover>
      </div>
      {incomplete && (
        <div className="text-sm text-muted-foreground pt-1">
          Enter all 6 digits, like ff5c00.
        </div>
      )}
    </div>
  );
}
