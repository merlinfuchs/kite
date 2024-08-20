import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useVariables } from "@/lib/hooks/api";
import VariableListEntry from "./VariableListEntry";
import VariableCreateDialog from "./VariableCreateDialog";

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
          title="There are no variables"
          description="You can start now by creating the first variable!"
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
