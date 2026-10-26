import AdminPage from "@/components/app/AdminPage";
import LoadingButton from "@/components/common/LoadingButton";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import {
  useBundleRemoveMutation,
  useBundleReportDismissMutation,
} from "@/lib/api/mutations";
import { useOpenBundleReports } from "@/lib/hooks/api";
import { OpenBundleReport } from "@/lib/types/wire.gen";
import { useState } from "react";
import { toast } from "sonner";

// AdminReportsPage is where admins dismiss reports of bundles, or take the
// bundles down for breaking the rules.
export default function AdminReportsPage() {
  return (
    <AdminPage
      title="Reports"
      description="Reports of users' templates, oldest first. Taking a template down resolves all of its reports and is final."
      className="container max-w-3xl py-12"
    >
      <ReportList />
    </AdminPage>
  );
}

function ReportList() {
  const reports = useOpenBundleReports();
  if (reports && !reports.length) {
    return <p className="text-sm text-muted-foreground">Nothing to look at.</p>;
  }
  return (
    <div className="space-y-4">
      {reports?.map((r) => r && <Report key={r.id} report={r} />)}
    </div>
  );
}

function Report({ report }: { report: OpenBundleReport }) {
  const dismissMutation = useBundleReportDismissMutation();
  const removeMutation = useBundleRemoveMutation();
  const [reason, setReason] = useState("");

  async function dismiss() {
    const res = await dismissMutation.mutateAsync(report.id);
    if (!res.success) {
      toast.error(
        `Failed to dismiss: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success("Dismissed the report");
  }

  async function remove() {
    const res = await removeMutation.mutateAsync({
      bundleId: report.bundle_id,
      req: { reason: reason.trim() },
    });
    if (!res.success) {
      toast.error(
        `Failed to take it down: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(`Took ${report.bundle_name} down`);
  }

  return (
    <Card className="p-5 space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <a
          href={`/bundles/${report.bundle_id}`}
          target="_blank"
          className="font-semibold hover:underline"
        >
          {report.bundle_name}
        </a>
        <span className="text-sm text-muted-foreground">
          by {report.owner?.display_name} (@{report.owner?.discord_username})
        </span>
        <Badge variant="outline">{report.category}</Badge>
      </div>
      {report.bundle_description && (
        <p className="text-sm text-muted-foreground">
          {report.bundle_description}
        </p>
      )}
      <blockquote className="border-l-2 pl-3 text-sm whitespace-pre-wrap">
        {report.details}
      </blockquote>
      <div className="text-xs text-muted-foreground">
        Reported {new Date(report.created_at).toLocaleString()} by{" "}
        {report.reporter?.display_name} (@{report.reporter?.discord_username})
      </div>
      <Textarea
        value={reason}
        maxLength={1000}
        rows={2}
        placeholder="Why it's taken down, its owner sees this"
        onChange={(e) => setReason(e.target.value)}
      />
      <div className="flex gap-2">
        <LoadingButton
          variant="outline"
          loading={dismissMutation.isPending}
          onClick={dismiss}
        >
          Dismiss
        </LoadingButton>
        <LoadingButton
          variant="destructive"
          loading={removeMutation.isPending}
          disabled={!reason.trim()}
          onClick={remove}
        >
          Take down
        </LoadingButton>
      </div>
    </Card>
  );
}
