import { Button } from "@/components/ui/button";
import CustomSelect from "@/components/common/CustomSelect";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useTableCreateMutation } from "@/lib/api/mutations";
import { useAppId } from "@/lib/hooks/params";
import { TableColumn } from "@/lib/types/wire.gen";
import { ReactNode, useState } from "react";
import { toast } from "sonner";

const types = ["text", "number", "boolean", "date", "json"] as const;

export default function TableCreateDialog({
  children,
}: {
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [columns, setColumns] = useState<TableColumn[]>([
    { id: "name", name: "Name", type: "text" },
  ]);
  const mutation = useTableCreateMutation(useAppId());

  function submit() {
    mutation.mutate(
      { name, columns },
      {
        onSuccess: (res) => {
          if (res.success) {
            toast.success("Table created!");
            setOpen(false);
            setName("");
          } else toast.error(`Failed to create table: ${res.error.message}`);
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create table</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <Input
            placeholder="Table name"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <div className="space-y-2">
            {columns.map((column, index) => (
              <div className="flex gap-2" key={column.id}>
                <Input
                  value={column.name}
                  placeholder="Column name"
                  onChange={(e) =>
                    setColumns((current) =>
                      current.map((item, i) =>
                        i === index
                          ? {
                              ...item,
                              name: e.target.value,
                              id: e.target.value
                                .toLowerCase()
                                .replace(/[^a-z0-9_]/g, "_"),
                            }
                          : item
                      )
                    )
                  }
                />
                <CustomSelect
                  value={column.type}
                  options={types.map((type) => ({ value: type, label: type }))}
                  placeholder="Type"
                  className="h-10 w-32"
                  onChange={(value) =>
                    setColumns((current) =>
                      current.map((item, i) =>
                        i === index
                          ? {
                              ...item,
                              type: value,
                            }
                          : item
                      )
                    )
                  }
                />
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() =>
                    setColumns((current) =>
                      current.filter((_, i) => i !== index)
                    )
                  }
                >
                  Remove
                </Button>
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                setColumns((current) => [
                  ...current,
                  {
                    id: `column_${current.length + 1}`,
                    name: `Column ${current.length + 1}`,
                    type: "text",
                  },
                ])
              }
            >
              Add column
            </Button>
          </div>
        </div>
        <DialogFooter>
          <Button
            onClick={submit}
            disabled={mutation.isPending || !name || columns.length === 0}
          >
            Create table
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
