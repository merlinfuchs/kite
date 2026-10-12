import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useVariables } from "@/lib/hooks/api";
import VariableListEntry from "./VariableListEntry";
import VariableCreateDialog from "./VariableCreateDialog";
import { VariableIcon } from "lucide-react";

export default function VariableList() {
  const variables = useVariables();

  const variableCreateButton = (
    <VariableCreateDialog>
      <Button>Create variable</Button>
    </VariableCreateDialog>
  );

  return (
    <AutoAnimate className="flex flex-col md:flex-1 space-y-5">
      {!variables ? (
        <>
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </>
      ) : variables.length === 0 ? (
        <AppEmptyPlaceholder
          icon={VariableIcon}
          title="No variables yet"
          description="Variables store values like counters or settings that your commands and events can read and change."
          action={variableCreateButton}
        />
      ) : (
        <>
          {variables.map((variable, i) => (
            <VariableListEntry variable={variable!} key={i} />
          ))}
          <div className="flex">{variableCreateButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}
