import { useState } from "react";
import { MessageComponentSelectMenu } from "@/lib/message/schema";
import { splitIds } from "@/lib/flow/flowTest";
import { ComponentTypeStringSelect } from "@/lib/types/message.gen";
import { FlowTestCheckboxList } from "./FlowTestField";
import { Button } from "../ui/button";
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

  const hasOptions = selectMenu.type === ComponentTypeStringSelect;
  const picked = hasOptions ? values : splitIds(ids);

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
        <FlowTestCheckboxList
          options={selectMenu.options.map((o) => ({
            label: o.label || o.value,
            value: o.value,
          }))}
          value={values}
          onChange={setValues}
        />
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
