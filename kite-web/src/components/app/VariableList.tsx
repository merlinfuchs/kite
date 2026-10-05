import { useState } from "react";
import ListSearchInput, { ListSearchEmpty } from "./ListSearchInput";
import { matchesSearch } from "@/lib/search";
import { Button } from "../ui/button";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";
import { Skeleton } from "../ui/skeleton";
import AutoAnimate from "../common/AutoAnimate";
import { useVariables } from "@/lib/hooks/api";
import VariableListEntry from "./VariableListEntry";
import VariableCreateDialog from "./VariableCreateDialog";

export default function VariableList() {
  const variables = useVariables();
  const [search, setSearch] = useState("");

  const filtered = variables?.filter((variable) =>
    matchesSearch(search, [variable!.name])
  );

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
          <ListSearchInput
            value={search}
            onChange={setSearch}
            placeholder="Search variables"
          />
          {filtered!.length === 0 ? (
            <ListSearchEmpty query={search} />
          ) : (
            filtered!.map((variable) => (
              <VariableListEntry variable={variable!} key={variable!.id} />
            ))
          )}
          <div className="flex">{variableCreateButton}</div>
        </>
      )}
    </AutoAnimate>
  );
}
