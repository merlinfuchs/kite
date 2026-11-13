import { useState } from "react";
import { MessageComponentSelectMenu } from "@/lib/message/schema";
import { Button } from "../ui/button";
import { Checkbox } from "../ui/checkbox";
import { Input } from "../ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";

// Picks the values of a select menu the flow sent in a test, which runs the
// test again with them.
export default function FlowTestSelectDialog({
  selectMenu,
  onClose,
  onSelect,
}: {
  selectMenu: MessageComponentSelectMenu | null;
  onClose: () => void;
  onSelect: (values: string[]) => void;
}) {
  return (
    <Dialog open={!!selectMenu} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-md">
        {selectMenu && (
          <SelectForm
            // Starts empty for each select menu.
            key={selectMenu.flow_source_id}
            selectMenu={selectMenu}
            onSelect={onSelect}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

function SelectForm({
  selectMenu,
  onSelect,
}: {
  selectMenu: MessageComponentSelectMenu;
  onSelect: (values: string[]) => void;
}) {
  const [values, setValues] = useState<string[]>([]);
  const [ids, setIds] = useState("");

  const hasOptions = selectMenu.type === 3;
  const picked = hasOptions
    ? values
    : ids
        .split(",")
        .map((v) => v.trim())
        .filter(Boolean);

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {selectMenu.placeholder || "Make a selection"}
        </DialogTitle>
        <DialogDescription>
          {hasOptions
            ? "Choose what a member picks in the select menu."
            : "Enter the IDs a member picks, separated by commas."}
        </DialogDescription>
      </DialogHeader>

      {hasOptions ? (
        <div className="flex flex-col gap-2">
          {selectMenu.options.map((o) => (
            <label key={o.id} className="flex items-center gap-2 text-sm">
              <Checkbox
                checked={values.includes(o.value)}
                onCheckedChange={(checked) =>
                  setValues((v) =>
                    checked ? [...v, o.value] : v.filter((x) => x !== o.value)
                  )
                }
              />
              {o.label || o.value}
            </label>
          ))}
        </div>
      ) : (
        <Input value={ids} onChange={(e) => setIds(e.target.value)} />
      )}

      <DialogFooter>
        <Button disabled={!picked.length} onClick={() => onSelect(picked)}>
          Pick
        </Button>
      </DialogFooter>
    </>
  );
}
