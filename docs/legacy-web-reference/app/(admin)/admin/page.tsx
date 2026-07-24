import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";

export default function AdminDashboard() {
  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-display font-bold text-slate-900 dark:text-slate-50">Branch Overview</h1>
          <p className="mt-1 text-slate-500 dark:text-slate-400">Northern Addis District - Real-time statistics.</p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline">Generate Report</Button>
          <Button variant="primary">New Underwriting Task</Button>
        </div>
      </div>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard title="GWP (YTD)" value="ETB 145M" trend="+12% vs last year" positive={true} />
        <MetricCard title="Claims Ratio" value="62%" trend="+4% vs last quarter" positive={false} />
        <MetricCard title="Pending Underwriting" value="45" trend="High priority: 12" positive={null} />
        <MetricCard title="New Policies (Today)" value="28" trend="+5 vs yesterday" positive={true} />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Underwriting Queue</CardTitle>
            <CardDescription>Policies requiring manual review</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <QueueItem id="UW-1029" type="Commercial Motor" assignee="Tadesse M." status="Pending" />
              <QueueItem id="UW-1030" type="Aviation Hull" assignee="Unassigned" status="Urgent" />
              <QueueItem id="UW-1031" type="Contractors All Risks" assignee="Aster G." status="In Progress" />
              <QueueItem id="UW-1032" type="Fire & Allied" assignee="Tadesse M." status="Pending" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader>
            <CardTitle>Recent Claims</CardTitle>
            <CardDescription>Latest filed claims across branches</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <ClaimItem id="CL-2026-105" amount="ETB 450,000" type="Motor Accident" status="Adjuster Assigned" />
              <ClaimItem id="CL-2026-106" amount="ETB 1.2M" type="Fire Damage" status="Investigation" />
              <ClaimItem id="CL-2026-107" amount="ETB 15,000" type="Medical" status="Approved" />
              <ClaimItem id="CL-2026-108" amount="ETB 30,000" type="Burglary" status="Pending Docs" />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function MetricCard({ title, value, trend, positive }: { title: string; value: string; trend: string; positive: boolean | null }) {
  const trendColor = positive === true ? "text-green-600 dark:text-green-400" : positive === false ? "text-red-600 dark:text-red-400" : "text-brand-blue-600 dark:text-brand-gold";
  return (
    <Card>
      <CardHeader className="p-4 pb-2">
        <CardTitle className="text-sm font-medium text-slate-500 dark:text-slate-400">{title}</CardTitle>
      </CardHeader>
      <CardContent className="p-4 pt-0">
        <div className="text-2xl font-bold text-slate-900 dark:text-slate-50">{value}</div>
        <p className={`text-xs font-medium mt-1 ${trendColor}`}>{trend}</p>
      </CardContent>
    </Card>
  );
}

function QueueItem({ id, type, assignee, status }: { id: string; type: string; assignee: string; status: string }) {
  const isUrgent = status === "Urgent";
  return (
    <div className="flex items-center justify-between border-b border-slate-100 pb-3 last:border-0 last:pb-0 dark:border-slate-800">
      <div>
        <p className="font-medium text-slate-900 dark:text-slate-100">{id} <span className="text-sm text-slate-500 font-normal ml-2">{type}</span></p>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">Assignee: {assignee}</p>
      </div>
      <span className={`text-xs font-semibold px-2 py-1 rounded-md ${isUrgent ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300'}`}>{status}</span>
    </div>
  );
}

function ClaimItem({ id, amount, type, status }: { id: string; amount: string; type: string; status: string }) {
  return (
    <div className="flex items-center justify-between border-b border-slate-100 pb-3 last:border-0 last:pb-0 dark:border-slate-800">
      <div>
        <p className="font-medium text-slate-900 dark:text-slate-100">{id}</p>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">{type}</p>
      </div>
      <div className="text-right">
        <p className="font-medium text-slate-900 dark:text-slate-100">{amount}</p>
        <p className="text-xs text-brand-blue-600 dark:text-brand-gold mt-1 font-medium">{status}</p>
      </div>
    </div>
  );
}
