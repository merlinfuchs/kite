import Picker from "@emoji-mart/react";
import { ReactNode } from "react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

interface Props {
  onEmojiSelect: (emoji: string) => void;
  children: ReactNode;
}

export default function EmojiPicker({ onEmojiSelect, children }: Props) {
  return (
    <Popover>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="p-0 border-none">
        <Picker
          data={async () => {
            const response = await fetch(
              "https://cdn.jsdelivr.net/npm/@emoji-mart/data/sets/15/twitter.json"
            );
            return response.json();
          }}
          onEmojiSelect={(data: any) => onEmojiSelect(data.native)}
          categories={[
            "frequent",
            "people",
            "nature",
            "foods",
            "activity",
            "places",
            "objects",
            "symbols",
            "flags",
          ]}
          theme="dark"
          set="twitter"
          getSpritesheetURL={() => {
            return "https://cdn.jsdelivr.net/npm/emoji-datasource-twitter@15.0.0/img/twitter/sheets-256/64.png";
          }}
        />
      </PopoverContent>
    </Popover>
  );
}
