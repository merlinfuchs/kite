import Link from "next/link";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "../ui/input";
import AppEmptyPlaceholder from "./AppEmptyPlaceholder";

export default function AppSettingsCollaborators() {
  return (
    <Card x-chunk="dashboard-04-chunk-2">
      <CardHeader>
        <CardTitle>Collaborators</CardTitle>
        <CardDescription>
          Add or remove other users who can manage this app.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <AppEmptyPlaceholder
          title="Under construction"
          description="This feature is not yet available. Please check back later."
        />
      </CardContent>
    </Card>
  );
}
