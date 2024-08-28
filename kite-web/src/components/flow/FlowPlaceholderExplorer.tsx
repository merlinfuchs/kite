import { VariableIcon } from "lucide-react";
import PlaceholderExplorer from "../common/PlaceholderExplorer";

export default function FlowPlaceholderExplorer({
  onSelect,
}: {
  onSelect: (value: string) => void;
}) {
  return (
    <div className="absolute top-1.5 right-1.5">
      <PlaceholderExplorer onSelect={onSelect}>
        <VariableIcon
          className="h-5.5 w-5.5 text-muted-foreground hover:text-foreground cursor-pointer"
          role="button"
        />
      </PlaceholderExplorer>
    </div>
  );
}
