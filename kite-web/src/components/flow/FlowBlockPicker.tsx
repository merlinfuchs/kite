import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { NodeCategory } from "@/lib/flow/categories";
import {
  BoxIcon,
  GitCompareIcon,
  LucideIcon,
  PlugIcon,
  TextCursorInputIcon,
} from "lucide-react";
import { useState } from "react";
import FlowNodeExplorer from "./FlowNodeExplorer";

const tabs: Record<NodeCategory, { label: string; icon: LucideIcon }> = {
  action: { label: "Action", icon: BoxIcon },
  control_flow: { label: "Control", icon: GitCompareIcon },
  option: { label: "Option", icon: TextCursorInputIcon },
  integration: { label: "Integrations", icon: PlugIcon },
};

// The blocks of the given categories in tabs, for the add block drawer and the
// insert block dialog.
export default function FlowBlockPicker({
  categories,
  onNodeSelect,
  onPick,
}: {
  categories: NodeCategory[];
  onNodeSelect?: () => void;
  onPick?: (type: string) => void;
}) {
  const [category, setCategory] = useState(categories[0]);

  return (
    <>
      <div className="px-4 py-2 border-b border-border/40 flex-none">
        <Tabs
          value={category}
          onValueChange={(val) => setCategory(val as NodeCategory)}
        >
          <TabsList className="flex w-full h-auto bg-muted/60 p-1 rounded-lg">
            {categories.map((c) => {
              const { label, icon: Icon } = tabs[c];
              return (
                <TabsTrigger
                  key={c}
                  value={c}
                  className="flex-auto flex-col gap-0.5 px-1.5 py-1.5 text-[11px]"
                >
                  <Icon className="size-4" />
                  <span>{label}</span>
                </TabsTrigger>
              );
            })}
          </TabsList>
        </Tabs>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden">
        <FlowNodeExplorer
          category={category}
          hideHeader
          onNodeSelect={onNodeSelect}
          onPick={onPick}
        />
      </div>
    </>
  );
}
